# معماری سیستم - SRE Challenge

## نمای کلی

این پروژه یک پلتفرم GeoIP کامل و production-ready است که تمام لایه‌های زیرساخت تا اپلیکیشن را پوشش می‌دهد.

## دیاگرام معماری

```
┌─────────────────────────────────────────────────────────────────┐
│                        External Users                           │
│                    GET /country?ip=8.8.8.8                      │
└──────────────────────────┬──────────────────────────────────────┘
                           │
                    ┌──────▼──────┐
                    │ NGINX Ingress│
                    └──────┬──────┘
                           │
              ┌────────────▼────────────┐
              │     GeoIP API (Go)      │
              │  ┌──────────────────┐   │
              │  │ Cache Check      │   │
              │  │ ↓ miss → ipapi.co│   │
              │  │ ↓ hit → return   │   │
              │  └──────────────────┘   │
              │  /metrics (Prometheus)    │
              └────────────┬────────────┘
                           │
              ┌────────────▼────────────┐
              │  PostgreSQL (CNPG)      │
              │  Primary + 2 Replicas   │
              │  SSD Persistent Volumes │
              └─────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Observability Stack                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────────┐     │
│  │Prometheus│→ │ Grafana  │  │Alertmgr  │→ │Email/Slack/  │     │
│  │          │  │Dashboard │  │          │  │Telegram      │     │
│  └──────────┘  └──────────┘  └──────────┘  └──────────────┘     │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐                      │
│  │Fluent Bit│→ │Elastic-  │→ │ Kibana   │                      │
│  │(DaemonSet│  │ search   │  │          │                      │
│  └──────────┘  └──────────┘  └──────────┘                      │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    Infrastructure Layer                         │
│  Terraform → Ansible → K8s (1 CP + 2 Workers) → Calico CNI    │
└─────────────────────────────────────────────────────────────────┘

┌─────────────────────────────────────────────────────────────────┐
│                    CI/CD & GitOps                               │
│  GitLab CI: Lint → Test → Build → Push → Deploy                │
│  ArgoCD: Declarative sync from Git repository                   │
└─────────────────────────────────────────────────────────────────┘
```

## اجزای اصلی

### 1. Infrastructure (Terraform + Ansible)

- **Terraform**: ایجاد 3 VM (1 Control Plane + 2 Worker) روی AWS
- **Ansible**: OS Hardening، نصب containerd، Kubernetes 1.34، Calico CNI

### 2. Kubernetes Cluster

| Component | Role |
|-----------|------|
| Control Plane | etcd, api-server, scheduler, controller-manager |
| Worker 1 & 2 | kubelet, kube-proxy, containerd |
| CoreDNS | Cluster DNS |
| Metrics Server | Resource metrics |
| cert-manager | TLS certificates |
| External Secrets | Secret management |

### 3. GeoIP API (Go)

**Flow:**
1. دریافت IP از query parameter
2. جستجو در PostgreSQL cache
3. در صورت miss، فراخوانی ipapi.co
4. ذخیره نتیجه در cache
5. بازگشت JSON response

**Metrics:**
- `geoip_http_requests_total` - تعداد درخواست‌ها
- `geoip_http_request_duration_seconds` - latency
- `geoip_cache_hits_total` / `geoip_cache_misses_total` - نسبت cache
- `geoip_external_api_calls_total` - فراخوانی‌های خارجی

### 4. PostgreSQL HA (CloudNativePG)

- 1 Primary + 2 Replicas
- Automatic failover
- Persistent Volumes (SSD)
- PodMonitor برای Prometheus

### 5. Monitoring Stack

- **Prometheus**: جمع‌آوری metrics از app و cluster
- **Grafana**: Dashboard اختصاصی GeoIP API
- **Alertmanager**: ارسال alert به Email, Slack, Telegram

### 6. Logging Stack (ELK)

- **Fluent Bit**: DaemonSet برای جمع‌آوری log تمام podها
- **Elasticsearch**: ذخیره و index کردن logs
- **Kibana**: جستجو و visualization

## Network Flow

```
Client → Ingress (80/443) → GeoIP Service → Pod
                                              ↓
                                         PostgreSQL RW Service
                                              ↓
                                         Primary Instance
```

## Security Considerations

- OS hardening (SSH, UFW, fail2ban)
- Non-root containers
- Read-only root filesystem
- Secrets via Kubernetes Secrets / External Secrets
- Network policies (Calico)
- TLS via cert-manager
