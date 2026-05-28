# Getting Started

Welcome to the `losh` project! This guide will help you set up your local development environment and run the services.

## Project Structure

- **`server/`**: A standalone Go module containing the server application.
- **`client/`**: A standalone Go module containing the client application.
- **`doc/`**: Documentation for the project.

## Prerequisites

Ensure you have the following installed on your machine:
- **Go** (version 1.16 or later is recommended). You can download it from [go.dev](https://go.dev/).

---

## Local Setup

### 1. Clone the repository
```bash
git clone <repository-url>
cd losh
```

### 2. Running the Server

Navigate to the `server` directory and run the application:
```bash
cd server
go run main.go
```

To build a production binary:
```bash
go build -o server main.go
```

### 3. Running the Client

Navigate to the `client` directory and run the application:
```bash
cd client
go run main.go
```

To build a production binary:
```bash
go build -o client main.go
```

---

## Next Steps

- Explore the `server/main.go` and `client/main.go` files to understand their core entry points.
- Customize the server port or client endpoints as you begin your implementation.
