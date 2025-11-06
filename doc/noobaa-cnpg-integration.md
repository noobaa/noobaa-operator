# NooBaa CloudNative-PG Integration

[NooBaa Operator](../README.md) /

## Introduction

NooBaa operator integrates with [CloudNative-PG (CNPG)](https://cloudnative-pg.io/) to provide high availability PostgreSQL database deployments for NooBaa metadata storage. This integration eliminates the single point of failure that existed with the previous standalone PostgreSQL StatefulSet approach.

CloudNative-PG is an open-source Kubernetes operator designed to manage PostgreSQL workloads on any supported Kubernetes cluster. It provides automated failover, self-healing capabilities, and declarative management of PostgreSQL configurations, making it an ideal solution for NooBaa's database requirements.

### Key Benefits

- **High Availability**: Automated failover and self-healing capabilities
- **Zero Data Loss**: Optional synchronous replication support
- **Declarative Management**: Kubernetes-native database configuration
- **Scalability**: Support for read replicas and cluster scaling
- **Monitoring**: Prometheus-compatible metrics export

## High-level CloudNative-PG Concepts

### Core Components

CloudNative-PG introduces several custom Kubernetes resources:

- **[Cluster](https://cloudnative-pg.io/documentation/1.27/cloudnative-pg.v1/#postgresql-cnpg-io-v1-Cluster)**: Represents a PostgreSQL cluster with one primary instance and optional replicas
- **[ImageCatalog](https://cloudnative-pg.io/documentation/1.27/cloudnative-pg.v1/#postgresql-cnpg-io-v1-ImageCatalog)**: Maps PostgreSQL major versions to specific container images
- **Services**: Automatically created read-write and read-only services for database access
- **Secrets**: Automatically managed database credentials and certificates

### Key Features

- **Asynchronous Replication**: Default replication mode for better performance
- **Synchronous Replication**: Optional zero data loss guarantee (RPO=0)
- **Automated Failover**: Promotes the most aligned replica when primary fails
- **Self-healing**: Automatic recreation of failed replicas
- **Planned Switchover**: Manual promotion of selected replicas
- **Monitoring**: Built-in Prometheus metrics export
- **TLS Support**: Secure connections with client certificate authentication

### Architecture

CloudNative-PG manages PostgreSQL clusters directly without using StatefulSets, instead directly managing Persistent Volume Claims (PVCs). This provides more flexibility and better integration with Kubernetes storage systems.

## High-level Integration Description

### CLI Integration

The NooBaa CLI provides comprehensive CloudNative-PG management commands:

```bash
noobaa cnpg --help
```

Available commands:
- `install` - Install CloudNative-PG operator
- `uninstall` - Uninstall CloudNative-PG operator  
- `yaml` - Show bundled CloudNative-PG YAML manifests
- `status` - Check CloudNative-PG operator deployment status

The CLI automatically installs CloudNative-PG when creating a new NooBaa system unless the `--use-standalone-db` flag is specified. When installing via OLM, the CloudNative-PG operator deployment and resources are included in the ClusterServiceVersion (CSV).

### Operator Reconciliation Loop

The NooBaa operator integrates CloudNative-PG reconciliation into its main reconciliation phases:

#### Phase 2: Creating
During the "Creating" phase, the operator:
1. **Creates ImageCatalog**: Maps PostgreSQL major versions to container images
2. **Creates Cluster**: Establishes the PostgreSQL cluster with appropriate configuration
3. **Handles Database Import**: For upgrades, imports data from existing standalone databases

#### Database Reconciliation Logic
The operator handles three main scenarios:

1. **Fresh Install**: Creates a new empty CNPG cluster
2. **Upgrade from Standalone DB**: Imports data from existing StatefulSet-based PostgreSQL
3. **Existing CNPG Cluster**: Updates configuration and manages cluster lifecycle

### OLM and CSV Generation

CloudNative-PG resources are integrated into the Operator Lifecycle Manager (OLM) deployment:

#### CSV Generation
- CNPG resources are included in the ClusterServiceVersion (CSV) when `--include-cnpg` flag is used
- CNPG CRDs are bundled with NooBaa operator CRDs
- Security contexts are modified for OLM compliance
- Related images include CNPG operator image

#### OLM Integration
- CNPG operator is deployed alongside NooBaa operator
- RBAC permissions are properly configured for namespace-scoped operations
- Webhook configurations are included for admission control

### API Group Customization

To avoid conflicts with upstream CloudNative-PG installations, NooBaa uses a custom API group:

- **Default**: `postgresql.cnpg.noobaa.io/v1`
- **Upstream**: `postgresql.cnpg.io/v1` (when `USE_CNPG_API_GROUP=true`)

This customization is handled through environment variables and YAML manifest modifications.

## Low-level Code Description

### Core Integration Files

#### `pkg/cnpg/cnpg.go`
Main CloudNative-PG integration package containing:

- **Resource Management**: Loading and modifying CNPG operator manifests
- **CLI Commands**: Implementation of `noobaa cnpg` commands
- **API Group Handling**: Custom API group configuration
- **RBAC Configuration**: Namespace-scoped permissions setup

Key functions:
- `LoadCnpgResources()`: Loads CNPG operator manifests from embedded files
- `modifyResources()`: Customizes resources for NooBaa deployment
- `getCnpgAPIGroup()`: Returns appropriate API group based on configuration

#### `pkg/system/db_reconciler.go`
Database reconciliation logic handling:

- **Cluster Reconciliation**: Main reconciliation function for CNPG clusters
- **Image Catalog Management**: Handles PostgreSQL version and image mapping
- **Import Logic**: Manages data import from standalone databases
- **Status Updates**: Tracks cluster readiness and status

Key functions:
- `ReconcileCNPGCluster()`: Main reconciliation entry point
- `reconcileCNPGImageCatalog()`: Manages PostgreSQL image versions
- `reconcileDBCluster()`: Handles cluster creation and updates
- `reconcileClusterCreateOrImport()`: Manages database import process

#### `pkg/apis/noobaa/v1alpha1/noobaa_types.go`
API definitions for database configuration:

- **NooBaaDBSpec**: Database specification in NooBaa CR
- **NooBaaDBStatus**: Database status reporting

Key fields:
- `DBSpec`: Optional database specification for managed PostgreSQL
- `DBImage`: PostgreSQL container image override
- `PostgresMajorVersion`: PostgreSQL major version specification
- `Instances`: Number of database instances
- `DBResources`: Resource requirements for database pods
- `DBMinVolumeSize`: Minimum volume size for database storage
- `DBStorageClass`: Storage class for database volumes
- `DBConf`: PostgreSQL configuration overrides

#### `pkg/apis/noobaa/v1alpha1/cnpg_types.go`
CloudNative-PG type registrations:

- Registers CNPG types with custom API group
- Handles both upstream and custom API group variants
- Integrates with Kubernetes scheme for type recognition

### Database Configuration

#### Default Configuration
The operator sets optimized PostgreSQL parameters:

```go
overrideParameters := map[string]string{
    "huge_pages":                   "off",
    "max_connections":              "600",
    "effective_cache_size":         "3GB",
    "maintenance_work_mem":         "256MB",
    "checkpoint_completion_target": "0.9",
    "wal_buffers":                  "16MB",
    "default_statistics_target":    "100",
    "random_page_cost":             "1.1",
    "effective_io_concurrency":     "300",
    "work_mem":                     "1747kB",
    "min_wal_size":                 "2GB",
    "max_wal_size":                 "8GB",
    "pg_stat_statements.track":     "all",
}
```

#### Resource Management
- **CPU and Memory**: Configurable through `DBSpec.DBResources`
- **Storage**: Configurable through `DBSpec.DBMinVolumeSize` and `DBSpec.DBStorageClass`
- **Instances**: Default 2 instances (1 primary + 1 replica)

### Import Process

For upgrades from standalone PostgreSQL deployments:

1. **Pod Shutdown**: Stops NooBaa core and endpoint pods
2. **Volume Analysis**: Checks existing database volume usage
3. **Size Calculation**: Doubles volume size if usage > 33%
4. **Import Configuration**: Sets up external cluster connection
5. **Data Migration**: Uses CNPG's microservice import type
6. **Cleanup**: Scales down old StatefulSet to 0 replicas

### Monitoring and Status

#### Status Tracking
The operator tracks database status through `NooBaaDBStatus`:

- `DBClusterStatus`: Current cluster state (None, Creating, Importing, Ready, Updating, Failed)
- `DBCurrentImage`: Current PostgreSQL image
- `CurrentPgMajorVersion`: PostgreSQL major version
- `ActualVolumeSize`: Actual volume size

#### Monitoring Configuration
- **Default**: Prometheus monitoring enabled
- **Disable**: Set annotation `noobaa.io/disable_db_default_monitoring=true`
- **Metrics**: Exported on port 9187


### Security Considerations

#### RBAC Configuration
- **Namespace-scoped**: Most permissions limited to NooBaa namespace
- **Cluster-scoped**: Only nodes and clusterimagecatalogs require cluster permissions
- **Webhook Permissions**: Separate cluster role for admission webhooks

#### Network Security
- **Internal Communication**: Database pods communicate within cluster network
- **Service Discovery**: Uses Kubernetes services for database access
- **TLS Support**: Optional TLS encryption for database connections

### Troubleshooting

#### Common Issues
1. **Import Failures**: Check volume space and permissions
2. **Cluster Not Ready**: Verify CNPG operator deployment
3. **Resource Constraints**: Ensure adequate CPU/memory allocation
4. **Storage Issues**: Verify storage class and volume provisioning

#### Debugging Commands
```bash
# Check CNPG operator status
noobaa cnpg status

# View cluster status
kubectl get clusters.postgresql.cnpg.noobaa.io

# Check cluster events
kubectl describe cluster <cluster-name>

# View CNPG operator logs
kubectl logs -n <namespace> deployment/cnpg-controller-manager
```

## References

- [CloudNative-PG Documentation v1.27](https://cloudnative-pg.io/documentation/1.27/)
- [PostgreSQL High Availability](https://www.postgresql.org/docs/current/high-availability.html)