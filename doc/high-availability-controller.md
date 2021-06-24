### High Availability (HA) controller

## Abstract

Improve NooBaa pods recovery in the case of a node failure. Find NooBaa pods running on the failing node and force to delete them to speed up rescheduling in the healthy node.

## Problem

* In cases where a node is in `NotReady` state, it takes 5 minutes for deployment pods (kubernetes Deployment to failover on a different node.
* Statefulset pods, such as noobaa-core and noobaa-db pods, are not restarting automatically until the old pod is explicitly force deleted.
* For pods that are connected to a PV, such as noobaa-db, after the pod is force deleted, it takes more time (~8 minutes) for the PV to detached from the old pod so that the new pod can attach to the PV.

## Solution

To make the pods failover faster to a new node, noobaa-operator [watches](https://kubernetes.io/docs/reference/using-api/api-concepts/#efficient-detection-of-changes) the cluster node states. When a node transitions from `Ready` to `NotReady` status, the HA Controller looks for NooBaa pods on that node, these pods will be force deleted. Once deleted the pods will restart on a new `Ready` node.

```
                            +-----------------+
                            |  HA Controller  |
                            +--------+--------+
                                     |
                                     |
                                     |
+--------+     +---------------------+------------------------+
|  ETCD  +-----+                 API Server                   |
+--------+     +----+----------------+-----------------+------+
                    |                |                 |
                    |                |                 |
                    |                |                 |
               +----+-----+     +----+-----+      +----+-----+
               |   Node   |     |   Node   |      |   Node   |
               +----------+     +----------+      +----------+
```

## Implementation

[High Availability (HA) controller](https://github.com/noobaa/noobaa-operator/pull/672) is a [controller](https://sdk.operatorframework.io/docs/building-operators/golang/references/client/) defining kubernetes [Nodes](https://pkg.go.dev/k8s.io/api/core/v1#Node) as the source of its [events](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/event).

### Node failure flow

1. Communication between [K8S API server](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-apiserver/) and [kubelet](https://kubernetes.io/docs/reference/command-line-tools-reference/kubelet/) running on a worker node is severed
2. API server marks the worker node state as `NotReady`
3. HA Controller (HAC) watching cluster nodes states, detect a worker node state transition, reconciliation is initiated.
4. The HAC Reconciler lists NooBaa pods on the failing node and requests API server to delete those pods. The new pod state is committed into ETCD
5. [The pod controller](https://kubernetes.io/docs/reference/command-line-tools-reference/kube-controller-manager/) (Deployment, StatefulSet, etc) reacts to pod deletion and reschedules the pod on a healthy node

### Node readiness condition

Node is ready, if there is a `NodeReady` [node condition](https://pkg.go.dev/k8s.io/api/core/v1#NodeCondition) in node's status. A worker node becomes not ready if the connection between the worker and the master node was broken, the node rebooted, or any other communication error between the K8S API Server and the kubelet process.

### Predicate

[Predicates](https://pkg.go.dev/sigs.k8s.io/controller-runtime/pkg/predicate) allow controllers to filter events before they are provided to EventHandlers. There are several kinds of events, such as CreateEvents, GenericEvent, DeleteEvent and UpdateEvent. Update event where old Node state is `Ready` and current state is `NotReady` indicates a node goes down event. All other events are filtered out.

### Event handler

`Reconcile()` is called when a node in the cluster transitions from `Ready` to `NotReady` state. High Availability (HA) controller lists all NooBaa pods in the failing node filtering using pods label, namespace, and name of the node conditions:

- Pod is labeled with `app=noobaa`
- Pod runs in the watched namespace
- Pod runs on the failed node

All the pods matching the above are force-deleted to allow fast rescheduling on a healthy node.


## TODO

For noobaa-db pod which is attached to a PV it may take more time until the new pod can attach to the PV. Add noobaa-db PV handling (?) detach db PV from the failing node.
