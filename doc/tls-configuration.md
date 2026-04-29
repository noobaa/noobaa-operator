# TLS Configuration for NooBaa

NooBaa supports configuring TLS version, cipher suites, and key exchange group preferences for endpoint and core HTTPS servers. This enables TLS 1.3 enforcement and prepares the platform for Post-Quantum Cryptography (PQC) readiness.

## Interface: TLSProfile Annotation

TLS configuration is injected into NooBaa by annotating the NooBaa CR with the name of a `TLSProfile` CR (from `github.com/red-hat-storage/ocs-tls-profiles/api/v1`) in the same namespace:

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
  namespace: openshift-storage
  annotations:
    noobaa.io/tls-profile-name: odf-tls-profile
```

The operator fetches the referenced `TLSProfile` CR, resolves the `TLSConfig` rule that best matches `noobaa.io` (via `GetConfigForServer`), and applies the resulting TLS settings to the NooBaa endpoint and core pods, as well as to the in-process admission webhook server (when enabled).

If the annotation is absent or empty, NooBaa uses Node.js / OpenSSL defaults.

### Example TLSProfile

```yaml
apiVersion: ocs.openshift.io/v1
kind: TLSProfile
metadata:
  name: odf-tls-profile
  namespace: openshift-storage
spec:
  rules:
    - selectors:
        - "noobaa.io"
      config:
        version: "TLSv1.3"
        ciphers:
          - "TLS_AES_128_GCM_SHA256"
          - "TLS_AES_256_GCM_SHA384"
        groups:
          - "X25519MLKEM768"
          - "X25519"
          - "secp256r1"
```

The following is a minimal profile enabling TLS 1.3 with PQC key exchange:

```yaml
apiVersion: ocs.openshift.io/v1
kind: TLSProfile
metadata:
  name: cluster-tls
spec:
  rules:
    - selectors:
        - "noobaa.io"
      config:
        version: TLSv1.3
        ciphers:
          - TLS_AES_128_GCM_SHA256
          - TLS_AES_256_GCM_SHA384
        groups:
          - X25519MLKEM768
          - secp256r1
```

### Field Reference

| Field | Type | Values | Description |
|---|---|---|---|
| `version` | string **(required)** | `TLSv1.2`, `TLSv1.3` | TLS protocol version; propagated as `TLS_MIN_VERSION`. |
| `ciphers` | []string **(required, min 1)** | IANA TLS cipher suite names | Cipher algorithms; mapped to OpenSSL names for `TLS_CIPHERS`. |
| `groups` | []string **(required, min 1)** | `X25519`, `secp256r1`, `secp384r1`, `secp521r1`, `X25519MLKEM768`, `SecP256r1MLKEM768`, `SecP384r1MLKEM1024` | Key exchange groups; mapped to OpenSSL curve names for `TLS_GROUPS`. |

### Defaults

When the annotation is absent, NooBaa uses Node.js / OpenSSL defaults:
- TLS 1.2 minimum (Node.js default)
- All OpenSSL-supported ciphers
- OpenSSL default group negotiation (includes ML-KEM on OpenSSL 3.5+)

### PQC Readiness Out of the Box

Since Node.js v24.7.0, TLS 1.3 and X25519MLKEM768 are negotiated by default when the underlying OpenSSL supports them. NooBaa has upgraded to Node.js v24.13.0, so even without a custom TLS profile, NooBaa is PQC-ready — clients that support post-quantum key exchange will automatically negotiate X25519MLKEM768 with the endpoint.

## Applying Configuration

### Point the NooBaa CR at a TLSProfile

```bash
kubectl annotate noobaa noobaa -n openshift-storage \
  noobaa.io/tls-profile-name=odf-tls-profile
```

### Revert to defaults (remove annotation)

```bash
kubectl annotate noobaa noobaa -n openshift-storage \
  noobaa.io/tls-profile-name-
```

When the annotation is removed, the operator clears the TLS environment variables to empty strings on the next reconciliation, reverting to Node.js defaults.

## How It Works

On each reconciliation cycle the operator:

1. Reads the `noobaa.io/tls-profile-name` annotation from the NooBaa CR.
2. Fetches the referenced `TLSProfile` CR.
3. Calls `ocstlsv1.GetConfigForServer(profile, "noobaa.io", "")` to resolve the best-matching `TLSConfig`.
4. Maps the resolved config to pod environment variables on the endpoint deployment (phase 4) and core StatefulSet (phase 2):

| TLSConfig field | Environment Variable | Example value |
|---|---|---|
| `version: TLSv1.3` | `TLS_MIN_VERSION` | `TLSv1.3` |
| `ciphers: [...]` | `TLS_CIPHERS` | `TLS_AES_256_GCM_SHA384:TLS_AES_128_GCM_SHA256` |
| `groups: [...]` | `TLS_GROUPS` | `X25519MLKEM768:x25519:prime256v1` |

The admission webhook server (`ReloadTLSConfig`) independently reads the same annotation and fetches the same profile to configure its own Go `tls.Config`.

## Verification

### Check environment variables in the endpoint pod

```bash
kubectl exec -it noobaa-endpoint-<suffix> -n openshift-storage -- bash -c 'env | grep TLS'
```

### Check environment variables in the core pod

```bash
kubectl exec -it noobaa-core-0 -n openshift-storage -c core -- bash -c 'env | grep TLS'
```

Expected output when a TLS 1.3 profile is applied:

```
TLS_MIN_VERSION=TLSv1.3
TLS_CIPHERS=TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384
TLS_GROUPS=X25519MLKEM768:x25519:prime256v1
```

## Manual Testing

### Using openssl

Verify TLS 1.3 with PQC group negotiation (requires OpenSSL 3.5+ with ML-KEM support):

```bash
echo Q | openssl s_client -connect <endpoint-address>:6443 -tls1_3 -groups X25519MLKEM768 2>&1 | grep "Negotiated TLS1.3 group: X25519MLKEM768"
```

A successful match confirms the endpoint is negotiating TLS 1.3 with the X25519MLKEM768 key exchange group.

> **Note:** The `echo Q |` pipes a quit command so `s_client` terminates cleanly after the handshake.

### Using testssl.sh

```bash
docker run --rm -it \
  -v <path-local>:<path-in-container> \
  ghcr.io/testssl/testssl.sh \
  --forward-secrecy --protocols --server-preference --client-simulation \
  --quiet --wide \
  --jsonfile <path-in-container>/testssl_output.json \
  <endpoint-address>
```

Parse the JSON output to verify TLS 1.3 and PQC key exchange:

```bash
jq -r '["TLS1_3", "FS_KEMs"][] as $id | (map(select(.id == $id))[0] // {id: $id, severity: "not found", finding: ""}) | "\(.id): \(.severity) \(.finding)"' <path-local>/testssl_output.json
```

Expected output when TLS 1.3 with PQC is configured:

```
TLS1_3: OK offered
FS_KEMs: OK X25519MLKEM768
```

### Testing the admission webhook server

Pre-requisites — make sure `ENABLE_NOOBAA_ADMISSION == "true"` or pass `--admission` on CLI installation.

The admission webhook runs on port 8080 inside the operator pod.

#### Using openssl (with port-forward)

```bash
kubectl port-forward deploy/noobaa-operator -n openshift-storage 8080:8080
```

```bash
echo Q | openssl s_client -connect localhost:8080 -tls1_3 -groups X25519MLKEM768 2>&1 | grep "Negotiated TLS1.3 group: X25519MLKEM768"
```

#### Using testssl.sh (from within the cluster)

```bash
kubectl run testssl-webhook -it --restart=Never \
  --image=ghcr.io/testssl/testssl.sh -- \
  --forward-secrecy --protocols --server-preference --client-simulation \
  --quiet --wide \
  https://noobaa-operator.openshift-storage.svc.cluster.local:8080
```
