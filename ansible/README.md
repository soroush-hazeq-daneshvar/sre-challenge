# Ansible cluster bootstrap

Installs a kubeadm Kubernetes **1.34** cluster with containerd and Calico. Plays are split so package install, `kubeadm init`, worker join, and CNI are each run once and are safe to re-run.

## Prerequisites

```bash
cd ansible
ansible-galaxy collection install -r requirements.yml
cp inventory/hosts.yml.example inventory/hosts.yml
```

Terraform can also emit inventory:

```bash
terraform -chdir=../terraform output -raw ansible_inventory > inventory/hosts.yml
```

## Run

```bash
ansible all -m ping
ansible-playbook site.yml
```

Useful tags: `common`, `containerd`, `packages`, `init`, `join`, `calico`.

## Version pins

| Component | Default |
|-----------|---------|
| Kubernetes packages | `1.34.*` from `pkgs.k8s.io` |
| containerd | `containerd.io` from Docker apt |
| Calico | `v3.32.1` (tested with Kubernetes 1.34) |
| Pause image | `registry.k8s.io/pause:3.10` |

Override in `group_vars/all.yml`. Use `k8s_package_version: "1.34.1-1.1"` to pin a patch.
