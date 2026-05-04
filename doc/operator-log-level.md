# Operator Log Level Configuration

## Problem

The NooBaa operator reconciles frequently and generates a high volume of log messages
(e.g., "Update event detected for noobaa, queuing Reconcile" every few seconds).
For deployments at scale (e.g., 32 ODF clusters), this overwhelms central logging systems.
Customers cannot reduce retention periods due to regulatory requirements.

## Current State

| Component | Config mechanism | Persistent? |
|-----------|-----------------|-------------|
| Core / Endpoints | `NOOBAA_LOG_LEVEL` in `noobaa-config` ConfigMap | Yes (pod rollout via hash annotation) |
| Core runtime | `noobaa system set-debug-level` RPC | No |
| Operator | Hardcoded: `warn` → WarnLevel, anything else → DebugLevel | N/A (binary toggle only) |

The operator had no independent, runtime-configurable log level. Changing `NOOBAA_LOG_LEVEL`
affected core/endpoint pods (triggering a rolling restart via ConfigMap hash annotation) but
gave the operator only a binary warn-vs-debug toggle with no Info level support.

## Design

Add a **separate** `OPERATOR_LOG_LEVEL` key in the same `noobaa-config` ConfigMap.
This decouples operator logging from core/endpoint logging so that changing the operator's
verbosity does **not** trigger a pod rollout for core or endpoint pods.

### Why a separate key?

Changing `NOOBAA_LOG_LEVEL` updates the ConfigMap hash annotation on core StatefulSet and
endpoint Deployment pod templates, causing a rolling restart. An admin who only wants to
quiet the operator should not inadvertently restart data-path pods.

### Supported Values

| Value | logrus Level | Behavior |
|-------|-------------|----------|
| `warn` | WarnLevel | Minimal logging: only warnings and errors |
| `info` | InfoLevel | **Default.** Create/Delete events, phase transitions, reconcile start/done |
| `debug` | DebugLevel | Verbose: includes Update/Generic event predicates, all debug traces |

### Config Source

The existing `noobaa-config` ConfigMap with the new `OPERATOR_LOG_LEVEL` key:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: noobaa-config
data:
  NOOBAA_LOG_LEVEL: "default_level"   # controls core + endpoint pods
  OPERATOR_LOG_LEVEL: "info"           # controls operator process only
```

### Runtime Application

The operator reconciles `noobaa-config` on every cycle via `ReconcileObject(r.CoreAppConfig, ...)`.
After the ConfigMap is reconciled, the operator reads `OPERATOR_LOG_LEVEL` and calls
`util.InitLogger()` with the corresponding level. The change takes effect immediately on
the next reconcile — **no operator pod restart required**, and **no core/endpoint pod rollout**.

### Startup Behavior

On operator startup (`RunOperator`), before any reconciliation:
- If `--debug-level warn` → start at Warn level
- If `--debug-level debug` → start at Debug level
- Otherwise (including `default_level`) → start at Info level

Once the first reconcile completes and the ConfigMap is read, `OPERATOR_LOG_LEVEL` takes over.

### How to Change at Runtime

```bash
# Quiet the operator (warnings and errors only) — no core/endpoint restart
kubectl patch configmap noobaa-config -n <namespace> \
  -p '{"data":{"OPERATOR_LOG_LEVEL":"warn"}}'

# Verbose debugging for the operator
kubectl patch configmap noobaa-config -n <namespace> \
  -p '{"data":{"OPERATOR_LOG_LEVEL":"debug"}}'

# Back to normal
kubectl patch configmap noobaa-config -n <namespace> \
  -p '{"data":{"OPERATOR_LOG_LEVEL":"info"}}'
```

The operator picks up the change on the next reconcile cycle (no restart needed).
Core and endpoint pods are **not** affected.

## Files Changed

| File | Change |
|------|--------|
| `pkg/util/util.go` | Add `OperatorLogLevel()` helper to map string → logrus level |
| `pkg/util/predicates.go` | Update/Generic events log at Debug level (visible only at `debug`) |
| `pkg/system/phase2_creating.go` | Add `OPERATOR_LOG_LEVEL` to defaults; read it and apply via `OperatorLogLevel` |
| `pkg/system/system.go` | Seed `OPERATOR_LOG_LEVEL` from `--debug-level` flag in `LoadConfigMapFromFlags` |
| `pkg/operator/manager.go` | Update startup log level to use `OperatorLogLevel` (supports info as default) |
