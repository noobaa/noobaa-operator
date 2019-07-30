# NooBaa Operator

NooBaa is an object data service for hybrid and multi cloud environments. NooBaa runs on kubernetes, provides an S3 object store service (and Lambda with bucket triggers) to clients both inside and outside the cluster, and uses storage resources from within or outside the cluster, with flexible placement policies to automate data use cases.

[About NooBaa](doc/about-noobaa.md)

# Using the operator as CLI

- Download the compiled operator binary from the [releases page](https://github.com/noobaa/noobaa-operator/releases)

For Mac
```
wget https://github.com/noobaa/noobaa-operator/releases/download/v1.0.2/noobaa-mac-v1.0.2;mv noobaa-mac-* noobaa;chmod +x noobaa
```
For Linux
```
wget https://github.com/noobaa/noobaa-operator/releases/download/v1.0.2/noobaa-linux-v1.0.2;mv noobaa-linux-* noobaa;chmod +x noobaa
```

- Run: `./noobaa --help` for CLI usage
- Install the operator and noobaa with: `./noobaa install`
  The install output includes S3 service endpoint and credentials, as well as web management console address with credentials.
- Getting this information is always available with: `./noobaa status`
- Remove NooBaa deployment can be done with: `./noobaa uninstall`
# Troubleshooting

- The operator is running, but there is no noobaa-core-0 pod 

    Make sure that there is a single default storage class with `oc get sc`. run `oc describe sts` for more information
    
- The operator is running, but the noobaa-core-0 is pending

    Verify that there are enough resrouces. run `oc describe pod noobaa-core-0` for more information

# Operator Design

CRDs
- [NooBaa](doc/noobaa-crd.md) - The basic CRD to deploy a NooBaa system.
- [BackingStore](doc/backing-store-crd.md) - Connection to cloud or local storage to use in policies.
- [BucketClass](doc/bucket-class-crd.md) - Policies applied to a class of buckets.

Applications
- [S3 Account](doc/s3-account.md) - Method to obtain S3 account credentials for native S3 applications.
- [OBC Provisioner](doc/obc-provisioner.md) - Method to claim a new/existing bucket.

# Developing

- Fork and clone the repo: `git clone https://github.com/<username>/noobaa-operator`
- Use minikube: `minikube start`
- Use you package manager to install `go`, `operator-sdk` and `python3`.
- Source the devenv into your shell: `. devenv.sh`
- Build the project: `make`
- Test with the alias `nb` that runs the local operator from `build/_output/bin` (alias created by devenv)
- Install the operator and create the system with: `nb install`
