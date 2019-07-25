# NooBaa Operator

NooBaa is an object data service for hybrid and multi cloud environments. NooBaa runs on kubernetes, provides an S3 object store service (and Lambda with bucket triggers) to clients both inside and outside the cluster, and uses storage resources from within or outside the cluster, with flexible placement policies to automate data use cases.

[About NooBaa](doc/about-noobaa.md)

# Using the operator as CLI

- Download the compiled operator binary from the [releases page](/releases)
- Run: `noobaa --help` for CLI usage
- Install the operator and noobaa with: `noobaa install`

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
- Source the devenv into your shell: `. devenv.sh`
- Build the project: `make`
- Test with the alias `nb` that runs the local operator from `build/_output/bin` (alias created by devenv)
- `nb install`
- ...
