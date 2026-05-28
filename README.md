# losh

A placeholder multi-module Go project.

## Directory Structure

- `server/` - Go module containing the server application.
- `client/` - Go module containing the client application.
- `ops/` - Infrastructure and deployment files (Terraform).
- `doc/` - Documentation directory.

## Prerequisites

- [Go](https://go.dev/) (1.16 or later recommended)
- [Terraform](https://developer.hashicorp.com/terraform/downloads) (>= 1.0.0, optional for infrastructure deployment)
- [Google Cloud SDK](https://cloud.google.com/sdk/docs/install) (optional, for GCP authentication)

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

---

## Deployment (Server)

You can automatically provision a Google Cloud Free Tier VM (`e2-micro`) with pre-configured SSH access and firewall settings using Terraform.

### 1. Authenticate with GCP
Make sure your GCP credentials are configured locally:
```bash
gcloud auth application-default login
```

### 2. Initialize and Deploy Terraform
Navigate to the `ops/server` directory, initialize, and deploy by passing your GCP project ID:
```bash
cd ops/server
terraform init
terraform apply -var="project_id=YOUR_GCP_PROJECT_ID"
```

### 3. Connect via SSH
After the deployment completes, the public IP address will be printed as an output. Connect using your SSH key:
```bash
ssh -i /path/to/your/private_key ubuntu@<INSTANCE_PUBLIC_IP>
```
