# NooBaa Containerized - Monitoring

1. [Introduction](#introduction)
2. [Metrics Endpoint Configuration](#metrics-endpoint-configuration)
3. [Metrics Description](#metrics-description)
4. [Getting Started](#getting-started)
5. [Prometheus Alert Rules](#prometheus-alert-rules)
6. [Examples](#examples)

## Introduction

NooBaa exposes Prometheus metrics for monitoring system health, web server activity, and background workers. </br>
This document focuses on containerized deployments and highlights how to access metrics endpoints, scrape them with Prometheus, and verify outputs.

## Metrics Endpoint Configuration

This section provides details about the metrics URL and port configuration used by the NooBaa Operator services.

#### Prometheus Metrics HTTP URLs - </br>
- **Core metrics (aggregated)** are exposed on the S3 service metrics port: </br> `http://<s3-service>:7004/`
- **Web server metrics** are exposed on the management service: </br> `http://<mgmt-service>/metrics/web_server`
- **Background workers metrics** are exposed on the management service: </br> `http://<mgmt-service>/metrics/bg_workers`
- **Hosted agents metrics** are exposed on the management service: </br> `http://<mgmt-service>/metrics/hosted_agents`

#### Default services and ports - </br>
- S3 service name: `s3`, metrics port: `7004` (metrics path is `/`)
- Management service name: `<noobaa-name>-mgmt`, port: `80` (targets container port `8080`)

The operator creates ServiceMonitors for both services and injects Bearer token authorization automatically.

## Metrics Description

NooBaa exposes Prometheus metrics for multiple components.
NooBaa core metrics are prefixed by `NooBaa_` and endpoint metrics by `NooBaa_Endpoint_` (default `PROMETHEUS_PREFIX`). </br>
Core metrics are exposed at the S3 service metrics endpoint (`/` on port `7004`), while `/metrics/web_server`, `/metrics/bg_workers`, and `/metrics/hosted_agents` are exposed on the management service.

### NooBaa Core Metrics (`/` on port `7004`)

| Metric | What it shows | Labels |
|--------|----------------|--------|
| NooBaa_cloud_types | Cloud Resource Types in the System | `type` |
| NooBaa_projects_capacity_usage | Projects Capacity Usage | `project` |
| NooBaa_accounts_usage_read_count | Accounts Usage Read Count | `account` |
| NooBaa_accounts_usage_write_count | Accounts Usage Write Count | `account` |
| NooBaa_accounts_usage_logical | Accounts Usage Logical | `account` |
| NooBaa_bucket_class_capacity_usage | Bucket Class Capacity Usage | `bucket_class` |
| NooBaa_unhealthy_cloud_types | Unhealthy Cloud Resource Types in the System | `type` |
| NooBaa_object_histo | Object Sizes Histogram Across the System | `size`, `avg` |
| NooBaa_providers_bandwidth_write_size_total | Providers bandwidth write size | `type` |
| NooBaa_providers_bandwidth_read_size_total | Providers bandwidth read size | `type` |
| NooBaa_providers_ops_read_count | Providers number of read operations | `type` |
| NooBaa_providers_ops_write_count | Providers number of write operations | `type` |
| NooBaa_providers_physical_size | Providers Physical Stats | `type` |
| NooBaa_providers_logical_size | Providers Logical Stats | `type` |
| NooBaa_system_capacity | System capacity | - |
| NooBaa_system_info | System info | `system_name`, `system_address` |
| NooBaa_num_buckets | Object Buckets | - |
| NooBaa_num_namespace_buckets | Namespace Buckets | - |
| NooBaa_total_usage | Total Usage | - |
| NooBaa_accounts_num | Accounts Number | - |
| NooBaa_num_objects | Objects | - |
| NooBaa_num_unhealthy_buckets | Unhealthy Buckets | - |
| NooBaa_num_unhealthy_namespace_buckets | Unhealthy Namespace Buckets | - |
| NooBaa_num_unhealthy_pools | Unhealthy Resource Pools | - |
| NooBaa_num_unhealthy_namespace_resources | Unhealthy Namespace Resources | - |
| NooBaa_num_pools | Resource Pools | - |
| NooBaa_num_namespace_resources | Namespace Resources | - |
| NooBaa_num_unhealthy_bucket_claims | Unhealthy Bucket Claims | - |
| NooBaa_num_buckets_claims | Object Bucket Claims | - |
| NooBaa_num_objects_buckets_claims | Objects On Object Bucket Claims | - |
| NooBaa_reduction_ratio | Object Efficiency Ratio | - |
| NooBaa_object_savings_logical_size | Object Savings Logical | - |
| NooBaa_object_savings_physical_size | Object Savings Physical | - |
| NooBaa_rebuild_progress | Rebuild Progress | - |
| NooBaa_rebuild_time | Rebuild Time | - |
| NooBaa_bucket_status | Bucket Health | `bucket_name`, `bucket_mode` |
| NooBaa_namespace_bucket_status | Namespace Bucket Health | `bucket_name`, `bucket_mode` |
| NooBaa_bucket_tagging | Bucket Tagging | `bucket_name`, `tagging` |
| NooBaa_namespace_bucket_tagging | Namespace Bucket Tagging | `bucket_name`, `tagging` |
| NooBaa_bucket_capacity | Bucket Capacity Percent | `bucket_name` |
| NooBaa_backing_store_low_capacity | Backing Store Low Capacity Status (0=normal, 1=low capacity) | `backing_store_name` |
| NooBaa_bucket_size_quota | Bucket Size Quota Percent | `bucket_name` |
| NooBaa_bucket_quantity_quota | Bucket Quantity Quota Percent | `bucket_name` |
| NooBaa_bucket_object_count | Current Number of Objects per Bucket | `bucket_name` |
| NooBaa_bucket_max_objects_quota | Bucket Maximum Objects Quota | `bucket_name` |
| NooBaa_bucket_max_bytes_quota | Bucket Maximum Bytes Quota | `bucket_name` |
| NooBaa_resource_status | Resource Health | `resource_name` |
| NooBaa_namespace_resource_status | Namespace Resource Health | `namespace_resource_name` |
| NooBaa_system_links | System Links | `resources`, `buckets`, `dashboard` |
| NooBaa_health_status | Health status | - |
| NooBaa_odf_health_status | Health status | - |
| NooBaa_replication_status | Replication status | `replication_id`, `bucket_name`, `last_cycle_rule_id`, `last_cycle_src_cont_token` |
| NooBaa_replication_last_cycle_writes_size | Bytes replicated by replication_id in last cycle | `replication_id` |
| NooBaa_replication_last_cycle_writes_num | Objects replicated by replication_id in last cycle | `replication_id` |
| NooBaa_replication_last_cycle_error_writes_size | Error bytes by replication_id in last cycle | `replication_id` |
| NooBaa_replication_last_cycle_error_writes_num | Error objects by replication_id in last cycle | `replication_id` |
| NooBaa_bucket_last_cycle_total_objects_num | Total objects scanned per bucket in last cycle | `bucket_name` |
| NooBaa_bucket_last_cycle_replicated_objects_num | Objects replicated per bucket in last cycle | `bucket_name` |
| NooBaa_bucket_last_cycle_error_objects_num | Objects failed to replicate per bucket in last cycle | `bucket_name` |
| NooBaa_bucket_used_bytes | Object Bucket Used Bytes | `bucket_name` |
| NooBaa_replication_target_status | Replication target bucket reachability (1=reachable, 0=unreachable) | `source_bucket`, `target_bucket` |

### NooBaa Endpoint Metrics (`/metrics/*` on port `8080`)

| Metric | What it shows | Labels |
|--------|----------------|--------|
| NooBaa_Endpoint_hub_read_bytes | Hub read bytes in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_hub_write_bytes | Hub write bytes in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_cache_read_bytes | Cache read bytes in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_cache_write_bytes | Cache write bytes in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_cache_object_read_count | Entire object reads in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_cache_object_read_miss_count | Entire object read miss in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_cache_object_read_hit_count | Entire object read hit in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_cache_range_read_count | Range reads in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_cache_range_read_miss_count | Range read miss in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_cache_range_read_hit_count | Range read hit in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_hub_read_latency | Hub read latency in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_hub_write_latency | Hub write latency in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_cache_read_latency | Cache read latency in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_cache_write_latency | Cache write latency in namespace cache bucket | `bucket_name` |
| NooBaa_Endpoint_semaphore_waiting_value | Namespace semaphore waiting value | `type`, `average_interval` |
| NooBaa_Endpoint_semaphore_waiting_time | Namespace semaphore waiting time | `type`, `average_interval` |
| NooBaa_Endpoint_semaphore_waiting_queue | Namespace semaphore waiting queue size | `type`, `average_interval` |
| NooBaa_Endpoint_semaphore_value | Namespace semaphore value | `type`, `average_interval` |
| NooBaa_Endpoint_fork_counter | Number of fork hits | `code` |

#### Endpoint metrics discovery
The web_server, bg_workers, and hosted_agents endpoints export a large and evolving set of runtime metrics. If `NOOBAA_METRICS_AUTH_ENABLED=true`, ensure the token is set (see [Set JWT Token](#3-set-jwt-token)) and use either the port-forward from step 4 or exec from step 5, then run these queries. If `NOOBAA_METRICS_AUTH_ENABLED=false`, the Authorization header is not required.

```sh
# Web server metrics
curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://127.0.0.1:8080/metrics/web_server | head

# Background workers metrics
curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://127.0.0.1:8080/metrics/bg_workers | head

# Hosted agents metrics
curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://127.0.0.1:8080/metrics/hosted_agents | head
```


## Getting Started

This section will walk you through the initial steps required to access metrics in a containerized deployment.

#### 1. Verify NooBaa is running </br>
Ensure the NooBaa core and endpoint containers are up and ready in your namespace.

#### 2. Generate activity
Run a few S3 or IAM operations (create bucket, upload objects, list objects) to generate metrics.

#### 3. Set JWT Token
If `NOOBAA_METRICS_AUTH_ENABLED=true` (default), export the bearer token once and reuse it in the following steps. If it is `false`, skip this step and omit the Authorization header in the curl examples.

```sh
JWT_TOKEN=$(kubectl get secret/<noobaa-name>-metrics-auth-secret -o jsonpath={.data.metrics_token} | base64 -d)
```

#### 4. Fetch metrics (port-forward + local curl)
You can access metrics by port-forwarding the services and using local `curl`.

```sh
# Keep these port-forward commands running in separate terminal windows or tabs.
kubectl -n <namespace> port-forward svc/s3 7004:7004
kubectl -n <namespace> port-forward svc/<noobaa-name>-mgmt 8080:80

curl -s -H "Authorization: Bearer ${JWT_TOKEN}" http://127.0.0.1:7004/ | head
curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://127.0.0.1:8080/metrics/web_server | head
curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://127.0.0.1:8080/metrics/bg_workers | head
curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://127.0.0.1:8080/metrics/hosted_agents | head
```

#### 5. Fetch metrics from inside the core pod (no port-forward)
If you prefer not to use port-forward, you can query the metrics endpoints directly from inside the core pod using the same bearer token.

```sh
kubectl exec -it <noobaa-core-pod> -- \
  curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://localhost:7004/ | head
kubectl exec -it <noobaa-core-pod> -- \
  curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://localhost:8080/metrics/web_server | head
kubectl exec -it <noobaa-core-pod> -- \
  curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://localhost:8080/metrics/bg_workers | head
kubectl exec -it <noobaa-core-pod> -- \
  curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://localhost:8080/metrics/hosted_agents | head
```

#### 6. Prometheus dashboard testing (local/minikube)
The following commands install kube-prometheus-stack locally, ensure the Prometheus instance selects the NooBaa ServiceMonitors (default label: `app: noobaa`), and port-forward the Prometheus UI.

```sh
helm install prometheus prometheus-community/kube-prometheus-stack \
  --set prometheus.prometheusSpec.serviceMonitorSelector.matchLabels.app=noobaa \
  --set prometheus.prometheusSpec.ruleSelector.matchLabels.prometheus=k8s \
  --set prometheus.prometheusSpec.ruleSelector.matchLabels.role=alert-rules
# The flags above make Prometheus select:
# - ServiceMonitors labeled `app: noobaa`
# - PrometheusRules labeled `prometheus: k8s` and `role: alert-rules`
# in the same namespace as Prometheus (default).
kubectl port-forward --namespace='default' prometheus-prometheus-kube-prometheus-prometheus-0 9090
```
#### Notes (local/minikube) - </br>
- Prometheus often uses label selectors, so ensure the ServiceMonitor labels match (default: `app: noobaa`), or update the label to match your selector.
- PrometheusRules created by the operator use labels `prometheus: k8s` and `role: alert-rules`, so your Prometheus rule selector must match those.

## Prometheus Alert Rules

NooBaa ships PrometheusRule resources with recording rules and alerts. Prometheus evaluates these rules on each scrape interval and sends firing alerts to Alertmanager (if configured). To see active rules and alerts:

- Prometheus UI (see [Prometheus dashboard testing](#6-prometheus-dashboard-testing-localminikube)): `http://127.0.0.1:9090/alerts` and `http://127.0.0.1:9090/rules`
- Confirm rules exist in the namespace:

```sh
kubectl -n <namespace> get prometheusrules
```

Reference: [Prometheus alerting rules](https://prometheus.io/docs/prometheus/latest/configuration/alerting_rules/)

## Examples

The examples below assume you have the port-forward from step 4 running and `JWT_TOKEN` set (see [Set JWT Token](#3-set-jwt-token)).
Values will vary based on runtime and workload.

### Direct Metrics Fetch Example

The following is an example of Prometheus text format output from the core metrics endpoint -

```shell
> curl -s -H "Authorization: Bearer ${JWT_TOKEN}" http://127.0.0.1:7004/ | head -n 12
# HELP NooBaa_Endpoint_process_cpu_user_seconds_total Total user CPU time spent in seconds.
# TYPE NooBaa_Endpoint_process_cpu_user_seconds_total counter
NooBaa_Endpoint_process_cpu_user_seconds_total 13.325259

# HELP NooBaa_Endpoint_process_cpu_system_seconds_total Total system CPU time spent in seconds.
# TYPE NooBaa_Endpoint_process_cpu_system_seconds_total counter
NooBaa_Endpoint_process_cpu_system_seconds_total 3.149571000000001

# HELP NooBaa_Endpoint_process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE NooBaa_Endpoint_process_cpu_seconds_total counter
NooBaa_Endpoint_process_cpu_seconds_total 16.47483
```

### Web Server Metrics Example

The following is an example of querying the web server metrics endpoint -

```shell
> curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://127.0.0.1:8080/metrics/web_server | head -n 12
# HELP NooBaa_WebServer_process_cpu_user_seconds_total Total user CPU time spent in seconds.
# TYPE NooBaa_WebServer_process_cpu_user_seconds_total counter
NooBaa_WebServer_process_cpu_user_seconds_total 39.02518

# HELP NooBaa_WebServer_process_cpu_system_seconds_total Total system CPU time spent in seconds.
# TYPE NooBaa_WebServer_process_cpu_system_seconds_total counter
NooBaa_WebServer_process_cpu_system_seconds_total 4.497256

# HELP NooBaa_WebServer_process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE NooBaa_WebServer_process_cpu_seconds_total counter
NooBaa_WebServer_process_cpu_seconds_total 43.522436000000006
```

### Background Workers Metrics Example

The following is an example of querying the background workers metrics endpoint -

```shell
> curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://127.0.0.1:8080/metrics/bg_workers | head -n 12
# HELP NooBaa_BGWorkers_process_cpu_user_seconds_total Total user CPU time spent in seconds.
# TYPE NooBaa_BGWorkers_process_cpu_user_seconds_total counter
NooBaa_BGWorkers_process_cpu_user_seconds_total 20.725111999999996

# HELP NooBaa_BGWorkers_process_cpu_system_seconds_total Total system CPU time spent in seconds.
# TYPE NooBaa_BGWorkers_process_cpu_system_seconds_total counter
NooBaa_BGWorkers_process_cpu_system_seconds_total 4.839238999999998

# HELP NooBaa_BGWorkers_process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE NooBaa_BGWorkers_process_cpu_seconds_total counter
NooBaa_BGWorkers_process_cpu_seconds_total 25.564351
```

### Hosted Agents Metrics Example

The following is an example of querying the hosted agents metrics endpoint -

```shell
> curl -k -H "Authorization: Bearer ${JWT_TOKEN}" http://127.0.0.1:8080/metrics/hosted_agents | head -n 12
# HELP NooBaa_HostedAgents_process_cpu_user_seconds_total Total user CPU time spent in seconds.
# TYPE NooBaa_HostedAgents_process_cpu_user_seconds_total counter
NooBaa_HostedAgents_process_cpu_user_seconds_total 13.149855

# HELP NooBaa_HostedAgents_process_cpu_system_seconds_total Total system CPU time spent in seconds.
# TYPE NooBaa_HostedAgents_process_cpu_system_seconds_total counter
NooBaa_HostedAgents_process_cpu_system_seconds_total 3.038472

# HELP NooBaa_HostedAgents_process_cpu_seconds_total Total user and system CPU time spent in seconds.
# TYPE NooBaa_HostedAgents_process_cpu_seconds_total counter
NooBaa_HostedAgents_process_cpu_seconds_total 16.188326999999997
```
