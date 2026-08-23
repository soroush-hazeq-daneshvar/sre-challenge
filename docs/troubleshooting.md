# Troubleshooting Guide

## Common Issues

### 1. Pods stuck in Pending

**Cause:** Insufficient resources or PVC not bound

```bash
kubectl describe pod <pod-name> -n <namespace>
kubectl get pvc -A
kubectl get storageclass
```

**Fix:**
- Ensure StorageClass `gp3` exists or update manifests
- Check node resources: `kubectl top nodes`
- Remove taints if scheduling on control plane

---

### 2. GeoIP API returns "database not ready"

**Cause:** PostgreSQL cluster not ready or wrong connection string

```bash
kubectl get cluster -n postgres
kubectl get secret geoip-api-secrets -n geoip -o yaml
kubectl logs -l app=geoip-api -n geoip | grep -i "database\|postgres"
```

**Fix:**
```bash
# Wait for PostgreSQL
kubectl wait --for=condition=Ready cluster/geoip-postgres -n postgres --timeout=600s

# Verify connection string matches CNPG service name
# Should be: geoip-postgres-rw.postgres.svc:5432
kubectl rollout restart deployment/geoip-api -n geoip
```

---

### 3. Ingress returns 502 Bad Gateway

**Cause:** Backend pods not ready or service misconfigured

```bash
kubectl get ingress -n geoip
kubectl get endpoints geoip-api -n geoip
kubectl get pods -n geoip -o wide
```

**Fix:**
```bash
kubectl rollout status deployment/geoip-api -n geoip
kubectl logs -l app=geoip-api -n geoip --tail=20
```

---

### 4. Prometheus not scraping GeoIP metrics

**Cause:** ServiceMonitor labels don't match Prometheus selector

```bash
kubectl get servicemonitor -n geoip -o yaml
kubectl get prometheus -n monitoring -o yaml | grep serviceMonitorSelector
```

**Fix:**
- Ensure ServiceMonitor has label `release: prometheus`
- Verify pod annotations: `prometheus.io/scrape: "true"`

---

### 5. Fluent Bit not sending logs to Elasticsearch

```bash
kubectl logs -l app=fluent-bit -n logging --tail=50
kubectl get pods -n logging
curl http://elasticsearch.logging.svc:9200/_cluster/health
```

**Fix:**
```bash
# Restart Fluent Bit
kubectl rollout restart daemonset/fluent-bit -n logging

# Check Elasticsearch is running
kubectl wait --for=condition=Ready pod/elasticsearch-0 -n logging --timeout=300s
```

---

### 6. Ansible playbook fails on kubeadm init

**Cause:** Swap enabled, ports blocked, or previous failed init

```bash
ssh ubuntu@<cp-ip> sudo kubeadm reset -f
ssh ubuntu@<cp-ip> sudo swapoff -a
```

**Fix:**
```bash
ansible-playbook site.yml --tags init
```

---

### 7. External GeoIP API rate limited

**Symptoms:** 429 errors in logs, `GeoIPExternalAPIFailures` alert

```bash
kubectl logs -l app=geoip-api -n geoip | grep "429"
```

**Fix:**
- Cached requests are unaffected
- Wait for rate limit reset
- Pre-warm cache with common IPs
- Upgrade ipapi.co plan or switch provider

---

### 8. ArgoCD sync failed

```bash
argocd app get geoip-api
argocd app diff geoip-api
kubectl get events -n geoip --sort-by='.lastTimestamp'
```

**Fix:**
```bash
argocd app sync geoip-api --force
# Or fix the manifest and push to Git
```

---

## Useful Debug Commands

```bash
# Cluster overview
kubectl get nodes,pods,svc,ingress -A

# Resource usage
kubectl top nodes
kubectl top pods -A

# Recent events
kubectl get events -A --sort-by='.lastTimestamp' | tail -20

# DNS test
kubectl run -it --rm debug --image=busybox -- nslookup geoip-postgres-rw.postgres.svc

# Network test from app pod
kubectl exec -it deploy/geoip-api -n geoip -- wget -qO- http://geoip-postgres-rw.postgres.svc:5432

# Full pod description
kubectl describe pod -l app=geoip-api -n geoip
```

## Log Locations

| Component | How to Access |
|-----------|---------------|
| GeoIP API | `kubectl logs -l app=geoip-api -n geoip` |
| PostgreSQL | `kubectl logs geoip-postgres-1 -n postgres` |
| Prometheus | `kubectl logs -l app.kubernetes.io/name=prometheus -n monitoring` |
| Fluent Bit | `kubectl logs -l app=fluent-bit -n logging` |
| All logs | Kibana → Discover → `geoip-logs-*` index |
