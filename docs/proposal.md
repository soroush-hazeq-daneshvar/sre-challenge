# پروپوزال معماری - SRE Challenge

## Architecture Proposal / پیشنهاد معماری

---

## 1. خلاصه اجرایی

این سند توجیه انتخاب معماری پیشنهادی برای چالش SRE را ارائه می‌دهد. معماری بر اساس اصول **Cloud-Native**، **Infrastructure as Code**، **GitOps** و **Observability-Driven Development** طراحی شده است.

### اهداف کلیدی

- **قابلیت اطمینان (Reliability):** HA در تمام لایه‌ها
- **مقیاس‌پذیری (Scalability):** افزایش ظرفیت بدون تغییر معماری
- **قابلیت نگهداری (Maintainability):** IaC و GitOps برای تکرارپذیری
- **قابلیت مشاهده (Observability):** Metrics, Logs, Alerts
- **امنیت (Security):** Hardening در تمام لایه‌ها

---

## 2. چرا Kubernetes؟

### مشکل
نیاز به deploy و manage کردن چندین سرویس (API, Database, Monitoring, Logging) با قابلیت HA.

### راه‌حل: Kubernetes

| مزیت | توضیح |
|------|-------|
| **Self-healing** | Pod crash → automatic restart |
| **Horizontal Scaling** | `kubectl scale` یا HPA |
| **Service Discovery** | DNS-based, no hardcoded IPs |
| **Rolling Updates** | Zero-downtime deployments |
| **Resource Management** | CPU/Memory limits per container |
| **Ecosystem** | CloudNativePG, Prometheus Operator, Fluent Bit |

### چرا 3 Node؟

```
1 Control Plane + 2 Workers = حداقل برای HA
```

- Control Plane: مدیریت cluster (در production جدا deploy می‌شود)
- 2 Workers: anti-affinity برای pod distribution
- PostgreSQL replicas روی workerهای مختلف

### جایگزین‌های رد شده

| Alternative | Reason for Rejection |
|-------------|---------------------|
| Docker Compose | No HA, no self-healing, single host |
| Manual VM setup | Not reproducible, error-prone |
| Managed K8s (EKS/GKE) | Challenge requires self-managed cluster |

---

## 3. چرا Go برای GeoIP API؟

### مقایسه Go vs Python

| Criteria | Go | Python |
|----------|-----|--------|
| Performance | ~10x faster | Adequate for low traffic |
| Memory | ~10-20MB per pod | ~50-100MB per pod |
| Concurrency | Native goroutines | GIL limitation |
| Container size | ~15MB (distroless) | ~100MB+ |
| Startup time | <100ms | ~1-2s |
| Type safety | Compile-time checks | Runtime errors |

### تصمیم: Go

- **Performance:** GeoIP lookup latency-critical است
- **Resource efficiency:** در cluster 3-node، resource مهم است
- **Static binary:** deploy ساده، no runtime dependency
- **Prometheus ecosystem:** client_golang mature و widely used

---

## 4. چرا PostgreSQL با CloudNativePG؟

### مشکل
Cache persistence نیاز به database reliable دارد.

### راه‌حل: CloudNativePG Operator

```
Primary (RW) ──→ Replica 1 (RO)
              └──→ Replica 2 (RO)
```

| Feature | Benefit |
|---------|---------|
| Automatic failover | <30s downtime on primary failure |
| Streaming replication | Real-time data sync |
| PVC management | Data survives pod restarts |
| PodMonitor integration | Native Prometheus metrics |
| Backup/Restore | Built-in Barman support |

### چرا نه SQLite/Redis؟

| Alternative | Issue |
|-------------|-------|
| SQLite | Single file, no HA, not suitable for K8s |
| Redis | Cache-only, no complex queries, extra component |
| In-memory | Data loss on restart |

### Cache Strategy

```
Request → Check PostgreSQL → Hit? Return
                           → Miss? Call ipapi.co → Save → Return
```

- **Persistent cache:** survives pod restarts
- **Reduces external API calls:** cost and rate limit management
- **Queryable:** analytics on cached data

---

## 5. چرا Prometheus + Grafana + Alertmanager؟

### Observability Stack

```
Application Metrics → Prometheus → Grafana (visualize)
                                  → Alertmanager → Notifications
Cluster Metrics    → kube-state-metrics, node-exporter
Database Metrics   → CloudNativePG PodMonitor
```

### Application Metrics (Custom)

| Metric | Why Important |
|--------|---------------|
| `cache_hit_ratio` | Cache effectiveness |
| `external_api_calls` | Provider dependency monitoring |
| `request_latency_p95` | User experience |
| `error_rate` | Service health |

### Alerting Strategy

| Severity | Channel | Example |
|----------|---------|---------|
| Critical | Email + Slack + Telegram | API down, DB failover |
| Warning | Slack | High latency, cache miss rate |
| Info | Slack | Deployment completed |

### چرا نه VictoriaMetrics؟

VictoriaMetrics excellent است اما:
- Prometheus ecosystem بزرگ‌تر (more exporters, dashboards)
- kube-prometheus-stack همه چیز را one-click deploy می‌کند
- Challenge requirements صراحتاً Prometheus را mention کرده

---

## 6. چرا ELK Stack؟

### Logging Requirements

- Application errors
- All pod logs
- Searchable in Kibana

### Architecture

```
Pods → Fluent Bit (DaemonSet) → Elasticsearch → Kibana
```

| Component | Why |
|-----------|-----|
| **Fluent Bit** | Lightweight (vs Fluentd), K8s native, DaemonSet pattern |
| **Elasticsearch** | Full-text search, aggregations, scalable |
| **Kibana** | Visual log exploration, dashboards |

### Alternative: Loki

Loki lighter است اما:
- ELK industry standard برای log management
- Challenge صراحتاً ELK را mention کرده
- Better full-text search capabilities

---

## 7. چرا Terraform + Ansible؟

### Separation of Concerns

```
Terraform: WHAT to create (VMs, network, security groups)
Ansible:   HOW to configure (OS, K8s, CNI)
```

| Tool | Responsibility |
|------|---------------|
| Terraform | Declarative infrastructure, state management, idempotent |
| Ansible | Configuration management, idempotent, agentless |

### Why Not One Tool?

| Approach | Issue |
|----------|-------|
| Terraform only | Poor at configuration management (file editing, service start) |
| Ansible only | No state management, cloud API integration weaker |
| Pulumi | Less common, team familiarity with TF/Ansible |

### Reproducibility

```bash
terraform apply  # Same infrastructure every time
ansible-playbook # Same configuration every time
```

---

## 8. چرا GitLab CI + ArgoCD (GitOps)?

### CI vs CD Separation

```
CI (GitLab):  Code → Test → Build → Push Image
CD (ArgoCD):  Git Manifest → Sync → Cluster State
```

### GitOps Benefits

| Benefit | Description |
|---------|-------------|
| **Single source of truth** | Git repository |
| **Audit trail** | Every change is a commit |
| **Easy rollback** | `git revert` |
| **Drift detection** | ArgoCD alerts on manual changes |
| **Self-healing** | Auto-sync corrects drift |

### Pipeline Security

- Staging: automatic deploy
- Production: manual approval
- Image tagged with commit SHA (traceability)

---

## 9. Security Architecture

### Defense in Depth

```
Layer 1: Network (Security Groups, UFW)
Layer 2: OS (SSH hardening, fail2ban, auto-updates)
Layer 3: Kubernetes (RBAC, Network Policies, non-root containers)
Layer 4: Application (input validation, read-only filesystem)
Layer 5: Secrets (K8s Secrets, External Secrets Operator)
Layer 6: TLS (cert-manager, ingress TLS)
```

---

## 10. Cost Analysis

### Infrastructure (AWS)

| Resource | Type | Monthly Cost (est.) |
|----------|------|-------------------|
| Control Plane | t3.medium | ~$30 |
| Worker 1 | t3.large | ~$60 |
| Worker 2 | t3.large | ~$60 |
| EBS Volumes | 160GB gp3 | ~$13 |
| **Total** | | **~$163/month** |

### Optimization Options

- Spot instances for workers (-60%)
- Reserved instances for production (-40%)
- Right-sizing after load testing

---

## 11. Scalability Path

### Current (3 nodes)

```
2 GeoIP API pods, 3 PostgreSQL instances
~1000 req/sec capacity
```

### Growth Path

```
Phase 1: HPA for API pods (2→10)
Phase 2: Add worker nodes (3→5)
Phase 3: PostgreSQL read replicas for cache reads
Phase 4: CDN for cached responses
Phase 5: Multi-region deployment
```

---

## 12. Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| External API downtime | Cache misses fail | Persistent cache, fallback provider |
| Node failure | Reduced capacity | K8s rescheduling, anti-affinity |
| DB primary failure | Brief write outage | CloudNativePG auto-failover |
| Disk full (logs) | Log loss | Index lifecycle management |
| Certificate expiry | TLS errors | cert-manager auto-renewal |

---

## 13. نتیجه‌گیری

این معماری balance مناسبی بین **سادگی** (3-node cluster) و **production-readiness** (HA, monitoring, logging, GitOps) ارائه می‌دهد:

1. **Infrastructure as Code** → reproducible, auditable
2. **Kubernetes** → self-healing, scalable orchestration
3. **Go API** → performant, resource-efficient
4. **CloudNativePG** → reliable data layer with HA
5. **Prometheus/Grafana** → full observability
6. **ELK** → centralized logging
7. **GitOps** → safe, traceable deployments

تمام اجزا open-source هستند و vendor lock-in ندارند. معماری قابلیت scale از development تا production را بدون redesign دارد.

---

## Appendix: Technology Decision Matrix

| Component | Choice | Score (1-5) | Alternatives Considered |
|-----------|--------|-------------|------------------------|
| IaC | Terraform | 5 | Pulumi, CloudFormation |
| Config Mgmt | Ansible | 5 | Chef, Salt |
| Orchestration | Kubernetes | 5 | Docker Swarm, Nomad |
| Language | Go | 5 | Python, Rust |
| Database | CloudNativePG | 5 | Helm PostgreSQL, RDS |
| Monitoring | Prometheus | 5 | VictoriaMetrics, Datadog |
| Logging | ELK | 4 | Loki, Datadog |
| CI/CD | GitLab CI | 4 | GitHub Actions, Jenkins |
| GitOps | ArgoCD | 5 | Flux, Spinnaker |
| Ingress | NGINX | 5 | Traefik, HAProxy |
| CNI | Calico | 4 | Flannel, Cilium |

**Total Architecture Score: 47/50**
