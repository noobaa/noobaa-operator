# Operator Log Level Configuration

## Design

Add a separate `OPERATOR_LOG_LEVEL` key in the `noobaa-config` ConfigMap to control
operator log verbosity independently from core/endpoint pods. Changing this key does
**not** trigger a pod rollout for core or endpoint pods.

Additionally, noisy Update/Generic event predicate logs are demoted from Info to Debug
level, so they only appear when `OPERATOR_LOG_LEVEL` is set to `debug`.

### Supported Values

| Value | logrus Level | Behavior |
|-------|-------------|----------|
| `warn` | WarnLevel | Minimal logging: only warnings and errors |
| `info` | InfoLevel | **Default.** Create/Delete events, phase transitions, reconcile start/done |
| `debug` | DebugLevel | Verbose: includes Update/Generic event predicates, all debug traces |

### Config Source

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: noobaa-config
data:
  NOOBAA_LOG_LEVEL: "default_level"   # controls core + endpoint pods
  OPERATOR_LOG_LEVEL: "info"           # controls operator process only
```

`OPERATOR_LOG_LEVEL` is excluded from the ConfigMap hash used in pod-template
annotations. The operator computes a hash of the ConfigMap data and stores it in
the `noobaa.io/configmap-hash` annotation on core/endpoint pod templates. When
this hash changes, Kubernetes detects a spec change and triggers a rolling
restart. By filtering out `OPERATOR_LOG_LEVEL` before computing the hash,
changing the operator's log verbosity does not cause core/endpoint pod rollouts.

### CLI Flag

A separate `--operator-log-level` flag controls the initial operator log level at startup,
independent from `--debug-level` (which controls core/endpoint via `NOOBAA_LOG_LEVEL`).

```bash
noobaa-operator operator --operator-log-level=warn
```

Once the first reconcile completes and the ConfigMap is read, `OPERATOR_LOG_LEVEL` takes over.

### How to Change at Runtime

```bash
# Quiet the operator (warnings and errors only)
kubectl patch configmap noobaa-config -n <namespace> \
  -p '{"data":{"OPERATOR_LOG_LEVEL":"warn"}}'

# Verbose debugging
kubectl patch configmap noobaa-config -n <namespace> \
  -p '{"data":{"OPERATOR_LOG_LEVEL":"debug"}}'

# Back to default
kubectl patch configmap noobaa-config -n <namespace> \
  -p '{"data":{"OPERATOR_LOG_LEVEL":"info"}}'
```

Or use the CLI:

```bash
nb operator set-log-level debug -n <namespace>
```

The operator picks up the change on the next reconcile cycle (no restart needed).
Core and endpoint pods are **not** affected.

### Example Output

```bash
nb operator set-log-level debug -n test1

INFO[0000] ✅ Updated: ConfigMap "noobaa-config"

Operator log level was set to "debug" successfully
The operator picks up the change on the next reconcile. Core and endpoint pods are not restarted.
```
