terraform {
  required_version = ">= 1.0.0"
  required_providers {
    google = {
      source  = "hashicorp/google"
      version = "~> 4.0"
    }
  }
}

provider "google" {
  project = var.project_id
  region  = var.region
  zone    = var.zone
}

variable "project_id" {
  type        = string
  description = "The GCP Project ID where resources will be created."
}

variable "region" {
  type        = string
  default     = "us-central1"
  description = "GCP Region (us-central1, us-east1, or us-west1 are eligible for the Free Tier)."
}

variable "zone" {
  type        = string
  default     = "us-central1-a"
  description = "GCP Zone within the selected region."
}

variable "ssh_user" {
  type        = string
  default     = "ubuntu"
  description = "SSH username to configure on the VM."
}

variable "ssh_pub_key_path" {
  type        = string
  default     = "~/.ssh/id_rsa.pub"
  description = "Path to the public SSH key file."
}

# VPC Network
resource "google_compute_network" "vpc_network" {
  name                    = "losh-vpc"
  auto_create_subnetworks = true
}

# Firewall rule to allow SSH access
resource "google_compute_firewall" "allow_ssh" {
  name    = "losh-allow-ssh"
  network = google_compute_network.vpc_network.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["ssh-enabled"]
}

# Firewall rule to allow reverse proxy ports (losh)
resource "google_compute_firewall" "allow_losh" {
  name    = "losh-allow-ports"
  network = google_compute_network.vpc_network.name

  allow {
    protocol = "tcp"
    # Port 7000 (control), 80/443 (HTTP/S), and 8080 (dashboard or alternative port)
    ports    = ["7000", "80", "443", "8080"]
  }

  source_ranges = ["0.0.0.0/0"]
  target_tags   = ["losh-server"]
}

# Free Tier VM Instance
resource "google_compute_instance" "free_vm" {
  name         = "losh-server"
  machine_type = "e2-micro" # Eligible for GCP Free Tier
  zone         = var.zone

  tags = ["ssh-enabled", "losh-server"]

  boot_disk {
    initialize_params {
      image = "ubuntu-os-cloud/ubuntu-2204-lts"
      size  = 30 # 30 GB standard persistent disk is Free Tier eligible
      type  = "pd-standard"
    }
  }

  network_interface {
    network = google_compute_network.vpc_network.name

    access_config {
      # Ephemeral public IP
    }
  }

  metadata = {
    ssh-keys = "${var.ssh_user}:${file(var.ssh_pub_key_path)}"
  }

  lifecycle {
    ignore_changes = [metadata["ssh-keys"]]
  }
}

output "instance_public_ip" {
  value       = google_compute_instance.free_vm.network_interface[0].access_config[0].nat_ip
  description = "The public IP address of the newly created GCP instance."
}
