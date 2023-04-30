[NooBaa Operator](../README.md) /
# Store Connection Secrets

In order for stores to successfully connect to their target storage provider, a secret needs to be provided.

This document outlines the different keys needed for a store's connection secret.

Please note that *all* values under the `data` section have to be Base64 encoded. 

Values that depend on the user's choice will be marked with `<>`.

## AWS-S3, S3-Compatible
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <>
  namespace: <>
type: Opaque
data:
  AWS_ACCESS_KEY_ID: <>
  AWS_SECRET_ACCESS_KEY: <>
```

## Google Cloud Platform
In this case, the value of `GoogleServiceAccountPrivateKeyJson` is the full contents of the appropriate `credentials.json` (acquired from GCP), encoded in Base64
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <>
  namespace: <>
type: Opaque
data:
  GoogleServiceAccountPrivateKeyJson: <>
```

## Azure
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <>
  namespace: <>
type: Opaque
data:
  STORAGE_ACCOUNT_NAME: <>
  STORAGE_ACCOUNT_KEY: <>
```

## IBM COS
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: <>
  namespace: <>
type: Opaque
data:
  IBM_COS_ACCESS_KEY_ID: <>
  IBM_COS_SECRET_ACCESS_KEY: <>
```