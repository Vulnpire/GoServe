# GoServe

This Go application provides a versatile server with two primary functionalities:
1. **TCP Server** - Listens on a specified interface and port, similar to Netcat.
2. **HTTP Server** - Serves files over HTTP with optional Basic Authentication and HTTPS support.

## Features
- **TCP Server**: Listen for incoming TCP connections and output data to stdout.
- **HTTP Server**: Serve files from the current directory with:
  - Basic Authentication
  - HTTPS support
- **Configurable Logging**: Set log level and log output file.

## Getting Started

To use this server, you'll need to have Go installed. You can get Go from [golang.org](https://golang.org/dl/).

1. Clone the Repository:

`git clone https://github.com/Vulnpire/GoServe`


2. Build the Application:

go build -o server main.go

## Usage

Start HTTP Server

To start the HTTP server with Basic Authentication and optional HTTPS:

`./server -serve <port> [-tls-cert <path>] [-tls-key <path>] [-auth-user <username>] [-auth-pass <password>]`

```
<interface>: Interface to listen on (default: 0.0.0.0).
<port>: Port to serve files over HTTP or TCP.
-tls-cert <path>: Path to TLS certificate file (optional).
-tls-key <path>: Path to TLS key file (optional).
-auth-user <username>: Username for Basic Authentication (optional).
-auth-pass <password>: Password for Basic Authentication (optional).
```

## Example:

`./server -i 0.0.0.0 -p 8080`

`./server -serve 8080 -tls-cert cert.pem -tls-key key.pem -auth-user admin -auth-pass admin`

## CLI Flags

```
-i: Interface to listen on (for TCP server).
-p: Port to listen on (for TCP server).
-serve: Port to serve files over HTTP.
-tls-cert: Path to TLS certificate file (for HTTPS).
-tls-key: Path to TLS key file (for HTTPS).
-auth-user: Username for HTTP Basic Authentication.
-auth-pass: Password for HTTP Basic Authentication.
-log-file: Path to log file (default: logs to stdout).
-log-level: Log level (debug, info, warn, error).
```
