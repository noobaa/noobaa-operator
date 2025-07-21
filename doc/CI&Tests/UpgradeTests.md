# NooBaa Operator - CI & Tests

1. [Introduction](#introduction)
2. [Upgrade Tests](#upgrade-tests)

## Introduction

NooBaa employs a Continuous Integration (CI) process to ensure the reliability and quality of its software. 
NooBaa Tests cover various aspects of functionality and integration. 
This proactive approach to testing enables NooBaa to deliver stable and efficient solutions for its users.

## Upgrade Tests

The NooBaa operator upgrade tests are part of the continuous integration (CI) pipeline and are designed to verify the stability and reliability of the system during version transitions.

The upgrade test process includes the following steps:

1. **Initial Deployment**: The latest release of the NooBaa operator is deployed on a Kubernetes cluster by default, see - [NooBaa Latest Release](https://api.github.com/repos/noobaa/noobaa-operator/releases/latest).
2. **System Initialization**: A fully functional NooBaa system is created, including resources such as buckets, objects, backing stores, and endpoints.
3. **Upgrade Execution**: The operator is upgraded to the target version using the noobaa CLI, the default target version is the latest noobaa-operator nightly build that could be found in [quay.io/noobaa/noobaa-operator](https://quay.io/repository/noobaa/noobaa-operator?tab=tags).
4. **Post-Upgrade Validation**: The health and functionality of the NooBaa system are verified after the upgrade. This includes checking the readiness of control plane components, data accessibility, and the integrity of configurations and runtime state.

These tests simulate upgrade scenarios and are continuously executed in the CI pipeline to detect upgrade-related issues early and ensure smooth operator version transitions.

### Run upgrade tests locally

In order to run upgrade tests locally with the default source and target versions, run the following command - 
```bash
make test-upgrade
```

In order to run upgrade tests locally with custom source and target versions, run the following command - 
```bash
UPGRADE_TEST_OPERATOR_SOURCE_VERSION=<upgrade-operator-source-version> UPGRADE_TEST_OPERATOR_TARGET_VERSION=<upgrade-operator-target-version> make test-upgrade
```

For instance - 
```bash
UPGRADE_TEST_OPERATOR_SOURCE_VERSION=5.17.0 UPGRADE_TEST_OPERATOR_TARGET_VERSION=5.18.6 make test-upgrade
```

### Manual Upgrade Tests Action 

The following action is a dispatchable GitHub Action for running the upgrade tests manually.  
It was implemented as part of the continuous integration (CI) process to allow on-demand execution of upgrade scenarios outside the scheduled or automated test pipeline.
See - [Manual Upgrade Tests Action](../../.github/workflows/manual-upgrade-tests.yaml).

### Nightly Upgrade Tests Action 

A nightly Upgrade tests github actions runs every night at 4AM UTC.
See - [Nightly Upgrade Tests Action](../../.github/workflows/nightly-upgrade-tests.yaml).