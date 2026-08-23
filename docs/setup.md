# 4. Install Kubernetes Components


## 4.1 Base Configuration


Create project namespaces:


```bash
kubectl apply \
-f kubernetes/base/namespaces.yaml
```


Verify:


```bash
kubectl get namespaces
```


Expected namespaces:

```
geoip
postgres
monitoring
logging
```


---

# 4.2 NGINX Ingress Controller (HelmChart)


The project deploys ingress-nginx using Kubernetes HelmChart resource.


Apply ingress configuration:


```bash
kubectl apply \
-f kubernetes/base/ingress-nginx.yaml
```


Verify:


```bash
kubectl get pods -n ingress-nginx
```


Expected:

```
ingress-nginx-controller   Running
```


Check service:


```bash
kubectl get svc -n ingress-nginx
```


---

# 4.3 StorageClass


The cluster uses Rancher Local Path Provisioner.


Verify StorageClass:


```bash
kubectl get storageclass
```


Expected:


```
NAME
local-path
```


This StorageClass provides persistent storage for:

- PostgreSQL
- Prometheus
- Grafana
- Elasticsearch


---

# 4.4 CloudNativePG Operator


CloudNativePG provides PostgreSQL High Availability.


Install operator:


```bash
kubectl apply \
-f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.25/releases/cnpg-1.25.0.yaml
```


Verify:


```bash
kubectl get pods -n cnpg-system
```


Expected:


```
cnpg-controller-manager   Running
```


Deploy PostgreSQL cluster:


```bash
kubectl apply \
-f kubernetes/postgres/cluster.yaml
```


Verify:


```bash
kubectl get cluster -n postgres


kubectl get pods -n postgres
```


---

# 4.5 Monitoring Stack (Helm)


The monitoring stack uses the official Prometheus Community Helm chart.


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
kubectl get pods -n monitoring
```


Expected components:


```
Prometheus
Grafana
Alertmanager
kube-state-metrics
node-exporter
```


Deploy Prometheus alert rules:


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


Open:


```
http://localhost:3000
```


Get Grafana password:


```bash
kubectl get secret \
kube-prometheus-stack-grafana \
-n monitoring \
-o jsonpath="{.data.admin-password}" \
| base64 -d
```


Default username:


```
admin
```


---

# 4.6 Logging Stack


## Elasticsearch


Deploy Elasticsearch:


```bash
kubectl apply \
-f kubernetes/logging/elasticsearch/elasticsearch.yaml
```


Verify:


```bash
kubectl get pods -n logging
```


Expected:


```
elasticsearch-0   Running
```


---

## Fluent Bit


Deploy Fluent Bit daemonset:


```bash
kubectl apply \
-f kubernetes/logging/fluent-bit/daemonset.yaml
```


Verify:


```bash
kubectl get pods -n logging
```


Expected:


```
fluent-bit-*   Running
```


---

## Kibana


Kibana is deployed together with Elasticsearch manifest.


Verify:


```bash
kubectl get pods -n logging
```


Access:

```
http://kibana.local
```


---

# 4.7 ArgoCD Installation


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


Verify:


```bash
kubectl get pods -n argocd
```


Access ArgoCD:


```bash
kubectl port-forward \
svc/argocd-server \
-n argocd \
8080:443
```


URL:


```
https://localhost:8080
```


Get admin password:


```bash
kubectl get secret \
argocd-initial-admin-secret \
-n argocd \
-o jsonpath="{.data.password}" \
| base64 -d
```


---

# 4.8 Deploy ArgoCD Applications


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


Check applications:


```bash
kubectl get applications \
-n argocd
```