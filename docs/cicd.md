# CI/CD Pipeline Documentation

## Overview

The project uses GitLab CI/CD for automated build, test, and deployment, combined with ArgoCD for GitOps-based continuous delivery.

## Pipeline Stages

```
validate → test → build → deploy
```

### Stage 1: Validate

| Job | Tool | Purpose |
|-----|------|---------|
| `validate_terraform` | Terraform | Validate IaC syntax and formatting |
| `validate_ansible` | ansible-lint | Lint Ansible playbooks |
| `lint_go` | go vet, go fmt | Static analysis for Go code |
| `validate_k8s_manifests` | conftest | Policy validation for K8s manifests |

### Stage 2: Test

| Job | Tool | Purpose |
|-----|------|---------|
| `test_go` | go test | Unit tests with race detection and coverage |

Uses PostgreSQL service container for integration testing.

### Stage 3: Build

| Job | Tool | Purpose |
|-----|------|---------|
| `build_image` | Docker | Multi-stage build, push to GitLab Container Registry |

**Docker Image:**
- Base: `gcr.io/distroless/static-debian12:nonroot`
- Tags: `$CI_COMMIT_SHA` and `latest`
- Runs as non-root user

### Stage 4: Deploy

| Job | Environment | Trigger | Approval |
|-----|-------------|---------|----------|
| `deploy_staging` | staging | Auto on `develop` branch | None |
| `deploy_production` | production | Manual on `main` branch | Required |

## GitOps with ArgoCD

```
Developer → Git Push → GitLab CI (build & push image)
                          ↓
                    Update manifest (image tag)
                          ↓
                    ArgoCD detects drift
                          ↓
                    Auto-sync to cluster
```

### ArgoCD Applications

| App | Path | Namespace | Auto-sync |
|-----|------|-----------|-----------|
| geoip-api | kubernetes/geoip-api | geoip | Yes (prune + selfHeal) |
| postgres | kubernetes/postgres | postgres | Yes |
| monitoring | kubernetes/monitoring | monitoring | Yes (selfHeal) |
| logging | kubernetes/logging | logging | Yes (selfHeal) |

## Required CI/CD Variables

| Variable | Description |
|----------|-------------|
| `CI_REGISTRY_USER` | GitLab registry username |
| `CI_REGISTRY_PASSWORD` | GitLab registry password/token |
| `KUBE_CONTEXT_STAGING` | kubectl context for staging |
| `KUBE_CONTEXT_PRODUCTION` | kubectl context for production |

## Branch Strategy

```
main (production) ← merge ← develop (staging) ← feature branches
```

- Feature branches: validate + test only
- `develop`: validate + test + build + deploy staging
- `main`: validate + test + build + manual deploy production

## Deployment Flow

1. Developer pushes code to feature branch
2. MR triggers validate + test stages
3. Merge to `develop` triggers staging deployment
4. QA validates on staging
5. Merge to `main`, manually trigger production deployment
6. ArgoCD ensures cluster matches Git state

## Rollback Procedure

```bash
# Via kubectl
kubectl rollout undo deployment/geoip-api -n geoip

# Via ArgoCD
argocd app rollback geoip-api

# Via Git (GitOps way)
git revert <commit-sha>
git push origin main
# ArgoCD auto-syncs the reverted state
```
