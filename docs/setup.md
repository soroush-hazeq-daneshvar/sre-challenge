# راهنمای نصب و راه‌اندازی

# Setup Guide - GeoIP Platform

This document describes the complete deployment process of the GeoIP SRE Platform.

The deployment flow:

```text
Terraform
    |
    v
Ansible
    |
    v
Kubernetes Cluster
    |
    +--> CloudNativePG PostgreSQL
    |
    +--> GeoIP API
    |
    +--> Monitoring Stack
    |       +--> Prometheus
    |       +--> Grafana
    |       +--> Alertmanager
    |
    +--> Logging Stack
    |       +--> Elasticsearch
    |       +--> Fluent Bit
    |       +--> Kibana
    |
    +--> NGINX Ingress
    |
    +--> ArgoCD GitOps
```

---

# 1. Prerequisites

Required tools:

| Tool      | Version | Purpose                       |
| --------- | ------- | ----------------------------- |
| Terraform | >= 1.5  | Infrastructure provisioning   |
| Ansible   | >= 2.14 | Configuration management      |
| kubectl   | >= 1.34 | Kubernetes management         |
| Docker    | >= 24   | Container build               |
| Go        | >= 1.22 | Application development       |
| Helm      | >= 3.12 | Kubernetes package management |

The Kubernetes cluster used by this project consists of:

```text
1 Control Plane
2 Worker Nodes
```

Example:

```text
Control Plane:
stg-hazegh-vm1.dc.snappcloud.io

Worker:
stg-hazegh-vm2.dc.snappcloud.io
stg-hazegh-vm3.dc.snappcloud.io
```

Verify the cluster:

```bash
kubectl get nodes -o wide
```

---

# 2. Build Application Image

The application image is hosted on Docker Hub.

Docker image:

```text
soroushmanhd/geoip-api:latest
```

Pull the image manually:

```bash
docker pull soroushmanhd/geoip-api:latest
```

Build locally:

```bash
cd app/geoip-api

docker build \
  -t soroushmanhd/geoip-api:latest \
  .
```

Login to Docker Hub:

```bash
docker login
```

Push the image:

```bash
docker push soroushmanhd/geoip-api:latest
```

---

# 3. Infrastructure Provisioning with Terraform

Terraform creates the required infrastructure.

Components:

* Virtual Machines
* Network
* Security Groups
* Required cloud resources

Go to the Terraform directory:

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

Generate the Ansible inventory:

```bash
terraform output -raw ansible_inventory \
  > ../ansible/inventory/hosts.yml
```

Expected infrastructure:

```text
1 Control Plane Node
2 Worker Nodes

Network
Security Groups
Storage
```

---

# 4. Kubernetes Cluster Installation with Ansible

Move to the Ansible directory:

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

Install the Kubernetes cluster:

```bash
ansible-playbook site.yml
```

Connect to the control-plane node:

```bash
ssh ubuntu@<control-plane-ip>
```

Check nodes:

```bash
kubectl get nodes
```

Expected:

```text
NAME              STATUS   ROLES
control-plane     Ready    control-plane
worker-01         Ready    <none>
worker-02         Ready    <none>
```

---

# 5. Install Kubernetes Components

## 5.1 Base Configuration

Create the required namespaces:

```bash
kubectl apply \
  -f kubernetes/base/namespaces.yaml
```

Verify:

```bash
kubectl get namespaces
```

Expected:

```text
geoip
postgres
monitoring
logging
```

---

# 5.2 NGINX Ingress Controller

The project uses `ingress-nginx`.

Add the Helm repository:

```bash
helm repo add ingress-nginx \
  https://kubernetes.github.io/ingress-nginx
```

Update Helm repositories:

```bash
helm repo update
```

Install or upgrade the ingress controller:

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

```text
ingress-nginx-controller   Running
```

Verify the service:

```bash
kubectl get svc \
  -n ingress-nginx \
  ingress-nginx-controller
```

The current cluster exposes NGINX using a `LoadBalancer` service with NodePorts.

Example:

```text
HTTP   80:31406/TCP
HTTPS  443:31469/TCP
```

Because the environment does not provide an external LoadBalancer IP, the NodePort is used for external access.

Check the actual assigned ports:

```bash
kubectl get svc \
  -n ingress-nginx \
  ingress-nginx-controller
```

---

# 5.3 StorageClass

The cluster uses Rancher Local Path Provisioner.

Verify:

```bash
kubectl get storageclass
```

Expected:

```text
NAME
local-path
```

The `local-path` StorageClass is used by stateful workloads such as:

* PostgreSQL
* Prometheus
* Grafana
* Elasticsearch

Verify:

```bash
kubectl get storageclass local-path
```

---

# 5.4 CloudNativePG Operator

The project uses CloudNativePG.

Install the operator:

```bash
kubectl apply \
  -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.26/releases/cnpg-1.26.0.yaml
```

Verify:

```bash
kubectl get pods \
  -n cnpg-system
```

Expected:

```text
cnpg-controller-manager   Running
```

Verify the operator version:

```bash
kubectl get deployment \
  -n cnpg-system
```

Deploy PostgreSQL:

```bash
kubectl apply \
  -f kubernetes/postgres/cluster.yaml
```

Verify the cluster:

```bash
kubectl get cluster \
  -n postgres
```

Verify PostgreSQL pods:

```bash
kubectl get pods \
  -n postgres
```

Verify PostgreSQL services:

```bash
kubectl get svc \
  -n postgres
```

Expected services:

```text
geoip-postgres-rw
geoip-postgres-ro
geoip-postgres-r
```

The application connects to the read-write service:

```text
geoip-postgres-rw.postgres.svc:5432
```

---

# 5.5 Monitoring Stack

The monitoring stack consists of:

* Prometheus
* Grafana
* Alertmanager
* kube-state-metrics
* node-exporter

Add the Prometheus Community repository:

```bash
helm repo add prometheus-community \
  https://prometheus-community.github.io/helm-charts
```

Update repositories:

```bash
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

Deploy the GeoIP dashboard:

```bash
kubectl apply \
  -f kubernetes/monitoring/grafana/dashboard-geoip.yaml
```

Verify monitoring services:

```bash
kubectl get svc \
  -n monitoring
```

Expected services include:

```text
kube-prometheus-stack-grafana
kube-prometheus-stack-prometheus
kube-prometheus-stack-alertmanager
```

---

## 5.5.1 Grafana

Grafana is exposed through the NGINX Ingress controller.

Ingress:

```text
grafana.local
```

Verify:

```bash
kubectl get ingress \
  -n monitoring
```

Expected:

```text
grafana   nginx   grafana.local
```

Get the Grafana password:

```bash
kubectl get secret \
  kube-prometheus-stack-grafana \
  -n monitoring \
  -o jsonpath="{.data.admin-password}" \
  | base64 -d
```

Username:

```text
admin
```

### Local Access

If using the ingress NodePort directly:

```bash
curl \
  --noproxy '*' \
  -H 'Host: grafana.local' \
  http://<NODE-IP>:31406
```

Grafana should return:

```text
HTTP/1.1 302 Found
Location: /login
```

You can also use port forwarding:

```bash
kubectl port-forward \
  -n monitoring \
  svc/kube-prometheus-stack-grafana \
  3000:80
```

Then open:

```text
http://localhost:3000
```

---

# 5.6 Prometheus

Prometheus is exposed through the NGINX Ingress controller.

Ingress:

```text
prometheus.local
```

Verify:

```bash
kubectl get ingress \
  -n monitoring
```

Expected:

```text
prometheus   nginx   prometheus.local
```

Test through the ingress NodePort:

```bash
curl \
  --noproxy '*' \
  -H 'Host: prometheus.local' \
  http://<NODE-IP>:31406
```

Expected response:

```text
HTTP/1.1 302 Found
Location: /query
```

Port-forward alternative:

```bash
kubectl port-forward \
  -n monitoring \
  svc/kube-prometheus-stack-prometheus \
  9090:9090
```

Open:

```text
http://localhost:9090
```

---

# 5.7 Logging Stack

Logging architecture:

```text
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

## 5.7.1 Elasticsearch

Deploy Elasticsearch:

```bash
kubectl apply \
  -f kubernetes/logging/elasticsearch/elasticsearch.yaml
```

Verify:

```bash
kubectl get pods \
  -n logging
```

Verify the service:

```bash
kubectl get svc \
  -n logging
```

Expected:

```text
elasticsearch   ClusterIP   9200/TCP
```

---

## 5.7.2 Fluent Bit

Deploy Fluent Bit:

```bash
kubectl apply \
  -f kubernetes/logging/fluent-bit/daemonset.yaml
```

Verify:

```bash
kubectl get pods \
  -n logging
```

Fluent Bit should run as a DaemonSet so that logs can be collected from each Kubernetes node.

---

## 5.7.3 Kibana

Verify:

```bash
kubectl get pods \
  -n logging
```

Verify the ingress:

```bash
kubectl get ingress \
  -n logging
```

Expected:

```text
kibana   nginx   kibana.local
```

Test through the ingress NodePort:

```bash
curl \
  --noproxy '*' \
  -H 'Host: kibana.local' \
  http://<NODE-IP>:31406
```

Kibana should redirect to:

```text
/spaces/enter
```

---

# 5.8 Monitoring and Logging External Access

The following services are exposed through NGINX Ingress:

| Service    | Host               | Backend    |
| ---------- | ------------------ | ---------- |
| ArgoCD     | `argocd.local`     | ArgoCD     |
| GeoIP API  | `geoip.local`      | GeoIP API  |
| Grafana    | `grafana.local`    | Grafana    |
| Prometheus | `prometheus.local` | Prometheus |
| Kibana     | `kibana.local`     | Kibana     |

Verify all ingresses:

```bash
kubectl get ingress -A
```

Expected:

```text
argocd       argocd-server   nginx   argocd.local
geoip        geoip-api       nginx   geoip.local
logging      kibana          nginx   kibana.local
monitoring   grafana         nginx   grafana.local
monitoring   prometheus      nginx   prometheus.local
```

---

## 5.8.1 Configure `/etc/hosts`

Because the cluster does not have an external LoadBalancer IP, map the hostnames to a Kubernetes node IP.

For example:

```text
10.160.1.22 argocd.local
10.160.1.22 geoip.local
10.160.1.22 grafana.local
10.160.1.22 prometheus.local
10.160.1.22 kibana.local
```

Edit:

```bash
sudo vim /etc/hosts
```

Or:

```bash
echo "10.160.1.22 argocd.local" | sudo tee -a /etc/hosts
echo "10.160.1.22 geoip.local" | sudo tee -a /etc/hosts
echo "10.160.1.22 grafana.local" | sudo tee -a /etc/hosts
echo "10.160.1.22 prometheus.local" | sudo tee -a /etc/hosts
echo "10.160.1.22 kibana.local" | sudo tee -a /etc/hosts
```

Verify DNS resolution:

```bash
ping -c 1 grafana.local
ping -c 1 prometheus.local
ping -c 1 kibana.local
```

---

## 5.8.2 Access Through the Ingress NodePort

The NGINX ingress controller exposes:

```text
HTTP   NodePort 31406
HTTPS  NodePort 31469
```

Check the actual values:

```bash
kubectl get svc \
  -n ingress-nginx \
  ingress-nginx-controller
```

For example:

```text
80:31406/TCP
443:31469/TCP
```

Test Grafana:

```bash
curl \
  --noproxy '*' \
  -H 'Host: grafana.local' \
  http://10.160.1.22:31406
```

Test Prometheus:

```bash
curl \
  --noproxy '*' \
  -H 'Host: prometheus.local' \
  http://10.160.1.22:31406
```

Test Kibana:

```bash
curl \
  --noproxy '*' \
  -H 'Host: kibana.local' \
  http://10.160.1.22:31406
```

Test ArgoCD:

```bash
curl \
  --noproxy '*' \
  -H 'Host: argocd.local' \
  http://10.160.1.22:31406
```

ArgoCD redirects HTTP to HTTPS:

```text
HTTP/1.1 308 Permanent Redirect
Location: https://argocd.local
```

Test GeoIP API:

```bash
curl \
  --noproxy '*' \
  -H 'Host: geoip.local' \
  http://10.160.1.22:31406/health
```

Expected:

```json
{
  "status": "healthy"
}
```

---

# 5.9 Deploy GeoIP API

The API deployment uses:

```text
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

Verify the service:

```bash
kubectl get svc \
  -n geoip
```

Verify ingress:

```bash
kubectl get ingress \
  -n geoip
```

Expected:

```text
geoip-api   nginx   geoip.local
```

---

# 6. ArgoCD Installation

ArgoCD provides GitOps-based deployment.

Create the namespace:

```bash
kubectl create namespace argocd
```

Install ArgoCD:

```bash
kubectl apply \
  -n argocd \
  -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

Wait for deployments:

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

Retrieve the initial admin password:

```bash
kubectl get secret \
  argocd-initial-admin-secret \
  -n argocd \
  -o jsonpath="{.data.password}" \
  | base64 -d
```

Username:

```text
admin
```

---

# 7. Deploy ArgoCD Applications

Apply GitOps applications:

```bash
kubectl apply \
  -f gitops/argocd/applications.yaml
```

Managed applications:

| Application | Namespace  |
| ----------- | ---------- |
| geoip-api   | geoip      |
| postgres    | postgres   |
| monitoring  | monitoring |
| logging     | logging    |

Check status:

```bash
kubectl get applications \
  -n argocd
```

Expected:

```text
NAME
geoip-api
postgres
monitoring
logging
```

---

# 8. Access ArgoCD UI

ArgoCD is exposed through the NGINX Ingress.

Verify:

```bash
kubectl get ingress \
  -n argocd
```

Expected:

```text
argocd-server   nginx   argocd.local
```

Add the hostname to `/etc/hosts`:

```text
10.160.1.22 argocd.local
```

Because ArgoCD redirects HTTP to HTTPS, access:

```text
https://argocd.local
```

If the NodePort is required explicitly:

```text
https://argocd.local:31469
```

Alternatively, use port forwarding:

```bash
kubectl port-forward \
  svc/argocd-server \
  -n argocd \
  8080:443
```

Open:

```text
https://localhost:8080
```

Get the initial password:

```bash
kubectl get secret \
  argocd-initial-admin-secret \
  -n argocd \
  -o jsonpath="{.data.password}" \
  | base64 -d
```

Username:

```text
admin
```

---

# 9. Application Testing

Configure the local hostnames:

```bash
sudo vim /etc/hosts
```

Example:

```text
10.160.1.22 argocd.local
10.160.1.22 geoip.local
10.160.1.22 grafana.local
10.160.1.22 prometheus.local
10.160.1.22 kibana.local
```

Check ingress configuration:

```bash
kubectl get ingress -A
```

---

## 9.1 GeoIP API

Test the API through the ingress NodePort:

```bash
curl \
  --noproxy '*' \
  -H 'Host: geoip.local' \
  http://10.160.1.22:31406/country?ip=8.8.8.8
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
curl \
  --noproxy '*' \
  -H 'Host: geoip.local' \
  http://10.160.1.22:31406/health
```

Expected:

```json
{
  "status": "healthy"
}
```

Ready:

```bash
curl \
  --noproxy '*' \
  -H 'Host: geoip.local' \
  http://10.160.1.22:31406/ready
```

Metrics:

```bash
curl \
  --noproxy '*' \
  -H 'Host: geoip.local' \
  http://10.160.1.22:31406/metrics
```

---

## 9.2 Grafana

Open:

```text
http://grafana.local
```

If direct port 80 is not exposed by the environment, use:

```text
http://grafana.local:31406
```

or access through the node IP:

```text
http://10.160.1.22:31406
```

with:

```text
Host: grafana.local
```

---

## 9.3 Prometheus

Open:

```text
http://prometheus.local
```

If direct port 80 is not exposed:

```text
http://prometheus.local:31406
```

---

## 9.4 Kibana

Open:

```text
http://kibana.local
```

If direct port 80 is not exposed:

```text
http://kibana.local:31406
```

---

## 9.5 ArgoCD

Open:

```text
https://argocd.local
```

If direct HTTPS port 443 is not exposed:

```text
https://argocd.local:31469
```

---

# 10. Useful Troubleshooting Commands

Check all nodes:

```bash
kubectl get nodes -o wide
```

Check all pods:

```bash
kubectl get pods -A
```

Check all services:

```bash
kubectl get svc -A
```

Check all ingresses:

```bash
kubectl get ingress -A
```

Check ingress controller:

```bash
kubectl get pods \
  -n ingress-nginx

kubectl get svc \
  -n ingress-nginx
```

Check ingress controller logs:

```bash
kubectl logs \
  -n ingress-nginx \
  deployment/ingress-nginx-controller
```

Check a specific ingress:

```bash
kubectl describe ingress \
  -n monitoring \
  grafana
```

Check service endpoints:

```bash
kubectl get endpoints \
  -n monitoring
```

Check EndpointSlices:

```bash
kubectl get endpointslice \
  -n monitoring
```

Test an ingress route directly:

```bash
curl \
  --noproxy '*' \
  -v \
  -H 'Host: grafana.local' \
  http://10.160.1.22:31406
```

---

# 11. CI/CD

CI/CD architecture:

```text
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
    |
    | soroushmanhd/geoip-api
    v
ArgoCD
    |
    v
Kubernetes
```

Pipeline responsibilities:

| Component  | Responsibility        |
| ---------- | --------------------- |
| GitLab CI  | Build, Test           |
| Docker Hub | Container Registry    |
| ArgoCD     | Continuous Deployment |
| Kubernetes | Runtime               |

---

# 12. Environment Variables

| Variable               | Default          | Description           |
| ---------------------- | ---------------- | --------------------- |
| LISTEN_ADDR            | :8080            | API listen address    |
| DATABASE_URL           | postgres://      | PostgreSQL connection |
| GEOIP_PROVIDER_URL     | https://ipapi.co | GeoIP provider        |
| GEOIP_PROVIDER_TIMEOUT | 5s               | Provider timeout      |

---

# 13. Rollback

Kubernetes rollback:

```bash
kubectl rollout undo \
  deployment/geoip-api \
  -n geoip
```

Check rollout history:

```bash
kubectl rollout history \
  deployment/geoip-api \
  -n geoip
```

GitOps rollback:

```bash
git revert <commit>

git push origin main
```

ArgoCD automatically reconciles the Kubernetes cluster with the Git repository state.

---

# 14. Final Verification

Before considering the platform operational, verify:

### Kubernetes

```bash
kubectl get nodes
kubectl get pods -A
```

### PostgreSQL

```bash
kubectl get cluster -n postgres
kubectl get pods -n postgres
kubectl get pvc -n postgres
```

### GeoIP API

```bash
kubectl get pods -n geoip
kubectl get ingress -n geoip
```

### Monitoring

```bash
kubectl get pods -n monitoring
kubectl get ingress -n monitoring
```

### Logging

```bash
kubectl get pods -n logging
kubectl get ingress -n logging
```

### ArgoCD

```bash
kubectl get pods -n argocd
kubectl get applications -n argocd
kubectl get ingress -n argocd
```

### Ingress

```bash
kubectl get svc \
  -n ingress-nginx \
  ingress-nginx-controller

kubectl get ingress -A
```

Expected external endpoints:

```text
GeoIP API       http://geoip.local
Grafana         http://grafana.local
Prometheus      http://prometheus.local
Kibana          http://kibana.local
ArgoCD          https://argocd.local
```

When direct ports 80/443 are unavailable, use the ingress NodePorts:

```text
HTTP  -> 31406
HTTPS -> 31469
```
