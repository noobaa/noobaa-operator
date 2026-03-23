[NooBaa Operator](../README.md) /
# TLS PQC Readiness Check

## Overview

The TLS PQC readiness check verifies that NooBaa services support **TLS 1.3** and **post-quantum cryptography (PQC) hybrid key exchange** (X25519MLKEM768).

It dynamically discovers all NooBaa services (label `app=noobaa`) and their ports. Ports with names containing "https" are scanned with testssl.sh, all others are reported as plaintext HTTP.

The following services and ports are typically discovered:

| Service | Port | Name | Protocol |
|---------|------|------|----------|
| noobaa-mgmt | 443 | mgmt-https | HTTPS (scanned) |
| noobaa-mgmt | 8445 | bg-https | HTTPS (scanned) |
| noobaa-mgmt | 8446 | hosted-agents-https | HTTPS (scanned) |
| s3 | 443 | s3-https | HTTPS (scanned) |
| s3 | 8444 | md-https | HTTPS (scanned) |
| s3 | 9443 | metrics-https | HTTPS (scanned) |
| sts | 443 | sts-https | HTTPS (scanned) |
| iam | 443 | iam-https | HTTPS (scanned) |
| noobaa-db-pg-cluster-r | 5432 | postgresql | HTTP (reported only) |
| noobaa-db-pg-cluster-ro | 5432 | postgresql | HTTP (reported only) |
| noobaa-db-pg-cluster-rw | 5432 | postgresql | HTTP (reported only) |
| noobaa-mgmt | 80 | mgmt | HTTP (reported only) |
| s3 | 80 | s3 | HTTP (reported only) |
| noobaa-syslog | 514 | syslog | HTTP (reported only) |

For each HTTPS endpoint, a [testssl.sh](https://github.com/testssl/testssl.sh) pod is launched inside the cluster to probe the port. The results are collected as JSON, merged into a single report, and a readiness summary is appended.

## testssl.sh Command

The scanner pod runs the following testssl.sh command against each HTTPS endpoint:

```bash
testssl.sh --forward-secrecy --protocols --server-preference --client-simulation --quiet --wide --jsonfile /tmp/res.json https://<service>.<namespace>.svc.cluster.local:<port>
```

| Flag | Description |
|------|-------------|
| `--protocols` | Tests which TLS/SSL protocols are supported (SSLv2, SSLv3, TLS 1.0–1.3) |
| `--forward-secrecy` | Checks forward secrecy support and lists key exchange curves, including PQC hybrid groups like X25519MLKEM768 |
| `--server-preference` | Checks whether the server enforces its own cipher order preference |
| `--client-simulation` | Simulates connections from common TLS clients to show which protocol and cipher each would negotiate |
| `--quiet` | Suppresses banner and other non-essential output |
| `--wide` | Uses wider output format for more detailed cipher and protocol information |
| `--jsonfile /tmp/res.json` | Writes structured results in JSON format for programmatic analysis |

## Prerequisites

- `kubectl` (or `oc`) with access to the target cluster
- `jq` installed locally
- NooBaa deployed in the target namespace

## Running the Check

### Via Make

```bash
make test-tls-pqc-readiness
```

### Directly

```bash
bash test/tls_pqc_readiness_check.sh
```

### With Custom Namespace

```bash
NAMESPACE=openshift-storage bash test/tls_pqc_readiness_check.sh
```

### On OpenShift (using oc)

```bash
KUBECTL=oc NAMESPACE=openshift-storage bash test/tls_pqc_readiness_check.sh
```

## Configuration

All configuration is via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `NAMESPACE` | `test` | Kubernetes namespace where NooBaa is deployed |
| `SCAN_TIMEOUT` | `600` | Maximum seconds to wait for each scan pod to complete |
| `KUBECTL` | `kubectl` | Command to interact with the cluster (`kubectl` or `oc`) |

## Output

### Console

A summary table is printed to stdout:

```
================================================================
                 NooBaa TLS Security Summary
================================================================
ENDPOINT                         TLS 1.3      HYBRID KEY EXCHANGE
----------------------------------------------------------------
noobaa-mgmt:443                  YES          YES
noobaa-mgmt:8445                 YES          YES
noobaa-mgmt:8446                 YES          YES
s3:443                           YES          YES
s3:8444                          YES          YES
s3:9443                          YES          YES
sts:443                          YES          YES
iam:443                          YES          YES
noobaa-db-pg-cluster-r:5432      HTTP         N/A (plaintext)
noobaa-db-pg-cluster-ro:5432     HTTP         N/A (plaintext)
noobaa-db-pg-cluster-rw:5432     HTTP         N/A (plaintext)
noobaa-mgmt:80                   HTTP         N/A (plaintext)
s3:80                            HTTP         N/A (plaintext)
noobaa-syslog:514                HTTP         N/A (plaintext)
================================================================
Full results saved to noobaa_pqc_readiness_report.json

RESULT: PASS - All endpoints support TLS 1.3 and hybrid (PQC) key exchange
```

### JSON Report

The file `noobaa_pqc_readiness_report.json` is written to the current working directory. It contains:

1. **Raw testssl results** — the full protocol and forward-secrecy scan output for each service
2. **Readiness summary** — appended as the last entry:

```json
{
  "id": "pqc_readiness_summary",
  "result": "PASS",
  "endpoints": [
    { "service": "noobaa-mgmt", "port": 443, "tls_1_3": "YES", "hybrid_key_exchange": "YES" },
    { "service": "noobaa-mgmt", "port": 8445, "tls_1_3": "YES", "hybrid_key_exchange": "YES" },
    { "service": "noobaa-mgmt", "port": 8446, "tls_1_3": "YES", "hybrid_key_exchange": "YES" },
    { "service": "s3", "port": 443, "tls_1_3": "YES", "hybrid_key_exchange": "YES" },
    { "service": "s3", "port": 8444, "tls_1_3": "YES", "hybrid_key_exchange": "YES" },
    { "service": "s3", "port": 9443, "tls_1_3": "YES", "hybrid_key_exchange": "YES" },
    { "service": "sts", "port": 443, "tls_1_3": "YES", "hybrid_key_exchange": "YES" },
    { "service": "iam", "port": 443, "tls_1_3": "YES", "hybrid_key_exchange": "YES" },
    { "service": "noobaa-db-pg-cluster-r", "port": 5432, "tls_1_3": "HTTP", "hybrid_key_exchange": "N/A" },
    { "service": "noobaa-db-pg-cluster-ro", "port": 5432, "tls_1_3": "HTTP", "hybrid_key_exchange": "N/A" },
    { "service": "noobaa-db-pg-cluster-rw", "port": 5432, "tls_1_3": "HTTP", "hybrid_key_exchange": "N/A" },
    { "service": "noobaa-mgmt", "port": 80, "tls_1_3": "HTTP", "hybrid_key_exchange": "N/A" },
    { "service": "s3", "port": 80, "tls_1_3": "HTTP", "hybrid_key_exchange": "N/A" },
    { "service": "noobaa-syslog", "port": 514, "tls_1_3": "HTTP", "hybrid_key_exchange": "N/A" }
  ]
}
```

### Field Descriptions

| Field | Value | Meaning |
|-------|-------|---------|
| `tls_1_3` | `YES` | The endpoint advertises TLS 1.3 as a supported protocol. This is required for PQC key exchange. |
| `tls_1_3` | `NO` | The endpoint does not support TLS 1.3. PQC key exchange will not be available. |
| `tls_1_3` | `HTTP` | The port is plaintext HTTP — TLS is not applicable. |
| `hybrid_key_exchange` | `YES` | The endpoint supports hybrid post-quantum key exchange (X25519MLKEM768), combining classical X25519 ECDH with ML-KEM-768 post-quantum KEM. Connections to this endpoint are protected against both classical and future quantum attacks. |
| `hybrid_key_exchange` | `NO` | The endpoint does not offer any PQC hybrid key exchange group. Connections use classical-only key exchange. |
| `hybrid_key_exchange` | `N/A` | The port is plaintext HTTP — key exchange is not applicable. |
| `result` | `PASS` | All HTTPS endpoints support both TLS 1.3 and hybrid PQC key exchange. |
| `result` | `FAIL` | One or more HTTPS endpoints are missing TLS 1.3 or hybrid PQC key exchange support. |

### Exit Codes

| Code | Meaning |
|------|---------|
| `0` | All endpoints support TLS 1.3 and hybrid key exchange |
| `1` | One or more endpoints failed the check |

## CI Integration

The check runs automatically via the **Run TLS PQC Test** GitHub Actions workflow (`.github/workflows/run_tls_pqc_test.yml`). The workflow:

1. Sets up a Minikube cluster
2. Builds and installs the NooBaa operator
3. Installs NooBaa into the `test` namespace
4. Runs `make test-tls-pqc-readiness`
5. Uploads `noobaa_pqc_readiness_report.json` as a workflow artifact

## Background

Post-quantum cryptography protects TLS connections against future quantum computer attacks. The hybrid key exchange X25519MLKEM768 combines the classical X25519 ECDH with the ML-KEM-768 post-quantum KEM, providing security against both classical and quantum adversaries. TLS 1.3 is required for PQC key exchange support.
