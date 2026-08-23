# Runbooks - راهنمای عملیاتی

## RB-001: GeoIP API Down

**Severity:** Critical  
**Symptoms:** `/health` returns error, no response from API

### Diagnosis

```bash
kubectl get pods -n geoip
kubectl logs -l app=geoip-api -n geoip --tail=50
kubectl describe pod -l app=geoip-api -n geoip
```

### Resolution

```bash
# Restart deployment
kubectl rollout restart deployment/geoip-api -n geoip

# Check database connectivity
kubectl exec -it deploy/geoip-api -n geoip -- wget -qO- http://localhost:8080/ready

# Scale up if needed
kubectl scale deployment/geoip-api -n geoip --replicas=3
```

---

## RB-002: PostgreSQL Failover

**Severity:** Critical  
**Symptoms:** Database connection errors, replication lag alerts

### Diagnosis

```bash
kubectl get cluster geoip-postgres -n postgres
kubectl cnpg status geoip-postgres -n postgres
kubectl get pods -n postgres -l cnpg.io/cluster=geoip-postgres
```

### Resolution

CloudNativePG handles automatic failover. Manual promotion if needed:

```bash
kubectl cnpg promote geoip-postgres <new-primary-pod> -n postgres
kubectl rollout restart deployment/geoip-api -n geoip
```

---

## RB-003: High Cache Miss Rate

**Severity:** Warning  
**Symptoms:** Alert `GeoIPCacheMissRateHigh` firing

### Diagnosis

```bash
# Check metrics in Grafana
# geoip_cache_hits_total vs geoip_cache_misses_total

# Check database size
kubectl exec -it geoip-postgres-1 -n postgres -- psql -U geoip -c "SELECT count(*) FROM geoip_cache;"
```

### Resolution

- Normal after cache flush or new deployment
- If persistent, check if PostgreSQL data is being lost (PVC issues)
- Consider increasing cache TTL or pre-warming cache

---

## RB-004: External GeoIP Provider Failure

**Severity:** Critical  
**Symptoms:** Alert `GeoIPExternalAPIFailures`, cache misses returning errors

### Diagnosis

```bash
kubectl logs -l app=geoip-api -n geoip | grep "external api"
curl -v "https://ipapi.co/8.8.8.8/json/"
```

### Resolution

- Cached IPs will still work
- Wait for provider recovery
- Consider switching provider URL via env var:
  ```bash
  kubectl set env deployment/geoip-api GEOIP_PROVIDER_URL=https://alternative-provider.com -n geoip
  ```

---

## RB-005: Node Not Ready

**Severity:** Critical  
**Symptoms:** Alert `NodeNotReady`, pods evicted

### Diagnosis

```bash
kubectl get nodes
kubectl describe node <node-name>
ssh ubuntu@<node-ip> systemctl status kubelet
```

### Resolution

```bash
# Restart kubelet
ssh ubuntu@<node-ip> sudo systemctl restart kubelet

# If unrecoverable, drain and replace
kubectl drain <node-name> --ignore-daemonsets --delete-emptydir-data
# Replace VM via Terraform, re-run Ansible join
```

---

## RB-006: Disk Space Full (Elasticsearch)

**Severity:** Warning  
**Symptoms:** Fluent Bit errors, logs not appearing in Kibana

### Diagnosis

```bash
kubectl exec -it elasticsearch-0 -n logging -- curl -s localhost:9200/_cat/indices?v
kubectl get pvc -n logging
```

### Resolution

```bash
# Delete old indices
curl -X DELETE "elasticsearch.logging.svc:9200/geoip-logs-2024.01.*"

# Increase PVC size
kubectl patch pvc data-elasticsearch-0 -n logging -p '{"spec":{"resources":{"requests":{"storage":"50Gi"}}}}'
```

---

## RB-007: Deployment Rollback

**Severity:** High  
**Symptoms:** New deployment causing errors

### Resolution

```bash
# Rollback to previous version
kubectl rollout undo deployment/geoip-api -n geoip

# Rollback to specific revision
kubectl rollout history deployment/geoip-api -n geoip
kubectl rollout undo deployment/geoip-api -n geoip --to-revision=2

# Via ArgoCD
argocd app rollback geoip-api <revision>
```

---

## RB-008: Certificate Renewal Failure

**Severity:** Warning  
**Symptoms:** TLS errors on ingress

### Diagnosis

```bash
kubectl get certificates -A
kubectl describe certificate -n geoip
kubectl logs -n cert-manager -l app=cert-manager
```

### Resolution

```bash
kubectl delete certificaterequest -n geoip --all
# cert-manager will recreate automatically
```
