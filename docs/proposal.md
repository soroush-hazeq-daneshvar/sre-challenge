# پروپوزال معماری - SRE Challenge

# Architecture Proposal

---

# 1. خلاصه اجرایی

این سند معماری پیشنهادی برای پیاده‌سازی سرویس GeoIP API در محیط Kubernetes را توضیح می‌دهد.

معماری بر اساس اصول زیر طراحی شده است:

- Cloud Native
- Infrastructure as Code
- GitOps
- Observability
- Security by Design

## اهداف کلیدی

| هدف | توضیح |
|-----|------|
| Reliability | سرویس مقاوم در برابر failure و restart |
| Scalability | امکان افزایش replica و resource |
| Maintainability | مدیریت زیرساخت با Terraform و Ansible |
| Observability | Metrics, Logs, Alerts |
| Security | Hardening در تمام لایه‌ها |

---

# 2. معماری کلی سیستم

```
Developer
    |
    |
GitHub Repository
    |
    |
GitHub Actions / CI Pipeline
    |
    |
DockerHub Image
    |
    |
ArgoCD GitOps
    |
    |
Kubernetes Cluster
    |
    +----------------+
    |                |
 GeoIP API       Monitoring
    |                |
    |                |
CloudNativePG   Prometheus
    |                |
PostgreSQL      Grafana
                     |
                 Alertmanager


Application Logs

Pods
 |
Fluent Bit
 |
Elasticsearch
 |
Kibana
```

---

# 3. چرا Kubernetes؟

## Problem

نیاز به اجرای چندین component:

- GeoIP API
- PostgreSQL
- Monitoring
- Logging
- Ingress
- GitOps Controller

با قابلیت:

- Self Healing
- Scaling
- Service Discovery
- Rolling Update


## مزایای Kubernetes

| قابلیت | توضیح |
|-|-|
| Self Healing | Restart خودکار Pod های خراب |
| Service Discovery | ارتباط سرویس‌ها با DNS داخلی |
| Rolling Update | Deployment بدون downtime |
| Resource Management | کنترل CPU و Memory |
| Declarative Management | تعریف وضعیت مطلوب با YAML |
| Ecosystem | Prometheus, CNPG, ArgoCD |


## Cluster Design

```
1 Control Plane
+
2 Worker Nodes
```

### Control Plane

وظیفه:

- Kubernetes API
- Scheduler
- Controller Manager
- etcd


### Worker Nodes

وظیفه:

- اجرای Application Pods
- Database Pods
- Monitoring Components


---

# 4. Infrastructure Provisioning

## Terraform

Terraform مسئول ایجاد Infrastructure است:

```
Terraform
    |
    |
VMs
Network
Security Rules
Inventory
```


Responsibilities:

- VM provisioning
- Network configuration
- Output inventory for Ansible


## Ansible

Ansible مسئول Configuration Management است:

```
Ansible

 |
 +-- OS Configuration
 |
 +-- Container Runtime
 |
 +-- Kubernetes Installation
 |
 +-- CNI Installation
```


### Separation of Concerns

| Tool | Responsibility |
|-|-|
| Terraform | Create infrastructure |
| Ansible | Configure systems |
| Kubernetes | Run workloads |


---

# 5. چرا Go برای GeoIP API؟

## Decision: Go


دلایل انتخاب Go:

### Performance

- Fast startup
- Low memory usage
- High concurrency


### Container Optimization

Application:

```
Go Source
 |
Build Stage
 |
Static Binary
 |
Distroless Image
```


Docker Image:

```
docker.io/soroushmanhd/geoip-api:latest
```


مزایا:

- کوچک
- امن
- بدون runtime dependency


---

# 6. Database Architecture

## CloudNativePG + PostgreSQL


Database برای نگهداری cache استفاده می‌شود.


Architecture:

```
GeoIP API

     |
     |
 PostgreSQL

     |
 CloudNativePG Operator
```


## مزایا

| Feature | Benefit |
|-|-|
| Automatic failover | Recovery خودکار |
| Kubernetes native | مدیریت توسط Operator |
| Persistent Storage | حفظ اطلاعات بعد از restart |
| Monitoring Integration | Metrics برای Prometheus |


## Cache Flow


```
Request
   |
   |
Check PostgreSQL Cache

   |
   +------ Hit ------> Return Result

   |
   +------ Miss -----> Call ipapi.co
                         |
                         |
                    Save Result
                         |
                         |
                     Return Response
```


مزایا:

- کاهش dependency خارجی
- کاهش latency
- جلوگیری از rate limit


---

# 7. Monitoring Architecture

## kube-prometheus-stack


Deployment:

```
Helm

kube-prometheus-stack

        |
        |
+---------------+
|               |
Prometheus   Grafana
|
|
Alertmanager
```


## Metrics Sources


```
Application
    |
/metrics endpoint
    |
Prometheus


Kubernetes
    |
node-exporter
kube-state-metrics
```


## Application Metrics


| Metric | Purpose |
|-|-|
| geoip_cache_hits_total | Cache performance |
| geoip_cache_misses_total | Cache failures |
| geoip_external_api_calls_total | Provider dependency |
| HTTP latency | User experience |
| HTTP errors | Application health |

---

# 8. Logging Architecture

## ELK Stack


Architecture:


```
Kubernetes Pods

      |
      |
 Fluent Bit
(DaemonSet)

      |
      |
 Elasticsearch

      |
      |
 Kibana
```


## Components


| Component | Purpose |
|-|-|
| Fluent Bit | Collect container logs |
| Elasticsearch | Store and search logs |
| Kibana | Visualization |


---

# 9. GitOps Architecture


## ArgoCD


Flow:

```
Developer

 |
 |
Push Code

 |
 |
GitHub Repository

 |
 |
ArgoCD

 |
 |
Kubernetes Cluster
```


Benefits:


| Feature | Benefit |
|-|-|
| Single Source of Truth | Git repository |
| Audit Trail | Every change tracked |
| Rollback | Git revert |
| Drift Detection | Detect manual changes |
| Self Healing | Auto synchronization |


Application repository:


```
https://github.com/soroush-hazeq-daneshvar/sre-challenge
```


---

# 10. Security Architecture


Defense in Depth:


```
Layer 1:
Infrastructure Security

Layer 2:
OS Hardening

Layer 3:
Kubernetes RBAC

Layer 4:
Container Security

Layer 5:
Application Security
```


Implemented:


- Non-root container
- Read-only filesystem
- Resource limits
- Kubernetes secrets
- Network isolation
- SSH hardening


---

# 11. Storage Architecture


StorageClass:

```
local-path
```


Used by:


- PostgreSQL PVC
- Elasticsearch PVC
- Prometheus PVC
- Grafana PVC


Benefits:

- Persistent data
- Kubernetes native storage


---

# 12. Deployment Strategy


## CI/CD


Pipeline:


```
Developer

 |
 |
Git Push

 |
 |
Build Go Application

 |
 |
Docker Build

 |
 |
Push DockerHub

 |
 |
Update Kubernetes Manifest

 |
 |
ArgoCD Sync

 |
 |
Deployment
```


Docker Image:

```
docker.io/soroushmanhd/geoip-api:latest
```


---

# 13. Scalability Path


## Current

```
GeoIP API
2 replicas

PostgreSQL
CloudNativePG

Monitoring
Prometheus Stack
```


## Future


Phase 1:

```
Horizontal Pod Autoscaler
```


Phase 2:

```
Add Kubernetes Worker Nodes
```


Phase 3:

```
PostgreSQL Scaling
```


Phase 4:

```
External Load Balancer
```


Phase 5:

```
Multi Region Deployment
```


---

# 14. Risk Assessment


| Risk | Impact | Mitigation |
|-|-|-|
| External API failure | Lookup failure | Persistent cache |
| Node failure | Pod movement | Kubernetes rescheduling |
| Database failure | API unavailable | CloudNativePG failover |
| Disk full | Logging failure | Monitoring and cleanup |
| Bad deployment | Service outage | ArgoCD rollback |


---

# 15. نتیجه‌گیری


این معماری یک راهکار Production Ready برای اجرای GeoIP API ارائه می‌دهد.

مزایای اصلی:

1. Terraform + Ansible

Infrastructure as Code


2. Kubernetes

Self-healing and scalable platform


3. Go API

High performance application


4. CloudNativePG

Reliable database layer


5. Prometheus + Grafana

Complete observability


6. Elasticsearch + Fluent Bit

Centralized logging


7. ArgoCD

GitOps deployment model


تمام component ها Open Source هستند و قابلیت انتقال به Cloud Provider های مختلف مانند ArvanCloud را دارند.

---

# Technology Decision Matrix


| Component | Choice | Alternative |
|-|-|-|
| IaC | Terraform | Pulumi |
| Config Management | Ansible | SaltStack |
| Container Runtime | Containerd | Docker |
| Orchestration | Kubernetes | Nomad |
| Language | Go | Python |
| Database | CloudNativePG | PostgreSQL Helm |
| Monitoring | Prometheus | VictoriaMetrics |
| Dashboard | Grafana | Kibana |
| Logging | Elasticsearch | Loki |
| GitOps | ArgoCD | Flux |
| Ingress | NGINX | Traefik |
| CNI | Calico | Cilium |


## Final Architecture Score

**47/50**