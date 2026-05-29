# losh

A placeholder multi-module Go project.

## Directory Structure

- `server/` - Go module containing the server application.
- `client/` - Go module containing the client application.
- `ops/` - Infrastructure and deployment files:
  - `server/terraform/` - Terraform configuration to provision the GCP VM.
  - `server/ansible/` - Ansible playbook to configure the VM and install Nginx.
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
Navigate to the `ops/server/terraform` directory, initialize, and deploy by passing your GCP project ID:
```bash
cd ops/server/terraform
terraform init
terraform apply -var="project_id=YOUR_GCP_PROJECT_ID"
```

### 3. Connect via SSH
After the deployment completes, the public IP address will be printed as an output. Connect using your SSH key:
```bash
ssh -i /path/to/your/private_key ubuntu@<INSTANCE_PUBLIC_IP>
```

### 4. Provisioning (Ansible)

> [!NOTE]
> **Terraform is now configured to automatically generate the inventory file, wait for SSH readiness, and run the Ansible playbook for you during `terraform apply`!**

If you make subsequent changes to the playbook or custom HTML page and want to re-run the configuration manually, follow these steps:

#### Prerequisites
Ensure Ansible is installed on your local machine:
- **macOS**: `brew install ansible`
- **Linux/Pip**: `pip install ansible`

#### Step 1: Test the Connection
Verify that Ansible can successfully connect to the remote VM:
```bash
cd ops/server/ansible
ansible all -m ping -i inventory.ini
```
*(If prompted to trust the SSH key fingerprint, type `yes`.)*

#### Step 2: Run the Playbook
Execute the playbook manually:
```bash
ansible-playbook -i inventory.ini playbook.yml
```

#### Step 3: Verify Nginx
Open your web browser and navigate to:
```
http://<INSTANCE_PUBLIC_IP>
```
You should see the custom **losh - Local Share Server** landing page.

### 5. Releasing the Server

We have automated the build and release process using **GitHub Actions**. To trigger a new production release and automatically build and attach Ubuntu-compatible binaries (`linux-amd64` and `linux-arm64`):

1. **Tag your commit**:
   Create a version tag (it must start with `v`, e.g., `v1.0.0`):
   ```bash
   git tag v1.0.0
   ```

2. **Push the tag to GitHub**:
   ```bash
   git push origin v1.0.0
   ```

3. **Download your binaries**:
   Go to your GitHub repository's **Releases** tab. The `Releaser Losh Server` workflow will have automatically created a release page containing the optimized, pre-compiled binaries ready for deployment.

### 6. Clean Up (Teardown)
To completely remove all the provisioned infrastructure and avoid any potential cloud costs:
```bash
cd ops/server/terraform
terraform destroy -var="project_id=YOUR_GCP_PROJECT_ID"
```
