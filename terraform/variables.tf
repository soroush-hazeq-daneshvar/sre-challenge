variable "project_name" {
  description = "Project name used for resource naming"
  type        = string
  default     = "sre-challenge"
}

variable "environment" {
  description = "Environment name"
  type        = string
  default     = "production"
}

variable "region" {
  description = "Cloud provider region"
  type        = string
  default     = "us-east-1"
}

variable "control_plane_instance_type" {
  description = "Instance type for Kubernetes control plane"
  type        = string
  default     = "t3.medium"
}

variable "worker_instance_type" {
  description = "Instance type for Kubernetes worker nodes"
  type        = string
  default     = "t3.large"
}

variable "ssh_public_key_path" {
  description = "Path to SSH public key for node access"
  type        = string
  default     = "~/.ssh/id_rsa.pub"
}

variable "allowed_cidr_blocks" {
  description = "CIDR blocks allowed to access the cluster"
  type        = list(string)
  default     = ["0.0.0.0/0"]
}

variable "node_count" {
  description = "Number of worker nodes"
  type        = number
  default     = 2
}

variable "tags" {
  description = "Common tags for all resources"
  type        = map(string)
  default = {
    Project     = "sre-challenge"
    ManagedBy   = "terraform"
    Environment = "production"
  }
}
