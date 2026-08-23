output "control_plane_public_ip" {
  description = "Public IP of the Kubernetes control plane"
  value       = aws_instance.control_plane.public_ip
}

output "control_plane_private_ip" {
  description = "Private IP of the Kubernetes control plane"
  value       = aws_instance.control_plane.private_ip
}

output "worker_public_ips" {
  description = "Public IPs of Kubernetes worker nodes"
  value       = aws_instance.workers[*].public_ip
}

output "worker_private_ips" {
  description = "Private IPs of Kubernetes worker nodes"
  value       = aws_instance.workers[*].private_ip
}

output "ansible_inventory" {
  description = "Ansible inventory snippet"
  value = templatefile("${path.module}/templates/inventory.tpl", {
    control_plane_ip = aws_instance.control_plane.public_ip
    worker_ips       = aws_instance.workers[*].public_ip
  })
}

output "vpc_id" {
  description = "VPC ID"
  value       = aws_vpc.main.id
}
