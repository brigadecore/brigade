#!/usr/bin/env sh

set -o errexit

# Create a local Docker image registry that we'll hook up to kind
reg_name="${KIND_REGISTRY_NAME:-kind-registry}"
reg_port="${KIND_REGISTRY_PORT:-5000}"
running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  docker run \
    -d --restart=always -p "${reg_port}:5000" --name "${reg_name}" \
    registry:2
fi

# Create a kind cluster with the local Docker image registry enabled
kind get clusters | grep brigade || cat <<EOF | kind create cluster --name brigade --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
  image: brigadecore/kind-node:v1.23.3
  extraPortMappings:
  - containerPort: 31600
    hostPort: 31600
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:5000"]
EOF

# Make sure the local Docker image registry is connected to the same network
# as the kind cluster
docker network connect "kind" "${reg_name}" || true

# Tell each node to use the local Docker image registry
for node in $(kind --name brigade get nodes); do
  kubectl annotate node --overwrite "${node}" "kind.x-k8s.io/registry=localhost:${reg_port}"
done

# Set up NFS
helm repo ls | grep https://charts.helm.sh/stable || helm repo add stable https://charts.helm.sh/stable
helm upgrade nfs stable/nfs-server-provisioner --install --create-namespace --namespace nfs
