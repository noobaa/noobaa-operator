# TLS Configuration for NooBaa Endpoints

NooBaa supports configuring TLS version, cipher suites, and key exchange group preferences for endpoint HTTPS servers. This enables TLS 1.3 enforcement and prepares the platform for Post-Quantum Cryptography (PQC) readiness.

## NooBaa CR Interface

TLS configuration is set under `spec.security.apiServerSecurity` on the NooBaa custom resource. The StorageCluster propagates the platform API Server TLS profile here and NooBaa applies it to endpoint HTTPS servers.

```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
  namespace: openshift-storage
spec:
  security:
    apiServerSecurity:
      tlsMinVersion: "VersionTLS13"
      tlsCiphers:
        - "TLS_AES_128_GCM_SHA256"
        - "TLS_AES_256_GCM_SHA384"
      tlsGroups:
        - "X25519MLKEM768"
        - "X25519"
        - "secp256r1"
```

### Field Reference

| Field | Type | Values | Description |
|---|---|---|---|
| `tlsMinVersion` | string (optional, nullable) | `VersionTLS12`, `VersionTLS13` | Minimum TLS protocol version negotiated during handshake. |
| `tlsCiphers` | []string (optional) | OpenSSL cipher names | Cipher algorithms negotiated during the TLS handshake. |
| `tlsGroups` | []TLSGroup (optional) | `X25519`, `secp256r1`, `secp384r1`, `secp521r1`, `X25519MLKEM768` | Key exchange group preferences WIP - waiting for final list from DF |

### Defaults

When no fields are set, NooBaa uses Node.js / OpenSSL defaults:
- TLS 1.2 minimum (Node.js default)
- All OpenSSL-supported ciphers
- OpenSSL default group negotiation (includes ML-KEM on OpenSSL 3.5+)

### PQC Readiness Out of the Box

Since Node.js v24.7.0, TLS 1.3 and X25519MLKEM768 are negotiated by default when the underlying OpenSSL supports them (see [release notes](https://github.com/nodejs/node/releases/tag/v24.7.0)). NooBaa has upgraded to Node.js v24.13.0 (see [PR #9458](https://github.com/noobaa/noobaa-core/pull/9458)), which includes this behavior. This means that even without setting any custom TLS properties on the NooBaa CR, NooBaa is PQC-ready — clients that support post-quantum key exchange will automatically negotiate X25519MLKEM768 with the endpoint.

## Applying Configuration

### Set TLS 1.3 with PQC groups

```bash
kubectl patch noobaa noobaa -n openshift-storage --type merge -p '{
  "spec": {
    "security": {
      "apiServerSecurity": {
        "tlsMinVersion": "VersionTLS13",
        "tlsGroups": ["X25519MLKEM768", "X25519", "secp256r1"]
      }
    }
  }
}'
```

### Set cipher suites only

```bash
kubectl patch noobaa noobaa -n openshift-storage --type merge -p '{
  "spec": {
    "security": {
      "apiServerSecurity": {
        "tlsCiphers": ["TLS_AES_256_GCM_SHA384", "TLS_AES_128_GCM_SHA256"]
      }
    }
  }
}'
```

### Revert to defaults (remove TLS configuration)

```bash
kubectl patch noobaa noobaa -n openshift-storage --type json -p '[
  {"op": "remove", "path": "/spec/security/apiServerSecurity"}
]'
```

When the configuration is removed, the operator sets the corresponding environment variables to empty strings on the next reconciliation, reverting the endpoint to Node.js defaults.

## How It Works

The operator reconciler (`SetDesiredDeploymentEndpoint`) maps the CR fields to endpoint pod environment variables:

| CR Field | Environment Variable | Example Value |
|---|---|---|
| `tlsMinVersion: VersionTLS13` | `TLS_MIN_VERSION` | `TLSv1.3` |
| `tlsCiphers: [...]` | `TLS_CIPHERS` | `TLS_AES_256_GCM_SHA384:TLS_AES_128_GCM_SHA256` |
| `tlsGroups: [...]` | `TLS_GROUPS` | `X25519MLKEM768:X25519:secp256r1` |

These env vars are defined in the endpoint deployment template and updated in-place on every reconciliation cycle. Updating the CR triggers a deployment rollout with the new TLS settings.

## Verification

### Check environment variables in the endpoint pod

```bash
kubectl exec -it noobaa-endpoint-<suffix> -n openshift-storage -- bash -c 'env | grep TLS'
```

Expected output when TLS is configured:

```
TLS_MIN_VERSION=TLSv1.3
TLS_CIPHERS=TLS_AES_128_GCM_SHA256:TLS_AES_256_GCM_SHA384
TLS_GROUPS=X25519MLKEM768:X25519:secp256r1
```

When TLS configuration is removed from the CR, the variables will be present but empty.

## Manual Testing

### Applying TLS config and verifying via pod logs

#### 1. Patch the NooBaa CR

```bash
kubectl patch noobaa noobaa -n <namespace> --type=merge -p '{
  "spec": {
    "security": {
      "apiServerSecurity": {
        "tlsMinVersion": "TLSv1.3",
        "tlsCiphers": [
          "TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256",
          "BLALALA",
          "TLS_CHACHA20_POLY1305_SHA256"
        ],
        "tlsGroups": ["X25519"]
      }
    }
  }
}'
```

> **Note:** Unrecognised cipher names (e.g. `BLALALA`) are skipped with a warning and do not cause an error.

#### 2. Verify endpoint pod logs

```bash
kubectl logs noobaa-endpoint-<suffix> -n <namespace> | grep "Updating TLS config for service"
```

Expected output:

```
Apr-30 12:42:23.447 [Endpoint/13]    [L0] core.util.ssl_utils:: Updating TLS config for service S3 config.TLS_MIN_VERSION=TLSv1.3, config.TLS_CIPHERS=ECDHE-RSA-AES128-GCM-SHA256:TLS_CHACHA20_POLY1305_SHA256 config.TLS_GROUPS=X25519
Apr-30 12:42:23.544 [Endpoint/13]    [L0] core.util.ssl_utils:: Updating TLS config for service STS config.TLS_MIN_VERSION=TLSv1.3, config.TLS_CIPHERS=ECDHE-RSA-AES128-GCM-SHA256:TLS_CHACHA20_POLY1305_SHA256 config.TLS_GROUPS=X25519
Apr-30 12:42:23.723 [Endpoint/13]    [L0] core.util.ssl_utils:: Updating TLS config for service IAM config.TLS_MIN_VERSION=TLSv1.3, config.TLS_CIPHERS=ECDHE-RSA-AES128-GCM-SHA256:TLS_CHACHA20_POLY1305_SHA256 config.TLS_GROUPS=X25519
Apr-30 12:42:23.855 [Endpoint/13]    [L0] core.util.ssl_utils:: Updating TLS config for service VECTOR config.TLS_MIN_VERSION=TLSv1.3, config.TLS_CIPHERS=ECDHE-RSA-AES128-GCM-SHA256:TLS_CHACHA20_POLY1305_SHA256 config.TLS_GROUPS=X25519
Apr-30 12:42:24.007 [Endpoint/13]    [L0] core.util.ssl_utils:: Updating TLS config for service METRICS config.TLS_MIN_VERSION=TLSv1.3, config.TLS_CIPHERS=ECDHE-RSA-AES128-GCM-SHA256:TLS_CHACHA20_POLY1305_SHA256 config.TLS_GROUPS=X25519
```

#### 3. Verify core pod logs

```bash
kubectl logs noobaa-core-0 -n <namespace> | grep "Updating TLS config for service"
```

Expected output:

```
Apr-30 12:42:39.418 [WebServer/34]    [L0] core.util.ssl_utils:: Updating TLS config for service MGMT config.TLS_MIN_VERSION=TLSv1.3, config.TLS_CIPHERS=ECDHE-RSA-AES128-GCM-SHA256:TLS_CHACHA20_POLY1305_SHA256 config.TLS_GROUPS=X25519
```

#### 4. Verify operator logs (admission server)

```bash
kubectl logs deploy/noobaa-operator -n <namespace> | grep TLS
```

Expected output:

```
time="..." level=info msg="Admission server TLS min version set to TLSv1.3"
time="..." level=info msg="MapCipherSuites: TLS config cipher suites TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256:TLS_CHACHA20_POLY1305_SHA256"
time="..." level=info msg="MapGroupPreferences: TLS group preferences set to X25519"
time="..." level=info msg="Admission server TLS configuration reloaded"
```

---

### Disabling TLS security config

Setting `DISABLE_TLS_SECURITY_CONFIG=true` on the operator deployment causes the operator to clear all TLS environment variables on the next reconciliation, reverting every configurable server to Node.js defaults. The operator pod, core pod, and endpoint pods all restart.

#### 1. Set the flag

```bash
kubectl set env deployment/noobaa-operator \
  DISABLE_TLS_SECURITY_CONFIG=true \
  -n <namespace>
```

#### 2. Watch pods restart

```
NAME                                READY   STATUS              RESTARTS   AGE
noobaa-core-0                       1/2     Running             0          11s
noobaa-operator-<suffix>            1/1     Running             0          17s
noobaa-core-0                       2/2     Running             0          22s
noobaa-endpoint-<new-suffix>        0/1     ContainerCreating   0          0s
noobaa-endpoint-<new-suffix>        1/1     Running             0          1s
```

#### 3. Verify core pod logs show empty TLS config

```bash
kubectl logs noobaa-core-0 -n <namespace> | grep "Updating TLS config for service"
```

Expected output (all values empty — Node.js defaults):

```
core.util.ssl_utils:: Updating TLS config for service MGMT config.TLS_MIN_VERSION=, config.TLS_CIPHERS= config.TLS_GROUPS=
```

#### 4. Verify endpoint pod logs show empty TLS config

```bash
kubectl logs noobaa-endpoint-<suffix> -n <namespace> | grep "Updating TLS config for service"
```

Expected output:

```
core.util.ssl_utils:: Updating TLS config for service S3 config.TLS_MIN_VERSION=, config.TLS_CIPHERS= config.TLS_GROUPS=
core.util.ssl_utils:: Updating TLS config for service STS config.TLS_MIN_VERSION=, config.TLS_CIPHERS= config.TLS_GROUPS=
core.util.ssl_utils:: Updating TLS config for service IAM config.TLS_MIN_VERSION=, config.TLS_CIPHERS= config.TLS_GROUPS=
core.util.ssl_utils:: Updating TLS config for service VECTOR config.TLS_MIN_VERSION=, config.TLS_CIPHERS= config.TLS_GROUPS=
core.util.ssl_utils:: Updating TLS config for service METRICS config.TLS_MIN_VERSION=, config.TLS_CIPHERS= config.TLS_GROUPS=
```

#### 5. Verify operator logs show the flag is active

```bash
kubectl logs deploy/noobaa-operator -n <namespace> | grep TLS
```

Expected output:

```
time="..." level=info msg="TLS security config disabled via DISABLE_TLS_SECURITY_CONFIG, using default TLS config for admission server"
time="..." level=info msg="Admission server TLS configuration reloaded"
```

---

### Re-enabling TLS security config

Remove the env var override to restore normal operator behaviour. On the next reconciliation the operator re-applies `apiServerSecurity` from the NooBaa CR to all pods.

```bash
kubectl set env deployment/noobaa-operator \
  DISABLE_TLS_SECURITY_CONFIG- \
  -n <namespace>
```

The operator, core, and endpoint pods restart and log the configured TLS values as shown in [Applying TLS config and verifying via pod logs](#applying-tls-config-and-verifying-via-pod-logs).

---

## NooBaa TLS Security Summary

The table below shows the TLS posture for all NooBaa endpoints as reported by a PQC readiness scan. Ports that carry HTTPS support configurable TLS via `spec.security.apiServerSecurity`.

```
================================================================
                 NooBaa TLS Security Summary
================================================================
ENDPOINT                         TLS 1.3      HYBRID KEY EXCHANGE
----------------------------------------------------------------
admission-webhook-service:443    YES          NO
iam:443                          YES          NO
noobaa-mgmt:443                  YES          NO
noobaa-mgmt:8445                 YES          YES
noobaa-mgmt:8446                 YES          YES
s3:443                           YES          NO
s3:8444                          YES          YES
s3:9443                          YES          NO
sts:443                          YES          NO
vectors:443                      YES          NO
noobaa-db-pg-cluster-r:5432      HTTP         N/A (plaintext)
noobaa-db-pg-cluster-ro:5432     HTTP         N/A (plaintext)
noobaa-db-pg-cluster-rw:5432     HTTP         N/A (plaintext)
noobaa-mgmt:80                   HTTP         N/A (plaintext)
noobaa-syslog:514                HTTP         N/A (plaintext)
s3:80                            HTTP         N/A (plaintext)
================================================================
Full results saved to noobaa_pqc_readiness_report.json
```

**Configurable endpoints** — all ports listed as `YES` for TLS 1.3 above honour the `tlsMinVersion`, `tlsCiphers`, and `tlsGroups` fields from the NooBaa CR. This includes `noobaa-mgmt:443`, which is externally accessible and was added to the configurable set alongside the other HTTPS ports.

**Hybrid key exchange** — ports marked `YES` in the HYBRID KEY EXCHANGE column already negotiate X25519MLKEM768 with supporting clients out of the box (Node.js v24 + OpenSSL 3.5+). Explicitly setting `tlsGroups: [X25519MLKEM768, ...]` in the CR forces this preference even on clusters that do not default to it.

---

### Using openssl

Verify TLS 1.3 with PQC group negotiation (requires OpenSSL 3.5+ with ML-KEM support):

```bash
echo Q | openssl s_client -connect <endpoint-address>:6443 -tls1_3 -groups X25519MLKEM768 2>&1 | grep "Negotiated TLS1.3 group: X25519MLKEM768"
```

A successful match confirms the endpoint is negotiating TLS 1.3 with the X25519MLKEM768 key exchange group.

> **Note:** The `echo Q |` pipes a quit command so `s_client` terminates cleanly after the handshake. Without it, the server may close the connection causing an `unexpected eof` error.

### Using testssl.sh

Run a TLS scan against the endpoint using the `testssl.sh` Docker image:

```bash
docker run --rm -it \
  -v <path-local>:<path-in-container> \
  ghcr.io/testssl/testssl.sh \
  --forward-secrecy --protocols --server-preference --client-simulation \
  --quiet --wide \
  --jsonfile <path-in-container>/testssl_output.json \
  <endpoint-address>
```

Key things to look for in the output:
- **Protocols**: Only TLS 1.3 should be listed as offered when `VersionTLS13` is set
- **Forward secrecy / key exchange groups**: Should reflect `tlsGroups` (e.g., `X25519MLKEM768`)
- **Server preference / ciphers**: Should match the configured `tlsCiphers`
- **Client simulation**: Shows how various clients negotiate with the server

Parse the JSON output to verify TLS 1.3 and PQC key exchange:

```bash
jq -r '["TLS1_3", "FS_KEMs"][] as $id | (map(select(.id == $id))[0] // {id: $id, severity: "not found", finding: ""}) | "\(.id): \(.severity) \(.finding)"' <path-local>/testssl_output.json
```

Expected output when TLS 1.3 with PQC is configured:

```
TLS1_3: OK offered
FS_KEMs: OK X25519MLKEM768
```


## TODO
- Add PQC scanner info

### Testing the admission webhook server

Pre-requisites - make sure that `ENABLE_NOOBAA_ADMISSION == "true"` or on CLI installation `--admission` in order to start the admission webhook server.

The admission webhook runs on port 8080 inside the operator pod.

#### Using openssl (with port-forward)

In one terminal, start a port-forward:

```bash
kubectl port-forward deploy/noobaa-operator -n openshift-storage 8080:8080
```

In another terminal, verify TLS 1.3 with PQC group negotiation:

```bash
echo Q | openssl s_client -connect localhost:8080 -tls1_3 -groups X25519MLKEM768 2>&1 | grep "Negotiated TLS1.3 group: X25519MLKEM768"
```

#### Using testssl.sh (from within the cluster)

Run testssl.sh as a pod inside the cluster to connect directly to the operator service without port-forwarding:

```bash
kubectl run testssl-webhook -it --restart=Never \
  --image=ghcr.io/testssl/testssl.sh -- \
  --forward-secrecy --protocols --server-preference --client-simulation \
  --quiet --wide \
  https://noobaa-operator.openshift-storage.svc.cluster.local:8080
```

#### Using testssl.sh with JSON output (from within the cluster)

```bash
kubectl run testssl-webhook -it --restart=Never \
  --image=ghcr.io/testssl/testssl.sh -- \
  --forward-secrecy --protocols --server-preference --client-simulation \
  --quiet --wide \
  --jsonfile /tmp/testssl_webhook_output.json \
  https://noobaa-operator.openshift-storage.svc.cluster.local:8080
```

To extract the JSON results, copy them from the pod before it terminates, or parse the console output directly.
