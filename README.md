# losh

A placeholder multi-module Go project.

## Directory Structure

- `server/` - Go module containing the server application.
- `client/` - Go module containing the client application.
- `doc/` - Documentation directory.

## Prerequisites

- [Go](https://go.dev/) (1.16 or later recommended)

## How to Run & Build

### Server Module

To run the server application:
```bash
cd server && go run main.go
```

To build the server application:
```bash
cd server && go build -o server main.go
```

---

### Client Module

To run the client application:
```bash
cd client && go run main.go
```

To build the client application:
```bash
cd client && go build -o client main.go
```
