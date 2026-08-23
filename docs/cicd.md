# CI/CD Pipeline Documentation

## Overview

The project uses a CI/CD pipeline for automated validation, testing, container image building, and Kubernetes deployment.

The deployment model follows GitOps principles:

- Source code management: GitHub
- Container registry: Docker Hub
- Continuous Deployment: ArgoCD
- Infrastructure provisioning: Terraform
- Configuration management: Ansible

Architecture:

```
Developer
    |
    v
GitHub Repository
    |
    v
CI Pipeline
    |
    +--> Validate Code
    |
    +--> Run Tests
    |
    +--> Build Docker Image
    |
    +--> Push Image to Docker Hub
    |
    v
Update Kubernetes Manifest
    |
    v
ArgoCD Sync
    |
    v
Kubernetes Cluster
```

---

# Pipeline Stages

```
validate → test → build → deploy
```

---

# Stage 1: Validate

The validation stage checks infrastructure, application code, and Kubernetes manifests.

| Job | Tool | Purpose |
|-----|------|---------|
| validate_terraform | Terraform | Validate IaC syntax |
| validate_ansible | ansible-lint | Validate Ansible roles and playbooks |
| lint_go | go fmt, go vet | Static analysis for Go application |
| validate_k8s | kubectl/kubeconform | Kubernetes manifest validation |

Examples:

```bash
terraform fmt -check
terraform validate

ansible-lint ansible/

go fmt ./...
go vet ./...

kubectl apply --dry-run=client -f kubernetes/
```

---

# Stage 2: Test

Application tests are executed before building the container image.

| Job | Tool | Purpose |
|-----|------|---------|
| test_go | go test | Execute unit tests |
| integration_test | PostgreSQL container | Validate database integration |

Example:

```bash
cd app/geoip-api

go test ./...
```

---

# Stage 3: Build Docker Image

The GeoIP API is packaged as a container image.

Dockerfile characteristics:

- Multi-stage Go build
- Static binary compilation
- Distroless runtime image
- Non-root container execution


Image:

```
soroushmanhd/geoip-api:latest
```

Build example:

```bash
cd app/geoip-api

docker build \
-t soroushmanhd/geoip-api:latest .
```

Push:

```bash
docker login

docker push soroushmanhd/geoip-api:latest
```

Image security features:

- Runs as non-root user
- Minimal runtime filesystem
- No unnecessary packages
- Small attack surface

---

# Stage 4: Deployment

Deployment is managed by ArgoCD.

The CI pipeline does not directly execute kubectl against production.

Flow:

```
CI Pipeline
      |
      v
Docker Hub Image
      |
      v
Kubernetes Manifest Update
      |
      v
GitHub Commit
      |
      v
ArgoCD detects change
      |
      v
Kubernetes Deployment
```

---

# GitOps with ArgoCD

ArgoCD watches:

```
gitops/argocd/applications.yaml
```

Repository:

```
https://github.com/soroush-hazeq-daneshvar/sre-challenge
```

Application structure:

| Application | Path | Namespace | Sync |
|-------------|------|-----------|------|
| geoip-api | kubernetes/geoip-api | geoip | Automated |
| postgres | kubernetes/postgres | postgres | Automated |
| monitoring | kubernetes/monitoring | monitoring | Automated |
| logging | kubernetes/logging | logging | Automated |

---

# ArgoCD Sync Policy

Applications use:

```yaml
automated:
  prune: true
  selfHeal: true
```

Meaning:

## Self Heal

If somebody changes resources manually:

```
kubectl edit deployment geoip-api
```

ArgoCD restores the Git defined state.

## Prune

Removed resources from Git are automatically deleted from Kubernetes.

---

# Branch Strategy

```
main
 |
 +---- feature branches
```

Recommended workflow:

```
Feature branch
        |
        v
Pull Request
        |
        v
Validation + Tests
        |
        v
Merge to main
        |
        v
Build Image
        |
        v
ArgoCD Deployment
```

---

# Deployment Flow

1. Developer changes application code.

2. Push changes to GitHub.

3. CI pipeline starts.

4. Pipeline executes:

```
validate
test
build
push image
```

5. New Docker image is published:

```
soroushmanhd/geoip-api:<tag>
```

6. Kubernetes manifest image tag is updated.

7. ArgoCD detects Git change.

8. ArgoCD synchronizes Kubernetes resources.

9. New application version becomes available.

---

# Required CI/CD Secrets

The following secrets are required:

| Variable | Description |
|----------|-------------|
| DOCKERHUB_USERNAME | Docker Hub username |
| DOCKERHUB_TOKEN | Docker Hub access token |
| KUBECONFIG | Kubernetes cluster access configuration |

---

# Infrastructure Deployment Flow

Infrastructure is separated from application deployment.

## Infrastructure

Terraform:

```
terraform/
```

Responsible for:

- VM provisioning
- Network creation
- Security groups
- Infrastructure state


Example:

```bash
cd terraform

terraform init

terraform plan

terraform apply
```


## Configuration

Ansible:

```
ansible/
```

Responsible for:

- OS configuration
- Container runtime
- Kubernetes installation
- Cluster initialization


Example:

```bash
cd ansible

ansible-playbook site.yml
```

---

# Rollback Procedure

## Kubernetes Rollback

```bash
kubectl rollout history deployment/geoip-api -n geoip

kubectl rollout undo deployment/geoip-api -n geoip
```

---

## ArgoCD Rollback

```bash
argocd app rollback geoip-api
```

---

## GitOps Rollback

Recommended method:

```bash
git revert <commit-id>

git push origin main
```

ArgoCD automatically restores the previous state.

---

# Current Deployment Components

```
Kubernetes Cluster

├── geoip-api
│   └── soroushmanhd/geoip-api:latest
│
├── PostgreSQL
│   └── CloudNativePG
│
├── Monitoring
│   ├── Prometheus
│   ├── Grafana
│   └── Alertmanager
│
└── Logging
    ├── Fluent Bit
    ├── Elasticsearch
    └── Kibana
```

---

# Final Architecture

```
GitHub
  |
  |
CI Pipeline
  |
  |
Docker Hub
  |
  |
ArgoCD
  |
  |
Kubernetes
  |
  +-- GeoIP API
  |
  +-- PostgreSQL HA
  |
  +-- Monitoring Stack
  |
  +-- Logging Stack
```

This provides:

- Automated delivery
- Immutable container deployment
- Git-based audit history
- Kubernetes self-healing
- Reproducible infrastructure