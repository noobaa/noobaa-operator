[NooBaa Operator](../README.md) /
# Connection Update

The `noobaa connection update` command migrates all BackingStores and NamespaceStores that reference a given endpoint URL to a new endpoint URL.

This is useful when the underlying S3-compatible storage service has moved to a new address (e.g. IP change, DNS migration, load-balancer swap) and the NooBaa configuration needs to follow.

## Supported Store Types

The command operates on stores whose spec contains an endpoint field:

- **BackingStore:** `s3-compatible`, `ibm-cos`
- **NamespaceStore:** `s3-compatible`, `ibm-cos`

Other store types (aws-s3, azure-blob, google-cloud-storage, pv-pool) do not carry a user-specified endpoint and are not affected.

## Usage

```shell
noobaa connection update --old-endpoint <CURRENT_URL> --new-endpoint <NEW_URL>
```

| Flag | Required | Description |
|---|---|---|
| `--old-endpoint` | Yes | The current endpoint URL to replace |
| `--new-endpoint` | Yes | The new endpoint URL to set |
| `-n / --namespace` | No | Kubernetes namespace (defaults to current context namespace) |

### Example

```shell
noobaa connection update \
  --old-endpoint http://minio-old.example.com:9000 \
  --new-endpoint http://minio-new.example.com:9000
```

## How It Works

The command performs the following steps in order. If any step fails, all changes made so far are rolled back automatically.

### 1. Discover matching stores

All BackingStores and NamespaceStores in the target namespace are listed. Stores whose spec endpoint matches `--old-endpoint` are selected.

### 2. Read credentials

For each matched store, the referenced Kubernetes Secret is read to extract the access credentials (identity, secret key, auth method) needed for validation.

### 3. Pre-validate the new endpoint

For every matched store, `CheckExternalConnection` is called against the NooBaa core to verify that the new endpoint is reachable and the credentials are accepted. No changes are made at this stage.

- **BackingStores** are validated with the target bucket (triggers `noobaa_blocks/` prefix verification).
- **NamespaceStores** are validated without a bucket parameter (connectivity-only check via `ListBuckets`).

If any store fails pre-validation the command aborts immediately with no side effects.

### 4. Pause reconciliation

The annotation `noobaa.io/pause-reconcile: "true"` is set on every matched store CR. While this annotation is present, the BackingStore and NamespaceStore reconcilers will skip reconciliation and requeue after 5 seconds. This prevents the reconcilers from acting on the partially-updated state.

### 5. Patch CR specs

Each matched store's `.spec.s3Compatible.endpoint` or `.spec.ibmCos.endpoint` field is updated to the new URL via `KubeUpdate`.

### 6. Update core connections

The command reads the NooBaa system info to discover all external connections that reference the old endpoint. Connections are deduplicated by name (the same connection can appear across multiple accounts). Each unique connection is updated via `UpdateExternalConnection`.

### 7. Resume reconciliation

The `noobaa.io/pause-reconcile` annotation is removed from all matched stores, allowing the reconcilers to resume normal operation against the new endpoint.

## Rollback Behaviour

The command implements automatic rollback at every failure point:

| Failure during | What gets rolled back |
|---|---|
| CR spec patching (step 5) | Already-patched CRs are reverted to the old endpoint; pause annotations are removed |
| Core connection update (step 6) | Already-updated connections are reverted to the old endpoint, already-patched CRs are reverted, pause annotations are removed |

If a connection revert itself fails during rollback, the command logs a `MANUAL ACTION REQUIRED` message listing the connection names that could not be reverted. An operator must then manually correct those connections via the NooBaa management API or UI.

## Pre-flight Checklist

Before running the command, verify:

1. **New endpoint is valid** 
2. **Credentials are valid** for the new endpoint.
2. **No other stores use the new endpoint yet** for a different purpose. The command does not check for conflicts with existing connections that already point to the new endpoint.

## Annotations Reference

| Annotation | Value | Purpose |
|---|---|---|
| `noobaa.io/pause-reconcile` | `"true"` | Temporarily halts reconciliation for the annotated BackingStore or NamespaceStore |

The annotation is set automatically by the command and removed on completion or rollback. In rare cases (e.g. operator crash mid-update), it may need to be removed manually:

```shell
kubectl annotate backingstore <STORE_NAME> noobaa.io/pause-reconcile- -n <NAMESPACE>
kubectl annotate namespacestore <STORE_NAME> noobaa.io/pause-reconcile- -n <NAMESPACE>
```

## Troubleshooting

### `MANUAL ACTION REQUIRED`
A rollback could not fully revert one or more NooBaa core connections. The listed connections still point to the new endpoint while the store CRs have been reverted to the old endpoint. Use the NooBaa management console or RPC to manually update those connections back to the old endpoint.

### Pause annotation left behind
If the operator or CLI crashes between steps 4 and 7, stores may be left with `noobaa.io/pause-reconcile: "true"`. Remove the annotation manually as shown in the Annotations Reference section above.
