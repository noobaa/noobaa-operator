# SCC (Security Context Constraints) Annotation in NooBaa Components

## Background

On OpenShift, a Security Context Constraint (SCC) is a permission profile for pods. It answers questions like:

- Can this pod run as root? (`runAsUser`)
- Can it use extra Linux permissions? (`allowedCapabilities`, `allowPrivilegeEscalation`)

When NooBaa is installed, the operator creates custom SCCs in the cluster: `noobaa-core`, `noobaa-endpoint`, and `noobaa-agent`. The core, endpoint, and agent pods use these dedicated profiles, while other components rely on the built-in `restricted-v2` SCC.

The [`openshift.io/required-scc`](https://docs.redhat.com/en/documentation/openshift_dedicated/4/html/authentication_and_authorization/managing-pod-security-policies#security-context-constraints-requiring_configuring-internal-oauth) annotation tells OpenShift which SCC profile a workload must use. During pod creation, OpenShift records the granted profile in the pod's `openshift.io/scc` annotation.

## For All NooBaa Deployments

| Component | Type | `openshift.io/required-scc` value |
| :--- | :--- | :--- |
| `noobaa-core` | StatefulSet| `noobaa-core` |
| `noobaa-endpoint` | Deployment| `noobaa-endpoint` |
| `noobaa-agent` | Pod | `noobaa-agent` |
| `noobaa-operator` | Deployment| `restricted-v2` |
| `cnpg-controller-manager` | Deployment | `restricted-v2` |
| `noobaa-db-pg-cluster-<number>` | Cluster | `restricted-v2` |

**Note:** For the DB pods, this is relevant only for new deployments (not upgrade).

## Other

| Component | Type | `openshift.io/required-scc` value |
| :--- | :--- | :--- |
| `noobaa-db-pg-cluster-<number>-init` | Job | `restricted-v2` |
| `prometheus-adapter` | Deployment| `restricted-v2` |
| `analyze-resource` | Job | `restricted-v2` |
