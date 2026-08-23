all:
  children:
    k8s_cluster:
      children:
        control_plane:
          hosts:
            control-plane:
              ansible_host: ${control_plane_ip}
              node_role: control-plane
        workers:
          hosts:
%{ for idx, ip in worker_ips ~}
            worker-${idx + 1}:
              ansible_host: ${ip}
              node_role: worker
%{ endfor ~}
