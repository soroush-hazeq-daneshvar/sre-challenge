# SRE Challenge - GeoIP Platform

A production-ready Site Reliability Engineering challenge implementation featuring Infrastructure as Code, Kubernetes, GeoIP API, PostgreSQL HA, Monitoring, Logging, and GitOps.

## Architecture Overview

```
Terraform → Ansible → Kubernetes → PostgreSQL → GeoIP API → Monitoring → ELK
```

| Layer | Technology | Purpose |
|-------|-----------|---------|
| Infrastructure | Terraform (AWS) | VM provisioning (1 CP + 2 Workers) |
| Configuration | Ansible | OS hardening, K8s, Calico CNI |
| Orchestration | Kubernetes 1.34 | Container orchestration |
| Database | CloudNativePG | HA PostgreSQL (1 Primary + 2 Replicas) |
| Application | Go + Gorilla Mux | GeoIP lookup API with caching |
| Ingress | NGINX Ingress | External traffic routing |
| Monitoring | Prometheus + Grafana | Metrics & dashboards |
| Alerting | Alertmanager | Email, Slack, Telegram notifications |
| Logging | Fluent Bit + ELK | Centralized log aggregation |
| CI/CD | GitLab CI | Build, test, deploy pipeline |
| GitOps | ArgoCD | Declarative cluster sync |

## Quick Start

```bash
# 1. Provision infrastructure
cd terraform
cp terraform.tfvars.example terraform.tfvars
terraform init && terraform apply

# 2. Configure Kubernetes cluster
cd ../ansible
ansible-galaxy collection install -r requirements.yml
ansible-playbook site.yml

# 3. Deploy platform components
kubectl apply -f kubernetes/base/namespaces.yaml
helm repo add ingress-nginx https://kubernetes.github.io/ingress-nginx

helm repo update

helm upgrade --install ingress-nginx \
  ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  --values kubernetes/base/ingress-nginx.yaml
kubectl apply -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/main/releases/cnpg-1.26.0.yaml
kubectl apply --server-side \
  -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/v1.26.0/config/crd/bases/postgresql.cnpg.io_poolers.yaml
  
kubectl apply -f kubernetes/postgres/
kubectl apply -f kubernetes/monitoring/
kubectl apply -f kubernetes/logging/
kubectl apply -f kubernetes/geoip-api/

# 4. Test the API
curl "http://geoip.local/country?ip=8.8.8.8"
```

## Project Structure

```
sre-challenge/
├── terraform/          # AWS VM provisioning
├── ansible/            # K8s cluster setup & hardening
├── app/geoip-api/      # Go GeoIP API application
├── kubernetes/         # K8s manifests
│   ├── base/           # Namespaces
│   ├── postgres/       # CloudNativePG cluster
│   ├── geoip-api/      # Application deployment
│   ├── monitoring/     # Prometheus, Grafana, Alertmanager
│   └── logging/        # Fluent Bit, Elasticsearch, Kibana
├── gitops/argocd/      # ArgoCD applications
├── docs/               # Full documentation
└── .gitlab-ci.yml      # CI/CD pipeline
```

## API Usage

```bash
# Get country for an IP
GET /country?ip=8.8.8.8

# Response
{
  "ip": "8.8.8.8",
  "country": "United States",
  "country_code": "US",
  "city": "Mountain View",
  "cache_hit": true,
  "source": "cache"
}

# Health checks
GET /health
GET /ready
GET /metrics
```

## Documentation

| Document | Description |
|----------|-------------|
| [Architecture](docs/architecture.md) | System design and component details |
| [Setup Guide](docs/setup.md) | Step-by-step installation |
| [Runbooks](docs/runbooks.md) | Operational procedures |
| [API Docs](docs/api.md) | API reference |
| [CI/CD](docs/cicd.md) | Pipeline explanation |
| [Troubleshooting](docs/troubleshooting.md) | Common issues and fixes |
| [Proposal](docs/proposal.md) | Architecture justification (FA/EN) |

## License

MIT
