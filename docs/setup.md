# راهنمای نصب و راه‌اندازی

# Setup Guide - GeoIP Platform

This document describes the complete deployment process of the GeoIP SRE Platform.

The deployment flow:

```
Terraform
    |
    v
Ansible
    |
    v
Kubernetes Cluster
    |
    +--> CloudNativePG PostgreSQL HA
    |
    +--> GeoIP API
    |
    +--> Monitoring Stack
    |
    +--> Logging Stack
    |
    +--> ArgoCD GitOps
```

---

# 1. Prerequisites

Required tools:

| Tool | Version | Purpose |
|------|---------|---------|
| Terraform | >= 1.5 | Infrastructure provisioning |
| Ansible | >= 2.14 | Configuration management |
| kubectl | >= 1.34 | Kubernetes management |
| Docker | >= 24 | Container build |
| Go | >= 1.22 | Application development |
| Helm | >= 3.12 | Kubernetes package management |

---

# 2. Build Application Image

The application image is hosted on Docker Hub.

Docker image:

```
soroushmanhd/geoip-api:latest
```


To pull the image manually:

```bash
docker pull soroushmanhd/geoip-api:latest
```


Build locally (optional):

```bash
cd app/geoip-api

docker build \
-t soroushmanhd/geoip-api:latest .
```


Push image:

```bash
docker login

docker push soroushmanhd/geoip-api:latest
```

---

# 3. Infrastructure Provisioning with Terraform

Terraform creates the required infrastructure.

Components:

- Virtual Machines
- Network
- Security Groups
- Required cloud resources


Go to Terraform directory:

```bash
cd terraform
```


Create variables:

```bash
cp terraform.tfvars.example terraform.tfvars
```


Edit variables:

```bash
vim terraform.tfvars
```


Initialize Terraform:

```bash
terraform init
```


Validate configuration:

```bash
terraform validate
```


Review changes:

```bash
terraform plan
```


Apply infrastructure:

```bash
terraform apply
```


Generate Ansible inventory:

```bash
terraform output -raw ansible_inventory \
> ../ansible/inventory/hosts.yml
```


Expected infrastructure:

```
1 Control Plane Node

2 Worker Nodes

Network
Security Groups
Storage
```

---

# 4. Kubernetes Cluster Installation with Ansible

Move to Ansible directory:

```bash
cd ../ansible
```


Install Ansible collections:

```bash
ansible-galaxy collection install \
-r requirements.yml
```


Test connectivity:

```bash
ansible all -m ping
```


Install Kubernetes cluster:

```bash
ansible-playbook site.yml
```


Verify cluster:

```bash
ssh ubuntu@<control-plane-ip>
```


Check nodes:

```bash
kubectl get nodes
```


Expected:

```
NAME              STATUS

control-plane     Ready

worker-01         Ready

worker-02         Ready
```

---

# 5. Install Kubernetes Components


## 5.1 Base Configuration

Create namespaces:

```bash
kubectl apply \
-f kubernetes/base/namespaces.yaml
```


Verify:

```bash
kubectl get namespaces
```


Expected:

```
geoip

postgres

monitoring

logging
```

---

# 5.2 NGINX Ingress Controller

The project uses ingress-nginx.

Add Helm repository:

```bash
helm repo add ingress-nginx \
https://kubernetes.github.io/ingress-nginx

# Expose ArgoCD through ingress
kubectl apply \
-f kubernetes/base/argocd-ingress.yaml

helm repo update
```


Install ingress controller:

```bash
helm upgrade --install ingress-nginx \
ingress-nginx/ingress-nginx \
--namespace ingress-nginx \
--create-namespace \
-f kubernetes/base/ingress-nginx.yaml
```


Verify:

```bash
kubectl get pods \
-n ingress-nginx
```


Expected:

```
ingress-nginx-controller   Running
```

---

# 5.3 StorageClass

The cluster uses Rancher Local Path Provisioner.

Verify:

```bash
kubectl get storageclass
```


Expected:

```
NAME

local-path
```


Used by:

- PostgreSQL
- Prometheus
- Grafana
- Elasticsearch

---

# 5.4 CloudNativePG Operator

Install CloudNativePG:

```bash
kubectl apply \
-f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.25/releases/cnpg-1.25.0.yaml
```


Verify:

```bash
kubectl get pods \
-n cnpg-system
```


Expected:

```
cnpg-controller-manager   Running
```


Deploy PostgreSQL HA cluster:

```bash
kubectl apply \
-f kubernetes/postgres/cluster.yaml
```


Verify:

```bash
kubectl get cluster \
-n postgres


kubectl get pods \
-n postgres
```

---

# 5.5 Monitoring Stack

Monitoring uses:

- Prometheus
- Grafana
- Alertmanager
- kube-state-metrics
- node-exporter


Add Helm repository:

```bash
helm repo add prometheus-community \
https://prometheus-community.github.io/helm-charts


helm repo update
```


Install kube-prometheus-stack:

```bash
helm upgrade --install kube-prometheus-stack \
prometheus-community/kube-prometheus-stack \
-n monitoring \
--create-namespace \
-f kubernetes/monitoring/prometheus/helm-values.yaml
```


Verify:

```bash
kubectl get pods \
-n monitoring
```


Deploy alerts:

```bash
kubectl apply \
-f kubernetes/monitoring/prometheus/alerts.yaml
```


Deploy Grafana dashboard:

```bash
kubectl apply \
-f kubernetes/monitoring/grafana/dashboard-geoip.yaml
```


Access Grafana:

```bash
kubectl port-forward \
svc/kube-prometheus-stack-grafana \
-n monitoring \
3000:80
```


URL:

```
http://localhost:3000
```


Get password:

```bash
kubectl get secret \
kube-prometheus-stack-grafana \
-n monitoring \
-o jsonpath="{.data.admin-password}" \
| base64 -d
```


Username:

```
admin
```

---

# 5.6 Logging Stack

Logging architecture:

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


## Elasticsearch

Deploy:

```bash
kubectl apply \
-f kubernetes/logging/elasticsearch/elasticsearch.yaml
```


Verify:

```bash
kubectl get pods \
-n logging
```


---

## Fluent Bit

Deploy:

```bash
kubectl apply \
-f kubernetes/logging/fluent-bit/daemonset.yaml
```


Verify:

```bash
kubectl get pods \
-n logging
```


---

## Kibana

Verify:

```bash
kubectl get pods \
-n logging
```


Access:

```
http://kibana.local
```

---

# 5.7 Deploy GeoIP API


The API deployment uses:

```
soroushmanhd/geoip-api:latest
```


Deploy:

```bash
kubectl apply \
-f kubernetes/geoip-api/
```


Verify:

```bash
kubectl get pods \
-n geoip
```


Check rollout:

```bash
kubectl rollout status \
deployment/geoip-api \
-n geoip
```

---

# 6. ArgoCD Installation


ArgoCD provides GitOps-based deployment.


Create namespace:

```bash
kubectl create namespace argocd
```


Install ArgoCD:

```bash
kubectl apply \
-n argocd \
-f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```


Wait for deployment:

```bash
kubectl wait \
--for=condition=Available \
deployment \
--all \
-n argocd \
--timeout=300s
```


Verify:

```bash
kubectl get pods \
-n argocd
```
Retrive the admin password 
```bash
kubectl get secret \
argocd-initial-admin-secret \
-n argocd \
-o jsonpath="{.data.password}" \
| base64 -d
```

---

# 7. Deploy ArgoCD Applications


Apply GitOps applications:

```bash
kubectl apply \
-f gitops/argocd/applications.yaml
```


Managed applications:

| Application | Namespace |
|-------------|-----------|
| geoip-api | geoip |
| postgres | postgres |
| monitoring | monitoring |
| logging | logging |


Check status:

```bash
kubectl get applications \
-n argocd
```

---

# 8. Access ArgoCD UI


Port forward:

```bash
kubectl port-forward \
svc/argocd-server \
-n argocd \
8080:443
```


Open:

```
https://localhost:8080
```


Get password:

```bash
kubectl get secret \
argocd-initial-admin-secret \
-n argocd \
-o jsonpath="{.data.password}" \
| base64 -d
```


Username:

```
admin
```

---

# 9. Application Testing


Add hosts:

```bash
echo "<ingress-ip> geoip.local kibana.local" \
| sudo tee -a /etc/hosts
```


Test API:

```bash
curl \
"http://geoip.local/country?ip=8.8.8.8"
```


Expected response:

```json
{
  "ip": "8.8.8.8",
  "country": "United States",
  "cache_hit": false
}
```


Health:

```bash
curl http://geoip.local/health
```


Ready:

```bash
curl http://geoip.local/ready
```


Metrics:

```bash
curl http://geoip.local/metrics
```

---

# 10. CI/CD

CI/CD architecture:

```
Developer

   |

   v

GitLab CI

   |

   v

Docker Build

   |

   v

Docker Hub

soroushmanhd/geoip-api

   |

   v

ArgoCD

   |

   v

Kubernetes
```


Pipeline responsibilities:

| Component | Responsibility |
|-|-|
| GitLab CI | Build, Test |
| Docker Hub | Container Registry |
| ArgoCD | Continuous Deployment |

---

# 11. Environment Variables

| Variable | Default | Description |
|-|-|-|
| LISTEN_ADDR | :8080 | API listen address |
| DATABASE_URL | postgres:// | PostgreSQL connection |
| GEOIP_PROVIDER_URL | https://ipapi.co | GeoIP provider |
| GEOIP_PROVIDER_TIMEOUT | 5s | Provider timeout |

---

# 12. Rollback


Kubernetes rollback:

```bash
kubectl rollout undo \
deployment/geoip-api \
-n geoip
```


GitOps rollback:

```bash
git revert <commit>

git push origin main
```


ArgoCD automatically restores the previous state.