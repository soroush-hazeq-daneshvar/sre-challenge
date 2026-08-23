# Runbooks - راهنمای عملیاتی

## RB-001: GeoIP API Down

**Severity:** Critical

**Symptoms:**
- API is not responding
- `/health` endpoint returns error
- Pods are restarting or unavailable

## Diagnosis

```bash
# Check pod status
kubectl get pods -n geoip

# Check deployment status
kubectl get deployment geoip-api -n geoip

# Check application logs
kubectl logs -l app=geoip-api -n geoip --tail=100

# Describe failed pod
kubectl describe pod -l app=geoip-api -n geoip

# Check service connectivity
kubectl get svc -n geoip
```

## Resolution

```bash
# Restart application deployment
kubectl rollout restart deployment/geoip-api -n geoip

# Watch rollout status
kubectl rollout status deployment/geoip-api -n geoip

# Test health endpoint internally
kubectl exec -it deploy/geoip-api -n geoip -- \
wget -qO- http://localhost:8080/health

# Test readiness endpoint
kubectl exec -it deploy/geoip-api -n geoip -- \
wget -qO- http://localhost:8080/ready

# Scale application if required
kubectl scale deployment/geoip-api -n geoip --replicas=3
```

---

# RB-002: PostgreSQL Failure / Failover

**Severity:** Critical

**Symptoms:**

- GeoIP API cannot connect to database
- `/ready` endpoint returns database error
- PostgreSQL pods unavailable

## Diagnosis

```bash
# Check CloudNativePG cluster
kubectl get cluster -n postgres

# Check PostgreSQL pods
kubectl get pods -n postgres

# Check cluster status
kubectl cnpg status geoip-postgres -n postgres

# Check events
kubectl describe cluster geoip-postgres -n postgres
```

## Resolution

CloudNativePG handles automatic failover.

Check cluster recovery:

```bash
kubectl get pods -n postgres

kubectl cnpg status geoip-postgres -n postgres
```

Restart application after database recovery:

```bash
kubectl rollout restart deployment/geoip-api -n geoip
```

Manual promotion if required:

```bash
kubectl cnpg promote geoip-postgres <new-primary-pod> -n postgres
```

---

# RB-003: High Cache Miss Rate

**Severity:** Warning

**Symptoms:**

- Alert `GeoIPCacheMissRateHigh`
- Increased external API calls
- Higher response latency

## Diagnosis

Check Prometheus/Grafana metrics:

```
geoip_cache_hits_total
geoip_cache_misses_total
geoip_external_api_calls_total
```

Check database cache records:

```bash
kubectl exec -it geoip-postgres-1 -n postgres -- \
psql -U geoip -c "SELECT count(*) FROM geoip_cache;"
```

## Resolution

Possible actions:

- Verify PostgreSQL storage health
- Verify cache table availability
- Check application logs
- Increase cache TTL if required
- Pre-warm cache with frequently requested IP addresses

---

# RB-004: External GeoIP Provider Failure

**Severity:** Critical

**Symptoms:**

- External lookup failures
- Alert `GeoIPExternalAPIFailures`
- New IP lookups return errors

## Diagnosis

Check application logs:

```bash
kubectl logs -l app=geoip-api -n geoip | grep "external api"
```

Test external provider:

```bash
curl -v https://ipapi.co/8.8.8.8/json/
```

Check metrics:

```
geoip_external_api_calls_total
geoip_external_api_duration_seconds
```

## Resolution

Cached responses should continue working.

If provider is unavailable:

```bash
# Change provider endpoint
kubectl set env deployment/geoip-api \
GEOIP_PROVIDER_URL=https://alternative-provider.com \
-n geoip
```

Restart deployment:

```bash
kubectl rollout restart deployment/geoip-api -n geoip
```

---

# RB-005: Kubernetes Node Not Ready

**Severity:** Critical

**Symptoms:**

- NodeNotReady alert
- Pods moved or evicted
- Application downtime

## Diagnosis

```bash
# Check nodes
kubectl get nodes

# Check node details
kubectl describe node <node-name>

# Check kubelet service
ssh ubuntu@<node-ip> systemctl status kubelet
```

## Resolution

Restart kubelet:

```bash
ssh ubuntu@<node-ip> sudo systemctl restart kubelet
```

Drain unhealthy node:

```bash
kubectl drain <node-name> \
--ignore-daemonsets \
--delete-emptydir-data
```

Infrastructure recovery:

```bash
# Recreate VM using Terraform
cd terraform

terraform plan
terraform apply
```

Reconfigure Kubernetes:

```bash
cd ansible

ansible-playbook site.yml
```

---

# RB-006: Elasticsearch Disk Full

**Severity:** Warning

**Symptoms:**

- Fluent Bit cannot send logs
- Kibana shows missing logs
- Elasticsearch rejects writes

## Diagnosis

Check Elasticsearch status:

```bash
kubectl get pods -n logging

kubectl exec -it elasticsearch-0 -n logging -- \
curl -s localhost:9200/_cluster/health?pretty
```

Check indices:

```bash
kubectl exec -it elasticsearch-0 -n logging -- \
curl -s localhost:9200/_cat/indices?v
```

Check storage:

```bash
kubectl get pvc -n logging
```

## Resolution

Delete old indexes if required:

```bash
kubectl exec -it elasticsearch-0 -n logging -- \
curl -X DELETE \
localhost:9200/geoip-logs-old-index
```

Increase PVC size:

```bash
kubectl patch pvc data-elasticsearch-0 \
-n logging \
-p '{"spec":{"resources":{"requests":{"storage":"50Gi"}}}}'
```

Restart logging components:

```bash
kubectl rollout restart deployment kibana -n logging
kubectl rollout restart daemonset fluent-bit -n logging
```

---

# RB-007: GeoIP API Deployment Rollback

**Severity:** High

**Symptoms:**

- New Docker image causes failures
- Health checks failing
- Application errors after deployment

## Diagnosis

```bash
kubectl rollout history deployment/geoip-api -n geoip
```

## Resolution

Rollback deployment:

```bash
kubectl rollout undo deployment/geoip-api -n geoip
```

Rollback to specific revision:

```bash
kubectl rollout undo deployment/geoip-api \
-n geoip \
--to-revision=2
```

GitOps rollback:

```bash
argocd app history geoip-api

argocd app rollback geoip-api <revision>
```

Recommended GitOps method:

```bash
git revert <commit-id>

git push origin main
```

ArgoCD will automatically synchronize the previous state.

---

# RB-008: ArgoCD Sync Failure

**Severity:** High

**Symptoms:**

- Application status is OutOfSync
- Kubernetes resources are not updated

## Diagnosis

```bash
argocd app list

argocd app get geoip-api

kubectl get applications -n argocd
```

## Resolution

Manual sync:

```bash
argocd app sync geoip-api
```

Restart ArgoCD components if required:

```bash
kubectl get pods -n argocd

kubectl rollout restart deployment argocd-server -n argocd
```

---

# RB-009: Docker Image Pull Failure

**Severity:** High

**Symptoms:**

- Pod status:
  - ImagePullBackOff
  - ErrImagePull

## Diagnosis

```bash
kubectl describe pod -n geoip -l app=geoip-api
```

Check image:

```bash
kubectl get deployment geoip-api \
-n geoip \
-o yaml | grep image
```

Current image:

```
docker.io/soroushmanhd/geoip-api:latest
```

## Resolution

Test image availability:

```bash
docker pull soroushmanhd/geoip-api:latest
```

Restart deployment:

```bash
kubectl rollout restart deployment/geoip-api -n geoip
```

---

# RB-010: Monitoring Stack Failure

**Severity:** Warning

**Symptoms:**

- Grafana unavailable
- Prometheus unavailable
- Missing metrics

## Diagnosis

```bash
kubectl get pods -n monitoring

kubectl get servicemonitor -A

kubectl get prometheus -n monitoring
```

Check Prometheus:

```bash
kubectl logs \
prometheus-kube-prometheus-stack-prometheus-0 \
-n monitoring
```

## Resolution

Restart monitoring components:

```bash
kubectl rollout restart deployment kube-prometheus-stack-grafana \
-n monitoring
```

Check Helm release:

```bash
helm list -n monitoring

helm status kube-prometheus-stack -n monitoring
```

---

# RB-011: Infrastructure Recovery

**Severity:** Critical

## Terraform Recovery

```bash
cd terraform

terraform init

terraform plan

terraform apply
```

## Ansible Recovery

```bash
cd ansible

ansible all -m ping

ansible-playbook site.yml
```

## Kubernetes Validation

```bash
kubectl get nodes

kubectl get pods -A

kubectl get events -A --sort-by=.lastTimestamp
```