# راهنمای نصب و راه‌اندازی

## پیش‌نیازها

| Tool | Version | Purpose |
|------|---------|---------|
| Terraform | >= 1.5 | Infrastructure provisioning |
| Ansible | >= 2.14 | Configuration management |
| kubectl | >= 1.34 | Kubernetes management |
| Docker | >= 24 | Container builds |
| Go | >= 1.22 | Application development |
| Helm | >= 3.12 | Package management |

## مرحله 1: Provisioning با Terraform

```bash
cd terraform

# تنظیم متغیرها
cp terraform.tfvars.example terraform.tfvars
# ویرایش terraform.tfvars و IP خود را در allowed_cidr_blocks قرار دهید

# اجرا
terraform init
terraform plan
terraform apply

# ذخیره inventory برای Ansible
terraform output -raw ansible_inventory > ../ansible/inventory/hosts.yml
```

**خروجی:**
- 1 Control Plane VM (t3.medium)
- 2 Worker VM (t3.large)
- VPC, Subnets, Security Groups

## مرحله 2: راه‌اندازی Kubernetes با Ansible

```bash
cd ansible
ansible-galaxy collection install -r requirements.yml
ansible all -m ping
ansible-playbook site.yml

# بررسی cluster
ssh ubuntu@<control-plane-ip>
kubectl get nodes
# Expected: 3 nodes Ready
```

## مرحله 3: نصب Cluster Add-ons

```bash
# NGINX Ingress Controller
kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.0/deploy/static/provider/cloud/deploy.yaml

# cert-manager
kubectl apply -f https://github.com/cert-manager/cert-manager/releases/download/v1.14.0/cert-manager.yaml

# CloudNativePG Operator
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.22/releases/cnpg-1.22.0.yaml

# Metrics Server
kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# ArgoCD
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

## مرحله 4: Deploy Platform Components

```bash
# Namespaces
kubectl apply -f kubernetes/base/namespaces.yaml

# PostgreSQL HA Cluster
kubectl apply -f kubernetes/postgres/cluster.yaml
kubectl wait --for=condition=Ready cluster/geoip-postgres -n postgres --timeout=300s

# Monitoring Stack
helm repo add prometheus-community https://prometheus-community.github.io/helm-charts
helm install prometheus prometheus-community/kube-prometheus-stack \
  -n monitoring --create-namespace \
  -f kubernetes/monitoring/prometheus/helm-values.yaml
kubectl apply -f kubernetes/monitoring/prometheus/alerts.yaml
kubectl apply -f kubernetes/monitoring/grafana/dashboard-geoip.yaml

# Logging Stack
kubectl apply -f kubernetes/logging/elasticsearch/elasticsearch.yaml
kubectl apply -f kubernetes/logging/fluent-bit/daemonset.yaml

# GeoIP API
kubectl apply -f kubernetes/geoip-api/deployment.yaml
kubectl rollout status deployment/geoip-api -n geoip
```

## مرحله 5: GitOps با ArgoCD

```bash
# Deploy ArgoCD applications
kubectl apply -f gitops/argocd/applications.yaml

# دسترسی به ArgoCD UI
kubectl port-forward svc/argocd-server -n argocd 8080:443
# Username: admin
# Password: kubectl -n argocd get secret argocd-initial-admin-secret -o jsonpath="{.data.password}" | base64 -d
```

## مرحله 6: تست

```bash
# Add to /etc/hosts
echo "<ingress-ip> geoip.local kibana.local" | sudo tee -a /etc/hosts

# Test API
curl "http://geoip.local/country?ip=8.8.8.8"
curl "http://geoip.local/country?ip=1.1.1.1"
curl "http://geoip.local/metrics"

# Test cache (second request should be faster)
curl "http://geoip.local/country?ip=8.8.8.8" | jq .cache_hit
# Expected: true

# Access Grafana
kubectl port-forward svc/prometheus-grafana -n monitoring 3000:80
# http://localhost:3000 (admin/changeme)

# Access Kibana
# http://kibana.local:5601
```

## CI/CD Setup (GitLab)

1. Push repository to GitLab
2. Configure CI/CD variables:
   - `CI_REGISTRY_USER` / `CI_REGISTRY_PASSWORD`
   - `KUBE_CONTEXT_STAGING` / `KUBE_CONTEXT_PRODUCTION`
3. Pipeline runs automatically on push

## Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `LISTEN_ADDR` | `:8080` | Server listen address |
| `DATABASE_URL` | postgres://... | PostgreSQL connection string |
| `GEOIP_PROVIDER_URL` | `https://ipapi.co` | External GeoIP provider |
| `GEOIP_PROVIDER_TIMEOUT` | `5s` | Provider request timeout |
