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

### 4. Provisioning (Ansible)
After the VM is running, you can use Ansible to automatically update Ubuntu and install/enable Nginx.

#### Prerequisites
Ensure Ansible is installed on your local machine:
- **macOS**: `brew install ansible`
- **Linux/Pip**: `pip install ansible`

#### Step 1: Configure the Inventory
1. Open the [ops/ansible/inventory.ini](file:///Users/ant/git/losh/ops/ansible/inventory.ini) file.
2. Replace `YOUR_VM_PUBLIC_IP` with your actual VM public IP address.
3. Verify that `ansible_ssh_private_key_file` correctly points to your private SSH key (e.g., `~/.ssh/id_rsa`).

#### Step 2: Test the Connection
Verify that Ansible can successfully connect to the remote VM:
```bash
cd ops/ansible
ansible all -m ping -i inventory.ini
```
*(If prompted to trust the SSH key fingerprint, type `yes`.)*

#### Step 3: Run the Playbook
Execute the playbook to update Ubuntu packages and install Nginx:
```bash
ansible-playbook -i inventory.ini playbook.yml
```

#### Step 4: Verify Nginx
Open your web browser and navigate to:
```
http://<INSTANCE_PUBLIC_IP>
```
You should see the default **"Welcome to nginx!"** landing page.

### 5. Clean Up (Teardown)
To completely remove all the provisioned infrastructure and avoid any potential cloud costs:
```bash
cd ops/server
terraform destroy -var="project_id=YOUR_GCP_PROJECT_ID"
```
