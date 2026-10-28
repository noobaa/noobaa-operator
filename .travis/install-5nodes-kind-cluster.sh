#!/bin/sh
set -o errexit

# Pin to 1.30 line (override with any v1.30.z via KIND_NODE_IMAGE)
KIND_NODE_IMAGE="${KIND_NODE_IMAGE:-kindest/node:v1.30.13}"
# Number of worker nodes (default 4 preserves the historical 5-node cluster).
# Set KIND_WORKER_COUNT=0 for a single control-plane node only.
KIND_WORKER_COUNT="${KIND_WORKER_COUNT:-4}"
KIND_CLUSTER_NAME="${KIND_CLUSTER_NAME:-5nodes}"

# create registry container unless it already exists
reg_name='kind-registry'
reg_port='5000'
running="$(docker inspect -f '{{.State.Running}}' "${reg_name}" 2>/dev/null || true)"
if [ "${running}" != 'true' ]; then
  docker run \
    -d --restart=always -p "127.0.0.1:${reg_port}:5000" --name "${reg_name}" \
    registry:2
fi

# Build kind node list: always one control-plane, then KIND_WORKER_COUNT workers
nodes_yaml="- role: control-plane
  image: ${KIND_NODE_IMAGE}"
i=0
while [ "${i}" -lt "${KIND_WORKER_COUNT}" ]; do
  nodes_yaml="${nodes_yaml}
- role: worker
  image: ${KIND_NODE_IMAGE}"
  i=$((i + 1))
done

# create a cluster with the local registry enabled in containerd (node image: 1.30 line, overridable via KIND_NODE_IMAGE)
cat <<EOF | kind create cluster --name "${KIND_CLUSTER_NAME}" --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
containerdConfigPatches:
- |-
  [plugins."io.containerd.grpc.v1.cri".registry.mirrors."localhost:${reg_port}"]
    endpoint = ["http://${reg_name}:${reg_port}"]
nodes:
${nodes_yaml}
EOF

# connect the registry to the cluster network
# (the network may already be connected)
docker network connect "kind" "${reg_name}" || true

# Document the local registry
# https://github.com/kubernetes/enhancements/tree/master/keps/sig-cluster-lifecycle/generic/1755-communicating-a-local-registry
cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: local-registry-hosting
  namespace: kube-public
data:
  localRegistryHosting.v1: |
    host: "localhost:${reg_port}"
    help: "https://kind.sigs.k8s.io/docs/user/local-registry/"
EOF
