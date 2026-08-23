# Architecture Documentation

# SRE Challenge Architecture

## High Level Overview

The platform follows a Cloud-Native architecture based on:

- Infrastructure as Code (Terraform)
- Configuration Management (Ansible)
- Kubernetes orchestration
- GitOps deployment with ArgoCD
- Containerized application delivery
- Observability stack (Prometheus + Grafana + Alertmanager)
- Centralized logging (Fluent Bit + Elasticsearch + Kibana)


```
                         Developer
                            |
                            |
                            v
                    GitHub Repository
                            |
                            |
          +-----------------+----------------+
          |                                  |
          v                                  v
 Infrastructure Code                 Application Code
 Terraform                           Go GeoIP API
 Ansible                             Dockerfile
          |                                  |
          |                                  |
          v                                  v
 Infrastructure Provisioning        Docker Image Build
          |                                  |
          |                                  v
          |                          Docker Hub
          |                                  |
          |                                  |
          +----------------+-----------------+
                           |
                           v
                       ArgoCD
                           |
                           v
                  Kubernetes Cluster
```

---

# Infrastructure Layer

## Terraform

Location:

```
terraform/
```

Terraform is responsible for infrastructure provisioning.

Responsibilities:

- Virtual Machine creation
- Network configuration
- Security groups
- Infrastructure state management


Workflow:

```
Terraform
    |
    v
Cloud Provider
    |
    v
Virtual Machines
```


---

## Ansible

Location:

```
ansible/
```

Ansible configures the provisioned machines.

Responsibilities:

- Operating system preparation
- Kernel configuration
- Container runtime installation
- Kubernetes installation
- Cluster initialization


Workflow:

```
Ansible
    |
    v
Kubernetes Cluster
```


---

# Application Delivery Flow


```
GitHub Repository

        |
        |
        v

Go Application Source

app/geoip-api

        |
        |
        v

Docker Build

        |
        |
        v

Docker Hub

soroushmanhd/geoip-api:latest

        |
        |
        v

ArgoCD

        |
        |
        v

Kubernetes Deployment
```

---

# Kubernetes Architecture


```
Kubernetes Cluster

|
|
+-- kube-system
|
|   +-- Kubernetes Control Plane
|   +-- CoreDNS
|   +-- kube-proxy
|
|
+-- calico-system
|
|   +-- Calico CNI
|
|
+-- ingress-nginx
|
|   +-- NGINX Ingress Controller
|
|
+-- argocd
|
|   +-- ArgoCD Server
|   +-- Application Controller
|
|
+-- geoip
|
|   +-- geoip-api Deployment
|   |
|   +-- Service
|   |
|   +-- Ingress
|
|
+-- postgres
|
|   +-- CloudNativePG Cluster
|
|
+-- monitoring
|
|   +-- kube-prometheus-stack
|       |
|       +-- Prometheus
|       +-- Grafana
|       +-- Alertmanager
|
|
+-- logging
|
    +-- Fluent Bit
    |
    +-- Elasticsearch
    |
    +-- Kibana
```

---

# Application Architecture

## GeoIP API


```
Client Request

      |
      v

NGINX Ingress

      |
      v

geoip-api Service

      |
      v

GeoIP API Pods


      |
      +----------------+
      |                |
      v                v

PostgreSQL        External Provider

CloudNativePG     ipapi.co
```

---

# GeoIP API Components


## Deployment

Namespace:

```
geoip
```


Deployment:

```
geoip-api
```

Features:

- Multiple replicas
- Health checks
- Readiness probes
- Prometheus metrics endpoint
- Non-root container


Container image:

```
soroushmanhd/geoip-api:latest
```

---

# Database Architecture


## CloudNativePG


Namespace:

```
postgres
```


Architecture:


```
             PostgreSQL Cluster


                 Primary

                    |
        +-----------+-----------+

        |                       |

        v                       v

     Replica 1              Replica 2

```


Features:

- Automatic failover
- Streaming replication
- Persistent storage
- Kubernetes native management


---

# Monitoring Architecture


Namespace:

```
monitoring
```


Stack:


```
Application Metrics

        |

        v

Prometheus

        |

        +------------+

        |            |

        v            v

    Grafana    Alertmanager

```


Collected metrics:

- API request count
- Request latency
- Cache hit ratio
- External API failures
- Kubernetes resources
- Node health


---

# Logging Architecture


Namespace:

```
logging
```


Flow:


```
Kubernetes Pods

        |

        v

Fluent Bit

(DaemonSet)

        |

        v

Elasticsearch

        |

        v

Kibana
```


Purpose:

- Centralized log collection
- Search
- Troubleshooting
- Operational visibility


---

# GitOps Architecture


ArgoCD watches:

```
gitops/argocd/applications.yaml
```


Repository:

```
https://github.com/soroush-hazeq-daneshvar/sre-challenge
```


Application synchronization:


```
Git Repository

        |

        v

ArgoCD

        |

        v

Kubernetes Desired State

        |

        v

Running Cluster State
```


ArgoCD features enabled:

- Automated sync
- Self healing
- Resource pruning


---

# Network Flow


```
User

 |

 |

 v

Ingress-NGINX

 |

 |

 v

geoip-api Service

 |

 |

 v

GeoIP API Pod

 |

 +----------------+
 |
 v

CloudNativePG PostgreSQL

```


---

# Complete Platform View


```
                         GitHub

                            |

                            v

                  CI / Docker Build

                            |

                            v

                       Docker Hub

                            |

                            v

                         ArgoCD

                            |

                            v

                  Kubernetes Cluster

        +-------------------+-------------------+

        |                   |                   |

        v                   v                   v


   Applications        Data Layer        Observability


   GeoIP API          PostgreSQL        Prometheus

   Ingress            CloudNativePG     Grafana

                                      Alertmanager


                                      Logging


                                      Fluent Bit

                                      Elasticsearch

                                      Kibana

```

---

# Design Principles

## Reliability

- Kubernetes self-healing
- PostgreSQL failover
- Multiple application replicas


## Scalability

- Horizontal pod scaling
- Additional worker nodes
- Database replication


## Maintainability

- Terraform IaC
- Ansible automation
- GitOps workflow


## Observability

- Metrics with Prometheus
- Dashboards with Grafana
- Logs with Elasticsearch/Kibana


## Security

- Non-root containers
- Kubernetes RBAC
- OS hardening
- Network isolation