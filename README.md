# SRE Challenge - GeoIP Platform

A production-ready Site Reliability Engineering platform implementing:

- Infrastructure as Code
- Kubernetes cluster automation
- Go-based GeoIP API
- PostgreSQL High Availability
- Monitoring and Alerting
- Centralized Logging
- GitOps Deployment with ArgoCD


Repository:

https://github.com/soroush-hazeq-daneshvar/sre-challenge


---

# Architecture Overview


```
Developer
    |
    v
GitHub Repository
    |
    |
    +----------------------+
    |                      |
    v                      v

Terraform              Application Code
Ansible               Go GeoIP API

    |                      |
    |                      v

    |                Docker Build

    |                      |
    |                      v

    |                Docker Hub

    |        soroushmanhd/geoip-api:latest

    |                      |
    +----------+-----------+
               |
               v

             ArgoCD

               |
               v

        Kubernetes Cluster

               |
     +---------+----------+
     |         |          |
     v         v          v

 GeoIP API PostgreSQL Monitoring

            Logging

```


---

# Technology Stack


| Layer | Technology | Purpose |
|---|---|---|
| Infrastructure | Terraform | VM and network provisioning |
| Configuration | Ansible | OS configuration, Kubernetes installation |
| Container Runtime | containerd | Kubernetes container runtime |
| CNI | Calico | Kubernetes networking |
| Orchestration | Kubernetes 1.34 | Container orchestration |
| Database | CloudNativePG | HA PostgreSQL cluster |
| Application | Go + Gorilla Mux | GeoIP REST API |
| Container Registry | Docker Hub | Container image storage |
| Ingress | NGINX Ingress Controller | External traffic routing |
| Monitoring | Prometheus + Grafana | Metrics and dashboards |
| Alerting | Alertmanager | Alert management |
| Logging | Fluent Bit + Elasticsearch + Kibana | Centralized logging |
| GitOps | ArgoCD | Kubernetes deployment synchronization |


---

# Infrastructure Architecture


## Terraform

Location:

```
terraform/
```


Terraform provisions:

- Virtual machines
- Networking
- Security rules
- Infrastructure state


Example:

```bash
cd terraform

cp terraform.tfvars.example terraform.tfvars

terraform init

terraform plan

terraform apply
```


---

## Ansible

Location:

```
ansible/
```


Ansible configures:

- Linux hosts
- Kernel parameters
- Container runtime
- Kubernetes components
- Calico CNI


Example:

```bash
cd ansible

ansible-galaxy collection install -r requirements.yml

ansible-playbook site.yml
```


---

# Kubernetes Platform


Namespaces:


```
kubernetes/

├── base/
│   └── namespaces.yaml
│
├── geoip-api/
│   └── deployment.yaml
│
├── postgres/
│   └── cluster.yaml
│
├── monitoring/
│   ├── Prometheus
│   ├── Grafana
│   └── Alertmanager
│
└── logging/
    ├── Fluent Bit
    ├── Elasticsearch
    └── Kibana
```


---

# Quick Start


## 1. Provision Infrastructure


```bash
cd terraform

cp terraform.tfvars.example terraform.tfvars

terraform init

terraform apply
```


---

## 2. Configure Kubernetes Cluster


```bash
cd ../ansible

ansible-galaxy collection install -r requirements.yml

ansible-playbook site.yml
```


Verify:

```bash
kubectl get nodes
```


Expected:

```
NAME        STATUS
node-1      Ready
node-2      Ready
node-3      Ready
```


---

# 3. Install Cluster Components


## Create Namespaces


```bash
kubectl apply -f kubernetes/base/namespaces.yaml
```


---

## Install NGINX Ingress


```bash
helm repo add ingress-nginx \
https://kubernetes.github.io/ingress-nginx


helm repo update


helm upgrade --install ingress-nginx \
ingress-nginx/ingress-nginx \
-n ingress-nginx \
--create-namespace
```


---

## Install CloudNativePG


```bash
kubectl apply -f \
https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/main/releases/cnpg-1.26.0.yaml
```


Verify:

```bash
kubectl get pods -n cnpg-system
```


---

# 4. Deploy Platform Components


## PostgreSQL


```bash
kubectl apply -f kubernetes/postgres/
```


Check:

```bash
kubectl get pods -n postgres
```


---

## Monitoring


Install kube-prometheus-stack:


```bash
helm repo add prometheus-community \
https://prometheus-community.github.io/helm-charts


helm repo update


helm upgrade --install monitoring \
prometheus-community/kube-prometheus-stack \
-n monitoring \
--create-namespace \
-f kubernetes/monitoring/prometheus/helm-values.yaml
```


Apply dashboards and alerts:


```bash
kubectl apply -f kubernetes/monitoring/
```


---

## Logging


```bash
kubectl apply -f kubernetes/logging/
```


---

## GeoIP API


The application image:

```
soroushmanhd/geoip-api:latest
```


Deploy:


```bash
kubectl apply -f kubernetes/geoip-api/
```


Check:


```bash
kubectl get pods -n geoip
```


---

# 5. Deploy with ArgoCD


Install ArgoCD:


```bash
kubectl create namespace argocd


kubectl apply -n argocd \
-f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```


Deploy applications:


```bash
kubectl apply -f gitops/argocd/applications.yaml
```


ArgoCD watches:

```
https://github.com/soroush-hazeq-daneshvar/sre-challenge
```


---

# API Usage


## GeoIP Lookup


Request:


```bash
curl "http://geoip.local/country?ip=8.8.8.8"
```


Response:


```json
{
  "ip": "8.8.8.8",
  "country": "United States",
  "country_code": "US",
  "city": "Mountain View",
  "cache_hit": true,
  "source": "cache"
}
```


---

## Health Checks


```bash
curl http://geoip.local/health

curl http://geoip.local/ready

curl http://geoip.local/metrics
```


---

# Project Structure


```
sre-challenge/

├── terraform/
│   └── Infrastructure provisioning

├── ansible/
│   └── Kubernetes cluster automation

├── app/
│   └── geoip-api/
│       └── Go application

├── kubernetes/
│   ├── base/
│   ├── geoip-api/
│   ├── postgres/
│   ├── monitoring/
│   └── logging/

├── gitops/
│   └── argocd/

├── docs/
│   ├── architecture.md
│   ├── setup.md
│   ├── runbooks.md
│   ├── api.md
│   ├── cicd.md
│   └── proposal.md

└── README.md
```


---

# Documentation


| Document | Description |
|---|---|
| Architecture | System design and component details |
| Setup Guide | Complete installation guide |
| Runbooks | Operational troubleshooting procedures |
| API Docs | API reference |
| CI/CD | Deployment pipeline explanation |
| Troubleshooting | Common issues and solutions |
| Proposal | Architecture decisions and justification |


Links:


- [Architecture](docs/architecture.md)
- [Setup Guide](docs/setup.md)
- [Runbooks](docs/runbooks.md)
- [API Documentation](docs/api.md)
- [CI/CD](docs/cicd.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Proposal](docs/proposal.md)


---

# Observability


Monitoring stack:


```
Application Metrics

        |
        v

Prometheus

        |
        +------------+

        |            |

        v            v

     Grafana    Alertmanager
```


Logging:


```
Kubernetes Pods

        |

        v

Fluent Bit

        |

        v

Elasticsearch

        |

        v

Kibana
```


---

# Security Features


Implemented security practices:

- Non-root containers
- Distroless application image
- Kubernetes RBAC
- OS hardening with Ansible
- Network policies
- GitOps controlled deployment
- Infrastructure as Code


---
