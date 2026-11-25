#!/bin/bash
#
# TLS PQC Readiness Check
#
# Dynamically discovers all NooBaa services (label app=noobaa) and their
# ports, then scans every HTTPS port for TLS 1.3 support and post-quantum
# cryptography (PQC) hybrid key exchange (X25519MLKEM768). HTTP ports are
# reported as plaintext but not scanned.
#
# Ports are classified as HTTPS when the port name contains "https", "ssl",
# "tls", or "secure", OR the port number is 443, OR the appProtocol is
# "https". All other ports are reported as HTTP/plaintext.
#
# All HTTPS endpoints are scanned concurrently: testssl.sh pods are launched
# in-cluster for every port at once, then the script waits for all pods to
# finish before collecting results. Results are merged into a single JSON
# report file with a per-endpoint pass/fail summary.
#
# Prerequisites:
#   - kubectl (or oc) access to the target cluster
#   - jq installed locally
#   - The target namespace must have NooBaa deployed
#
# Environment variables:
#   NAMESPACE     - Kubernetes namespace where NooBaa is deployed (default: test)
#   SCAN_TIMEOUT  - Max seconds to wait for each scan pod (default: 600)
#   KUBECTL       - kubectl or oc command to use (default: kubectl)
#
# Usage:
#   bash test/tls_pqc_readiness_check.sh
#   NAMESPACE=openshift-storage bash test/tls_pqc_readiness_check.sh
#   KUBECTL=oc NAMESPACE=openshift-storage bash test/tls_pqc_readiness_check.sh
#   make test-tls-pqc-readiness
#
# Output:
#   noobaa_pqc_readiness_report.json  - Full testssl results + pqc_readiness_summary
#   stdout                       - Per-endpoint summary table
#
# Exit codes:
#   0 - All endpoints support TLS 1.3 and hybrid key exchange
#   1 - One or more endpoints failed the check
#

NAMESPACE="${NAMESPACE:-test}"
OUTPUT_FILE="noobaa_pqc_readiness_report.json"
SCAN_TIMEOUT="${SCAN_TIMEOUT:-600}"
KUBECTL="${KUBECTL:-kubectl}"

echo "[]" > "$OUTPUT_FILE"

# Discover all ports from NooBaa services dynamically
# Output format per line: service:port:portname:appprotocol
# we may need to extend this check if there are cases of non-standard HTTPS ports without indicative names or appProtocol,
# but this should cover most common scenarios and expecially user facing services.
echo "Discovering NooBaa service ports in namespace $NAMESPACE..."
ALL_PORTS=$($KUBECTL get svc -n "$NAMESPACE" -l app=noobaa -o json | \
  jq -r '.items[] | .metadata.name as $svc | .spec.ports[] | "\($svc):\(.port):\(.name // ""):\(.appProtocol // "")"')

if [ -z "$ALL_PORTS" ]; then
  echo "ERROR: No NooBaa services found in namespace $NAMESPACE (label app=noobaa)" >&2
  exit 1
fi

ENDPOINTS=""
HTTP_ENDPOINTS=""
for ENTRY in $ALL_PORTS; do
  SVC=$(echo "$ENTRY" | cut -d: -f1)
  PORT=$(echo "$ENTRY" | cut -d: -f2)
  PORT_NAME=$(echo "$ENTRY" | cut -d: -f3)
  APP_PROTO=$(echo "$ENTRY" | cut -d: -f4)
  if echo "$PORT_NAME" | grep -qiE "https|ssl|tls|secure"; then
    ENDPOINTS="$ENDPOINTS $SVC:$PORT"
  elif [ "$PORT" = "443" ]; then
    ENDPOINTS="$ENDPOINTS $SVC:$PORT"
  elif echo "$APP_PROTO" | grep -qi "https"; then
    ENDPOINTS="$ENDPOINTS $SVC:$PORT"
  else
    HTTP_ENDPOINTS="$HTTP_ENDPOINTS $SVC:$PORT"
  fi
done

echo "  HTTPS endpoints:$ENDPOINTS"
echo "  HTTP endpoints:$HTTP_ENDPOINTS"
echo ""

SUMMARY=""
SUMMARY_JSON=""
ALL_PASS=true

# --- Phase 1: Launch all scanner pods concurrently ---
echo "Launching scanner pods for all HTTPS endpoints..."
for ENDPOINT in $ENDPOINTS; do
  SVC="${ENDPOINT%%:*}"
  PORT="${ENDPOINT##*:}"
  POD_NAME="testssl-scanner-${SVC}-${PORT}"

  $KUBECTL delete pod "$POD_NAME" -n "$NAMESPACE" --ignore-not-found 2>/dev/null

  $KUBECTL run "$POD_NAME" -n "$NAMESPACE" --restart=Never \
    --labels="app=noobaa" \
    --image=ghcr.io/testssl/testssl.sh \
    --command -- sh -c \
    "testssl.sh --forward-secrecy --protocols --server-preference --client-simulation --quiet --wide --jsonfile /tmp/res.json --quiet https://${SVC}.${NAMESPACE}.svc.cluster.local:${PORT} > /dev/null 2>&1; cat /tmp/res.json 2>/dev/null || echo '[]'"

  echo "  Started pod $POD_NAME for $SVC:$PORT"
done

# --- Phase 2: Wait for all scanner pods to complete ---
echo ""
echo "Waiting for all scanner pods to finish (timeout: ${SCAN_TIMEOUT}s)..."
ELAPSED=0
while [ $ELAPSED -lt $SCAN_TIMEOUT ]; do
  ALL_DONE=true
  for ENDPOINT in $ENDPOINTS; do
    SVC="${ENDPOINT%%:*}"
    PORT="${ENDPOINT##*:}"
    POD_NAME="testssl-scanner-${SVC}-${PORT}"
    PHASE=$($KUBECTL get pod "$POD_NAME" -n "$NAMESPACE" -o jsonpath='{.status.phase}' 2>/dev/null)
    case "$PHASE" in
      Succeeded|Failed) ;;
      *) ALL_DONE=false ;;
    esac
  done
  if [ "$ALL_DONE" = true ]; then
    echo "  All pods completed after ${ELAPSED}s"
    break
  fi
  sleep 10
  ELAPSED=$((ELAPSED + 10))

  STILL_RUNNING=""
  for ENDPOINT in $ENDPOINTS; do
    SVC="${ENDPOINT%%:*}"
    PORT="${ENDPOINT##*:}"
    POD_NAME="testssl-scanner-${SVC}-${PORT}"
    PHASE=$($KUBECTL get pod "$POD_NAME" -n "$NAMESPACE" -o jsonpath='{.status.phase}' 2>/dev/null)
    case "$PHASE" in
      Succeeded|Failed) ;;
      *) STILL_RUNNING="$STILL_RUNNING $SVC:$PORT" ;;
    esac
  done
  echo "  ${ELAPSED}s elapsed - still running:${STILL_RUNNING}"
done

# --- Phase 3: Collect results from all pods ---
echo ""
echo "Collecting results..."
for ENDPOINT in $ENDPOINTS; do
  SVC="${ENDPOINT%%:*}"
  PORT="${ENDPOINT##*:}"
  POD_NAME="testssl-scanner-${SVC}-${PORT}"

  EP_TLS13="NO_DATA"
  EP_HYBRID="NO_DATA"

  PHASE=$($KUBECTL get pod "$POD_NAME" -n "$NAMESPACE" -o jsonpath='{.status.phase}' 2>/dev/null)

  if [ "$PHASE" = "Succeeded" ] || [ "$PHASE" = "Failed" ]; then
    RAW_OUTPUT=$($KUBECTL logs "$POD_NAME" -n "$NAMESPACE" 2>/dev/null)
    echo "$RAW_OUTPUT" | sed -n '/^\[/,/^\]/p' > "temp_${SVC}_${PORT}.json"

    if [ -s "temp_${SVC}_${PORT}.json" ] && jq empty "temp_${SVC}_${PORT}.json" 2>/dev/null; then
      echo "  Got results for $SVC:$PORT"
      jq -s 'add' "$OUTPUT_FILE" "temp_${SVC}_${PORT}.json" > "${OUTPUT_FILE}.tmp" && mv "${OUTPUT_FILE}.tmp" "$OUTPUT_FILE"

      if jq -e '[.[] | select(.id | test("tls1_3"; "i")) | select(.finding | test("offered"))] | length > 0' "temp_${SVC}_${PORT}.json" > /dev/null 2>&1; then
        EP_TLS13="YES"
      else
        EP_TLS13="NO"
      fi

      if jq -e '[.[] | select(.finding | test("X25519MLKEM|MLKEM"; "i"))] | length > 0' "temp_${SVC}_${PORT}.json" > /dev/null 2>&1; then
        EP_HYBRID="YES"
      else
        EP_HYBRID="NO"
      fi
    else
      echo "  WARNING: No valid JSON results for $SVC:$PORT (pod phase: $PHASE)"
      if [ "$PHASE" = "Failed" ]; then
        echo "  Pod logs:"
        echo "$RAW_OUTPUT" | head -20
      fi
    fi
    rm -f "temp_${SVC}_${PORT}.json"
  else
    echo "  ERROR: Scan timed out for $SVC:$PORT after ${SCAN_TIMEOUT}s"
  fi

  SUMMARY="${SUMMARY}$(printf '%-32s %-12s %-20s\n' "$SVC:$PORT" "$EP_TLS13" "$EP_HYBRID")\n"
  SUMMARY_JSON="${SUMMARY_JSON:+$SUMMARY_JSON,}{\"service\":\"$SVC\",\"port\":$PORT,\"tls_1_3\":\"$EP_TLS13\",\"hybrid_key_exchange\":\"$EP_HYBRID\"}"
  if [ "$EP_TLS13" != "YES" ] || [ "$EP_HYBRID" != "YES" ]; then
    ALL_PASS=false
  fi

  $KUBECTL delete pod "$POD_NAME" -n "$NAMESPACE" --ignore-not-found 2>/dev/null
done

# Add HTTP endpoints to the summary (not scanned)
for ENDPOINT in $HTTP_ENDPOINTS; do
  SVC="${ENDPOINT%%:*}"
  PORT="${ENDPOINT##*:}"
  SUMMARY="${SUMMARY}$(printf '%-32s %-12s %-20s\n' "$SVC:$PORT" "HTTP" "N/A (plaintext)")\n"
  SUMMARY_JSON="${SUMMARY_JSON:+$SUMMARY_JSON,}{\"service\":\"$SVC\",\"port\":$PORT,\"tls_1_3\":\"HTTP\",\"hybrid_key_exchange\":\"N/A\"}"
done

echo ""
echo "================================================================"
echo "                 NooBaa TLS Security Summary"
echo "================================================================"
printf "%-32s %-12s %-20s\n" "ENDPOINT" "TLS 1.3" "HYBRID KEY EXCHANGE"
echo "----------------------------------------------------------------"
echo -e "$SUMMARY"
echo "================================================================"
RESULT_TEXT="PASS"
if [ "$ALL_PASS" != true ]; then
  RESULT_TEXT="FAIL"
fi

jq --argjson summary "[${SUMMARY_JSON}]" --arg result "$RESULT_TEXT" \
  '. + [{"id":"pqc_readiness_summary","result":$result,"endpoints":$summary}]' \
  "$OUTPUT_FILE" > "${OUTPUT_FILE}.tmp" && mv "${OUTPUT_FILE}.tmp" "$OUTPUT_FILE"

echo "Full results saved to $OUTPUT_FILE"
echo ""
if [ "$ALL_PASS" = true ]; then
  echo "RESULT: PASS - All endpoints support TLS 1.3 and hybrid (PQC) key exchange"
  exit 0
else
  echo "RESULT: FAIL - One or more endpoints missing TLS 1.3 or hybrid key exchange" >&2
  exit 1
fi
