# راهنمای نصب و راه‌اندازی

# Setup Guide - GeoIP SRE Platform

This document describes the complete deployment process of the GeoIP SRE Platform.

The platform consists of:

- Terraform infrastructure provisioning
- Ansible Kubernetes cluster configuration
- Kubernetes workloads
- NGINX Ingress
- GeoIP API
- CloudNativePG PostgreSQL
- Prometheus
- Grafana
- Alertmanager
- Elasticsearch
- Fluent Bit
- Kibana
- ArgoCD GitOps

---

# 1. Architecture

```text
                         +----------------------+
                         |      Developer       |
                         +----------+-----------+
                                    |
                                    v
                              GitLab CI/CD
                                    |
                                    v
                              Docker Hub
                                    |
                                    v
                         soroushmanhd/geoip-api
                                    |
                                    v
+----------------------------------------------------------------+
|                         Kubernetes Cluster                     |
|                                                                |
|  +-------------------+                                         |
|  |   NGINX Ingress   |                                         |
|  |   NodePort :31406 |                                         |
|  +---------+---------+                                         |
|            |                                                   |
|      +-----+------+----------------------+                     |
|      |            |                      |                     |
|      v            v                      v                     |
| geoip.local  grafana.local        prometheus.local             |
|      |            |                      |                     |
|      v            v                      v                     |
|  GeoIP API      Grafana             Prometheus                 |
|                                                                |
|      +------------------- Logging --------------------------+   |
|      |                                                     |   |
|      |  Kubernetes Pods                                    |   |
|      |       |                                             |   |
|      |       v                                             |   |
|      |   Fluent Bit (DaemonSet)                            |   |
|      |       |                                             |   |
|      |       v                                             |   |
|      |   Elasticsearch                                     |   |
|      |       |                                             |   |
|      |       v                                             |   |
|      |     Kibana                                          |   |
|      |                                                     |   |
|      +-----------------------------------------------------+   |
|                                                                |
|      +-------------------+                                     |
|      | CloudNativePG     |                                     |
|      | PostgreSQL HA     |                                     |
|      +-------------------+                                     |
|                                                                |
|      +-------------------+                                     |
|      | ArgoCD            |                                     |
|      | GitOps            |                                     |
|      +-------------------+                                     |
+----------------------------------------------------------------+
2. Prerequisites

Required tools:

Tool	Version	Purpose
Terraform	>= 1.5	Infrastructure provisioning
Ansible	>= 2.14	Configuration management
kubectl	>= 1.34	Kubernetes management
Docker	>= 24	Container build
Go	>= 1.22	Application development
Helm	>= 3.12	Kubernetes package management
3. Build Application Image

The application image is hosted on Docker Hub.

Docker image:

soroushmanhd/geoip-api:latest

Pull the image manually:

docker pull soroushmanhd/geoip-api:latest

Build locally:

cd app/geoip-api

docker build \
  -t soroushmanhd/geoip-api:latest .

Login to Docker Hub:

docker login

Push the image:

docker push soroushmanhd/geoip-api:latest
4. Infrastructure Provisioning with Terraform

Terraform creates the required infrastructure.

Components:

Virtual Machines
Network
Security Groups
Storage
Kubernetes nodes

Go to Terraform:

cd terraform

Create variables:

cp terraform.tfvars.example terraform.tfvars

Edit variables:

vim terraform.tfvars

Initialize Terraform:

terraform init

Validate:

terraform validate

Review changes:

terraform plan

Apply:

terraform apply

Generate Ansible inventory:

terraform output -raw ansible_inventory \
  > ../ansible/inventory/hosts.yml

Expected infrastructure:

1 Control Plane Node
2 Worker Nodes
Network
Security Groups
Storage
5. Kubernetes Cluster Installation with Ansible

Move to Ansible:

cd ../ansible

Install required collections:

ansible-galaxy collection install \
  -r requirements.yml

Test connectivity:

ansible all -m ping

Install Kubernetes:

ansible-playbook site.yml

Verify the cluster:

ssh ubuntu@<control-plane-ip>

Check nodes:

kubectl get nodes -o wide

Expected:

NAME                              STATUS   ROLES
stg-hazegh-vm1.dc.snappcloud.io   Ready    control-plane
stg-hazegh-vm2.dc.snappcloud.io   Ready    <none>
stg-hazegh-vm3.dc.snappcloud.io   Ready    <none>
6. Kubernetes Components
6.1 Base Namespaces

Create namespaces:

kubectl apply \
  -f kubernetes/base/namespaces.yaml

Verify:

kubectl get namespaces

Expected namespaces include:

geoip
postgres
monitoring
logging
argocd
ingress-nginx
7. NGINX Ingress Controller

The project uses ingress-nginx.

Add the Helm repository:

helm repo add ingress-nginx \
  https://kubernetes.github.io/ingress-nginx

Update Helm repositories:

helm repo update

Install ingress-nginx:

helm upgrade --install ingress-nginx \
  ingress-nginx/ingress-nginx \
  --namespace ingress-nginx \
  --create-namespace \
  -f kubernetes/base/ingress-nginx.yaml

Verify:

kubectl get pods \
  -n ingress-nginx

Verify the service:

kubectl get svc \
  -n ingress-nginx \
  ingress-nginx-controller

The current setup exposes HTTP through NodePort:

HTTP  -> 31406
HTTPS -> 31469

The Kubernetes control-plane node is currently:

10.160.1.22

Therefore the HTTP ingress endpoint is:

http://10.160.1.22:31406

Ingress traffic is routed using the HTTP Host header.

8. StorageClass

The cluster uses Rancher Local Path Provisioner.

Verify:

kubectl get storageclass

Expected:

NAME
local-path

The StorageClass is used by:

Elasticsearch
PostgreSQL
Prometheus
Grafana
9. CloudNativePG PostgreSQL

Install CloudNativePG:

kubectl apply \
  -f https://raw.githubusercontent.com/cloudnative-pg/cloudnative-pg/release-1.25/releases/cnpg-1.25.0.yaml

Verify:

kubectl get pods \
  -n cnpg-system

Deploy PostgreSQL:

kubectl apply \
  -f kubernetes/postgres/cluster.yaml

Verify:

kubectl get cluster \
  -n postgres

Check PostgreSQL pods:

kubectl get pods \
  -n postgres
10. Monitoring Stack

Monitoring consists of:

Prometheus
Grafana
Alertmanager
kube-state-metrics
node-exporter

Add the Helm repository:

helm repo add prometheus-community \
  https://prometheus-community.github.io/helm-charts

Update:

helm repo update

Install kube-prometheus-stack:

helm upgrade --install kube-prometheus-stack \
  prometheus-community/kube-prometheus-stack \
  -n monitoring \
  --create-namespace \
  -f kubernetes/monitoring/prometheus/helm-values.yaml

Verify:

kubectl get pods \
  -n monitoring

Deploy Prometheus alerts:

kubectl apply \
  -f kubernetes/monitoring/prometheus/alerts.yaml

Deploy GeoIP Grafana dashboard:

kubectl apply \
  -f kubernetes/monitoring/grafana/dashboard-geoip.yaml
11. Grafana

Grafana is exposed through the NGINX Ingress.

Verify:

kubectl get ingress \
  -n monitoring

Add the hostname to /etc/hosts:

sudo sh -c 'echo "10.160.1.22 grafana.local" >> /etc/hosts'

Access:

http://grafana.local:31406

Alternatively, test through curl:

curl --noproxy '*' \
  -H 'Host: grafana.local' \
  http://10.160.1.22:31406

Get Grafana password:

kubectl get secret \
  kube-prometheus-stack-grafana \
  -n monitoring \
  -o jsonpath="{.data.admin-password}" \
  | base64 -d

Username:

admin
12. Logging Stack

The logging architecture is:

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

Fluent Bit runs as a DaemonSet.

This means one Fluent Bit pod runs on each Kubernetes node.

Verify Fluent Bit:

kubectl get pods \
  -n logging \
  -l app=fluent-bit

Expected:

fluent-bit-xxxxx   1/1   Running
fluent-bit-xxxxx   1/1   Running
fluent-bit-xxxxx   1/1   Running
13. Elasticsearch

Elasticsearch runs as a single-node StatefulSet.

Verify:

kubectl get pods \
  -n logging

Expected:

elasticsearch-0   1/1   Running

Verify Elasticsearch service:

kubectl get svc \
  -n logging \
  elasticsearch

Expected:

elasticsearch   ClusterIP   <cluster-ip>   <none>   9200/TCP

Elasticsearch is available internally at:

http://elasticsearch.logging.svc:9200

Test from Kubernetes:

kubectl run curl-elasticsearch \
  --rm -it \
  --restart=Never \
  --image=curlimages/curl \
  -n logging \
  -- \
  curl http://elasticsearch:9200

Expected response:

{
  "name": "elasticsearch-0",
  "cluster_name": "docker-cluster",
  "version": {
    "number": "8.12.0"
  },
  "tagline": "You Know, for Search"
}

Check cluster health:

kubectl run curl-elasticsearch \
  --rm -it \
  --restart=Never \
  --image=curlimages/curl \
  -n logging \
  -- \
  curl http://elasticsearch:9200/_cluster/health
14. Elasticsearch Ingress

Elasticsearch is exposed through:

elasticsearch.local

The ingress is:

Client
  |
  v
elasticsearch.local
  |
  v
NGINX Ingress
  |
  v
elasticsearch:9200

Verify:

kubectl get ingress \
  -n logging

Expected:

NAME             CLASS   HOSTS
elasticsearch    nginx   elasticsearch.local
kibana           nginx   kibana.local

Add the hostname:

sudo sh -c 'echo "10.160.1.22 elasticsearch.local" >> /etc/hosts'

Test:

curl --noproxy '*' \
  -H 'Host: elasticsearch.local' \
  http://10.160.1.22:31406/

You should receive the Elasticsearch JSON response.

Check cluster health:

curl --noproxy '*' \
  -H 'Host: elasticsearch.local' \
  http://10.160.1.22:31406/_cluster/health
15. Fluent Bit

Fluent Bit collects container logs from Kubernetes nodes and sends them to Elasticsearch.

Verify the DaemonSet:

kubectl get daemonset \
  -n logging

Check Fluent Bit pods:

kubectl get pods \
  -n logging \
  -l app=fluent-bit

Check Fluent Bit logs:

kubectl logs \
  -n logging \
  -l app=fluent-bit \
  --tail=100

Look for successful Elasticsearch output messages.

16. Verify Logs in Elasticsearch

List Elasticsearch indices:

curl --noproxy '*' \
  -H 'Host: elasticsearch.local' \
  http://10.160.1.22:31406/_cat/indices?v

Search all documents:

curl --noproxy '*' \
  -H 'Host: elasticsearch.local' \
  http://10.160.1.22:31406/_search?pretty

Check document count:

curl --noproxy '*' \
  -H 'Host: elasticsearch.local' \
  http://10.160.1.22:31406/_count?pretty

If Fluent Bit is configured correctly, Elasticsearch should contain Kubernetes container logs.

17. Kibana

Kibana provides the web interface for Elasticsearch.

Kibana connects internally to Elasticsearch using:

http://elasticsearch.logging.svc:9200

Verify the Kibana pod:

kubectl get pods \
  -n logging \
  -l app=kibana

Verify the service:

kubectl get svc \
  -n logging \
  kibana

Expected:

kibana   ClusterIP   <cluster-ip>   <none>   5601/TCP

Test Kibana internally:

kubectl run curl-kibana \
  --rm -it \
  --restart=Never \
  --image=curlimages/curl \
  -n logging \
  -- \
  curl -I http://kibana:5601/

A successful response will normally be:

HTTP/1.1 302 Found
location: /spaces/enter

This confirms that Kibana is running correctly.

18. Kibana Ingress

Kibana is exposed separately from Elasticsearch.

Hostname:

kibana.local

Architecture:

Browser
   |
   v
kibana.local:31406
   |
   v
NGINX Ingress
   |
   v
Kibana Service :5601
   |
   v
Kibana
   |
   v
Elasticsearch :9200

Add the hostname:

sudo sh -c 'echo "10.160.1.22 kibana.local" >> /etc/hosts'

Verify:

kubectl get ingress \
  -n logging

Expected:

NAME             CLASS   HOSTS
elasticsearch    nginx   elasticsearch.local
kibana           nginx   kibana.local

Test:

curl --noproxy '*' \
  -v \
  -H 'Host: kibana.local' \
  http://10.160.1.22:31406/

A successful response should contain:

HTTP/1.1 302 Found
location: /spaces/enter

Open Kibana in a browser:

http://kibana.local:31406

Important:

elasticsearch.local -> Elasticsearch API
kibana.local        -> Kibana Web UI

Kibana should NOT return the Elasticsearch JSON response.

19. Create Kibana Data View

After opening Kibana:

http://kibana.local:31406

Go to:

Stack Management
    |
    +--> Data Views

Create a data view matching the Fluent Bit indices.

For example:

logstash-*

or:

fluent-bit-*

depending on the index configured in Fluent Bit.

Select the timestamp field if available:

@timestamp

Then open:

Analytics
    |
    +--> Discover

You should be able to see Kubernetes logs.

20. Kibana Log Dashboard

Recommended dashboard panels:

+------------------------------------------------+
|              Kubernetes Logs                   |
+------------------------------------------------+
| Total Logs       | Errors       | Warnings     |
+------------------------------------------------+
| Logs Over Time                                |
|                                                |
|              line chart                        |
+------------------------------------------------+
| Logs by Namespace                              |
|              bar chart                         |
+------------------------------------------------+
| Logs by Pod                                    |
|              bar chart                         |
+------------------------------------------------+
| Error Logs                                     |
|              table                             |
+------------------------------------------------+

Useful filters:

kubernetes.namespace_name
kubernetes.pod_name
kubernetes.container_name
log
level

Recommended saved searches:

All Logs
Errors
Warnings
GeoIP API Logs
PostgreSQL Logs
Kubernetes System Logs
21. Deploy GeoIP API

The API image is:

soroushmanhd/geoip-api:latest

Deploy:

kubectl apply \
  -f kubernetes/geoip-api/

Verify:

kubectl get pods \
  -n geoip

Check rollout:

kubectl rollout status \
  deployment/geoip-api \
  -n geoip
22. GeoIP API Ingress

Add:

sudo sh -c 'echo "10.160.1.22 geoip.local" >> /etc/hosts'

Verify:

kubectl get ingress \
  -n geoip

Test:

curl --noproxy '*' \
  -H 'Host: geoip.local' \
  http://10.160.1.22:31406/health

Expected:

{
  "status": "healthy"
}

Country lookup:

curl --noproxy '*' \
  -H 'Host: geoip.local' \
  http://10.160.1.22:31406/country?ip=8.8.8.8

Expected:

{
  "ip": "8.8.8.8",
  "country": "United States",
  "cache_hit": false
}

Readiness:

curl --noproxy '*' \
  -H 'Host: geoip.local' \
  http://10.160.1.22:31406/ready

Metrics:

curl --noproxy '*' \
  -H 'Host: geoip.local' \
  http://10.160.1.22:31406/metrics
23. Local Host Configuration

For local development/testing, add all ingress hostnames to /etc/hosts.

sudo sh -c 'cat >> /etc/hosts <<EOF
10.160.1.22 geoip.local
10.160.1.22 grafana.local
10.160.1.22 prometheus.local
10.160.1.22 kibana.local
10.160.1.22 elasticsearch.local
10.160.1.22 argocd.local
EOF'

Verify:

getent hosts \
  geoip.local \
  grafana.local \
  prometheus.local \
  kibana.local \
  elasticsearch.local \
  argocd.local
24. Service Access Summary

All HTTP services are accessed through the NGINX Ingress NodePort:

10.160.1.22:31406
Hostname	Backend	Port
geoip.local	GeoIP API	80
grafana.local	Grafana	80
prometheus.local	Prometheus	80
kibana.local	Kibana	5601
elasticsearch.local	Elasticsearch	9200
argocd.local	ArgoCD	443

Browser URLs:

http://geoip.local:31406
http://grafana.local:31406
http://prometheus.local:31406
http://kibana.local:31406
http://elasticsearch.local:31406

ArgoCD:

https://argocd.local:31469
25. ArgoCD Installation

ArgoCD provides GitOps-based deployment.

Create namespace:

kubectl create namespace argocd

Install:

kubectl apply \
  -n argocd \
  -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml

Wait:

kubectl wait \
  --for=condition=Available \
  deployment \
  --all \
  -n argocd \
  --timeout=300s

Verify:

kubectl get pods \
  -n argocd

Get admin password:

kubectl get secret \
  argocd-initial-admin-secret \
  -n argocd \
  -o jsonpath="{.data.password}" \
  | base64 -d

Username:

admin
26. Deploy ArgoCD Applications

Apply GitOps applications:

kubectl apply \
  -f gitops/argocd/applications.yaml

Managed applications:

Application	Namespace
geoip-api	geoip
postgres	postgres
monitoring	monitoring
logging	logging

Check:

kubectl get applications \
  -n argocd
27. ArgoCD UI

If using the ingress:

https://argocd.local:31469

Alternatively use port-forward:

kubectl port-forward \
  svc/argocd-server \
  -n argocd \
  8080:443

Open:

https://localhost:8080

Get password:

kubectl get secret \
  argocd-initial-admin-secret \
  -n argocd \
  -o jsonpath="{.data.password}" \
  | base64 -d

Username:

admin
28. Application Testing

Check Kubernetes resources:

kubectl get pods -A

Check services:

kubectl get svc -A

Check ingresses:

kubectl get ingress -A

Test GeoIP:

curl --noproxy '*' \
  -H 'Host: geoip.local' \
  http://10.160.1.22:31406/health

Test Elasticsearch:

curl --noproxy '*' \
  -H 'Host: elasticsearch.local' \
  http://10.160.1.22:31406/

Test Kibana:

curl --noproxy '*' \
  -H 'Host: kibana.local' \
  http://10.160.1.22:31406/

Test Prometheus:

curl --noproxy '*' \
  -H 'Host: prometheus.local' \
  http://10.160.1.22:31406/

Test Grafana:

curl --noproxy '*' \
  -H 'Host: grafana.local' \
  http://10.160.1.22:31406/
29. CI/CD

CI/CD architecture:

Developer
    |
    v
GitLab
    |
    v
GitLab CI
    |
    +--> Test
    |
    +--> Build Docker Image
    |
    v
Docker Hub
    |
    v
ArgoCD
    |
    v
Kubernetes

Pipeline responsibilities:

Component	Responsibility
GitLab CI	Build and test
Docker Hub	Container registry
ArgoCD	Continuous deployment
Kubernetes	Runtime platform

Docker image:

soroushmanhd/geoip-api:latest
30. Environment Variables
Variable	Default	Description
LISTEN_ADDR	:8080	API listen address
DATABASE_URL	postgres://	PostgreSQL connection
GEOIP_PROVIDER_URL	https://ipapi.co	GeoIP provider
GEOIP_PROVIDER_TIMEOUT	5s	Provider timeout
31. Rollback

Kubernetes rollback:

kubectl rollout undo \
  deployment/geoip-api \
  -n geoip

Check rollout:

kubectl rollout status \
  deployment/geoip-api \
  -n geoip

GitOps rollback:

git revert <commit>

git push origin main

ArgoCD will detect the Git change and restore the desired Kubernetes state.

32. Troubleshooting
Check NGINX Ingress
kubectl get pods \
  -n ingress-nginx
kubectl get svc \
  -n ingress-nginx
kubectl get ingress \
  -A
Check NGINX Configuration

To verify that a hostname is correctly routed:

kubectl exec \
  -n ingress-nginx \
  deployment/ingress-nginx-controller \
  -- nginx -T 2>/dev/null \
  | grep -A30 -B10 'kibana.local'

For Elasticsearch:

kubectl exec \
  -n ingress-nginx \
  deployment/ingress-nginx-controller \
  -- nginx -T 2>/dev/null \
  | grep -A30 -B10 'elasticsearch.local'
Check Kibana Backend
kubectl get endpoints \
  -n logging \
  kibana

Expected:

kibana   <pod-ip>:5601

Test directly from inside Kubernetes:

kubectl run curl-kibana \
  --rm -it \
  --restart=Never \
  --image=curlimages/curl \
  -n logging \
  -- \
  curl -I http://kibana:5601/

Expected:

HTTP/1.1 302 Found
location: /spaces/enter
Check Elasticsearch Backend
kubectl get endpoints \
  -n logging \
  elasticsearch

Test:

kubectl run curl-elasticsearch \
  --rm -it \
  --restart=Never \
  --image=curlimages/curl \
  -n logging \
  -- \
  curl http://elasticsearch:9200/
Check Fluent Bit
kubectl logs \
  -n logging \
  -l app=fluent-bit \
  --tail=200
Check Elasticsearch Indices
curl --noproxy '*' \
  -H 'Host: elasticsearch.local' \
  http://10.160.1.22:31406/_cat/indices?v
33. Final Validation

Run:

kubectl get nodes -o wide
kubectl get pods -A
kubectl get svc -A
kubectl get ingress -A

Verify the following endpoints:

GeoIP API:
http://geoip.local:31406

Grafana:
http://grafana.local:31406

Prometheus:
http://prometheus.local:31406

Kibana:
http://kibana.local:31406

Elasticsearch:
http://elasticsearch.local:31406

ArgoCD:
https://argocd.local:31469

Final logging flow:

Kubernetes Container
        |
        v
   Fluent Bit
        |
        v
 Elasticsearch
        |
        v
      Kibana
        |
        v
     Dashboard

Final monitoring flow:

Kubernetes
    |
    +--> node-exporter
    |
    +--> kube-state-metrics
    |
    v
 Prometheus
    |
    +--> Alertmanager
    |
    v
 Grafana

The complete platform is now composed of:

Terraform
    |
    v
Ansible
    |
    v
Kubernetes
    |
    +--> NGINX Ingress
    |
    +--> GeoIP API
    |
    +--> PostgreSQL / CloudNativePG
    |
    +--> Prometheus
    |
    +--> Grafana
    |
    +--> Alertmanager
    |
    +--> Fluent Bit
    |
    +--> Elasticsearch
    |
    +--> Kibana
    |
    +--> ArgoCD