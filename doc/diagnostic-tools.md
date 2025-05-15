[NooBaa Operator](../README.md) /
# Diagnostic Tools
This document explains and elaborates on the diagnostic tools provided by the NooBaa CLI when the `diagnostics` parameter is used. These tools are designed to help the user get a comprehensive overview of their system's status, as well as troubleshooting problems.

## Log Analysis
The log analyzer (`log-analysis`) lets the user search for strings in the logs of all NooBaa pods and containers, including `--previous` ones. It provides a clear and easy-to-use report with the number of times the searched string was found, alongside the first and last occurrence timestamps to provide an idea of the time window that the searched string applies to.
The first (and only) positional parameter expects to receive the search string; if no positional parameter is provided, the tool will prompt the user to enter the search string.

The analyzer supports several options to allow for more advanced queries:
- `-i, --case-insensitive` - Allows the user to search for a string in a case-insensitive manner (similar to `grep -i`)
- `--fuzzy` - Activates fuzzy-matching, finding matches for any cases where the search string is a subset of a log line (even if not sequentially; e.g. when used with the -i parameter, “s3error” would find “s3 error”, “S3Error”, and “s3-error”)
- `--noobaa-time` - By default, the tool uses Kubernetes timestamps, since in some cases, certain NooBaa log lines may not have timestamps. The NooBaa timestamps are at an offset of 5~20ms from the Kubernetes ones, thus it is recommended to use one type of timestamp consistently.
- `--prefer-line` - When used in conjunction with `--noobaa-time`, will print a log line instead of an error when no timestamp is present in the line
- `--tail` - Tells the tool to fetch only the last N lines of the logs (similar to `minikube logs --tail N`)
- `-v, --verbose` - Prints every matching log line
- `-w, --whole-string` - Activates whole-string matching, yielding results only if the search string appears in a log line while surrounded by non-word characters; e.g. "Reconcile" will match with "BucketClass Reconcile..." but not "ReconcileObject" (similar to `grep -w`)

### Example:
In this example, the tool is used to find case-insensitive, fuzzy occurrences of "s3error" in the last 1,000 lines - 
```bash
$ nb diagnostics log-analysis --tail 1000 -i --fuzzy s3error
INFO[0000]
INFO[0000] ✨────────────────────────────────────────────✨
INFO[0000]    Collecting and analyzing pod logs -
INFO[0000]    Search string: s3error
INFO[0000]    Case insensitivity: true
INFO[0000]    Match whole string: false
INFO[0000]    From the last 1000 lines
INFO[0000]    Using Kubernetes timestamps
INFO[0000]    Found occurrences will be printed below
INFO[0000]    in the format <pod name>:<container name>
INFO[0000] ✨────────────────────────────────────────────✨
INFO[0000] Analyzing noobaa-core-0:core
INFO[0001] Hits: 411
INFO[0001] Earliest appearance: 2024-12-11T09:04:05.245904744Z
INFO[0001] Latest appearance:   2024-12-11T09:11:51.054808004Z
INFO[0001] ──────────────────────────────────────────────────────────────────────────────────
INFO[0001] Analyzing noobaa-core-0:noobaa-log-processor
INFO[0001] No occurrences found
INFO[0001] ──────────────────────────────────────────────────────────────────────────────────
INFO[0001] Analyzing noobaa-db-pg-0:db
INFO[0001] No occurrences found
INFO[0001] ──────────────────────────────────────────────────────────────────────────────────
INFO[0001] Analyzing noobaa-default-backing-store-noobaa-pod-08270bcd:noobaa-agent
INFO[0001] Hits: 111
INFO[0001] Earliest appearance: 2024-12-11T08:36:42.050234365Z
INFO[0001] Latest appearance:   2024-12-11T09:11:20.498970755Z
INFO[0001] ──────────────────────────────────────────────────────────────────────────────────
INFO[0001] Analyzing noobaa-endpoint-7995576ddc-tq9p4:endpoint
INFO[0001] Hits: 6
INFO[0001] Earliest appearance: 2024-12-11T08:32:04.879477461Z
INFO[0001] Latest appearance:   2024-12-11T08:36:27.793515138Z
INFO[0001] ──────────────────────────────────────────────────────────────────────────────────
INFO[0001] Analyzing noobaa-operator-68f66b987b-dz4xw:noobaa-operator
INFO[0001] Hits: 10
INFO[0001] Earliest appearance: 2024-12-11T09:10:43.972246776Z
INFO[0001] Latest appearance:   2024-12-11T09:11:50.359484454Z
INFO[0001] ──────────────────────────────────────────────────────────────────────────────────
```

## Resource Analysis
TODO

## Diagnostic Information Collection
TODO

## System Reports
TODO