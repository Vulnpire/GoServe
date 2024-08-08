package main

import (
        "context"
        "flag"
        "fmt"
        "io"
        "log"
        "net"
        "net/http"
        "os"
        "os/signal"
        "strings"
        "syscall"
        "time"
)

var (
        iface      = flag.String("i", "0.0.0.0", "Interface to listen on (optional, default: 0.0.0.0)")
        port       = flag.String("p", "8080", "Port to listen on for TCP connections (optional, default: 8080)")
        serve      = flag.String("serve", "", "Port to serve files over HTTP (optional)")
        tlsCert    = flag.String("tls-cert", "", "Path to TLS certificate file (optional)")
        tlsKey     = flag.String("tls-key", "", "Path to TLS key file (optional)")
        authUser   = flag.String("auth-user", "", "Username for basic authentication (optional)")
        authPass   = flag.String("auth-pass", "", "Password for basic authentication (optional)")
        logFile    = flag.String("log-file", "", "Path to log file (optional)")
        logLevel   = flag.String("log-level", "info", "Log level: debug, info, warn, error (optional, default: info)")
)

func init() {
        flag.Usage = func() {
                fmt.Fprintf(os.Stderr, "Usage: %s [options]\n", os.Args[0])
                flag.PrintDefaults()
        }
}

func setupLogger() {
        log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile)
        if *logFile != "" {
                file, err := os.OpenFile(*logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
                if err != nil {
                        log.Fatalf("Failed to open log file: %v", err)
                }
                log.SetOutput(file)
        }
}

func logHTTPDownload(r *http.Request, status int, size int64) {
        log.Printf("HTTP %d %s %s %d bytes", status, r.Method, r.URL.Path, size)
}

type responseWriter struct {
        http.ResponseWriter
        status int
        size   int64
}

func (w *responseWriter) WriteHeader(status int) {
        w.status = status
        w.ResponseWriter.WriteHeader(status)
}

func (w *responseWriter) Write(b []byte) (int, error) {
        size, err := w.ResponseWriter.Write(b)
        w.size += int64(size)
        return size, err
}

func authMiddleware(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                if *authUser != "" && *authPass != "" {
                        user, pass, ok := r.BasicAuth()
                        if !ok || user != *authUser || pass != *authPass {
                                w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
                                http.Error(w, "Unauthorized", http.StatusUnauthorized)
                                return
                        }
                }
                next.ServeHTTP(w, r)
        })
}

func startHTTPServer(port string) {
        http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
                start := time.Now()
                wrappedWriter := &responseWriter{ResponseWriter: w, status: http.StatusOK}
                http.FileServer(http.Dir(".")).ServeHTTP(wrappedWriter, r)
                logHTTPDownload(r, wrappedWriter.status, wrappedWriter.size)
                log.Printf("Served %s in %v", r.URL.Path, time.Since(start))
        })

        address := fmt.Sprintf(":%s", port)
        log.Printf("Starting HTTP server on %s\n", address)

        server := &http.Server{
                Addr:    address,
                Handler: authMiddleware(http.DefaultServeMux),
        }

        go func() {
                if *tlsCert != "" && *tlsKey != "" {
                        log.Fatal(server.ListenAndServeTLS(*tlsCert, *tlsKey))
                } else {
                        log.Fatal(server.ListenAndServe())
                }
        }()

        stop := make(chan os.Signal, 1)
        signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

        <-stop
        log.Println("Shutting down HTTP server...")

        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        server.Shutdown(ctx)
        log.Println("HTTP server shut down gracefully")
}

func handleConnection(conn net.Conn) {
        defer conn.Close()
        log.Printf("TCP connection established from %s\n", conn.RemoteAddr().String())

        _, err := io.Copy(os.Stdout, conn)
        if err != nil {
                log.Printf("Error while reading from TCP connection: %v\n", err)
        }

        log.Printf("TCP connection closed from %s\n", conn.RemoteAddr().String())
}

func startTCPListener(address string) {
        listener, err := net.Listen("tcp", address)
        if err != nil {
                log.Fatalf("Error starting TCP listener: %v\n", err)
        }
        defer listener.Close()

        log.Printf("Listening for TCP connections on %s\n", address)

        for {
                conn, err := listener.Accept()
                if err != nil {
                        log.Printf("Error accepting TCP connection: %v\n", err)
                        continue
                }

                go handleConnection(conn)
        }
}

func main() {
        flag.Parse()
        setupLogger()

        if strings.ToLower(*logLevel) == "debug" {
                log.SetFlags(log.Ldate | log.Ltime | log.Lshortfile | log.Lmicroseconds)
        }

        if *serve != "" {
                startHTTPServer(*serve)
        } else {
                address := fmt.Sprintf("%s:%s", *iface, *port)
                startTCPListener(address)
        }
}
