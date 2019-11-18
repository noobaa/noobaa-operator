The noobaa operator creates and reconciles a NooBaa system in a Kubernetes/Openshift cluster.

NooBaa provides an S3 object-store service abstraction and data placement policies to create hybrid and multi cloud data solutions.

For more information on using NooBaa refer to [Github](https://github.com/noobaa/noobaa-core) / [Website](https://www.noobaa.io) / [Articles](https://noobaa.desk.com). 

## How does it work?

- The operator deploys the noobaa core pod and two services - Mgmt (UI/API) and S3 (object-store).
- Both services require credentials which you will get from a secret that the operator creates - use describe noobaa to locate it.
- The service addresses will also appear in the describe output - pick the one that is suitable for your client:
    - minikube - use the NodePort address.
    - remote cluster - probably need one of the External addresses.
    - connect an application on the same cluster - use Internal DNS (though any address should work)
    
- Feel free to email us or open github issues on any question.

## Getting Started

### Notes:
- The following instructions are for **minikube** but it works on any Kubernetes/Openshift clusters.
- This will setup noobaa in the **my-noobaa-operator** namespace.
- You will need **jq**, **curl**, **kubectl** or **oc**, **aws-cli**.

### 1. Install OLM (if you don't have it already):
```
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/0.12.0/install.sh | bash -s 0.12.0
```

### 2. Install noobaa-operator:
```
kubectl create -f https://operatorhub.io/install/noobaa-operator.yaml
```
Wait for it to be ready:
```
kubectl wait pod -n my-noobaa-operator -l noobaa-operator --for=condition=ready
```

### 3. Create noobaa system:
```
curl -sL https://operatorhub.io/api/operator?packageName=noobaa-operator | 
    jq '.operator.customResourceDefinitions[0].yamlExample | .metadata.namespace="my-noobaa-operator"' |
    kubectl create -f -
```
Wait for it to be ready:
```
kubectl wait pod -n my-noobaa-operator -l noobaa-core --for=condition=ready
kubectl get noobaa -n my-noobaa-operator -w
# NAME     PHASE   MGMT-ENDPOINTS                  S3-ENDPOINTS                    IMAGE                    AGE
# noobaa   **Ready**   [https://192.168.64.12:31121]   [https://192.168.64.12:32557]   noobaa/noobaa-core:4.0   19m
```

### 4. Get system information to your shell:
```
NOOBAA_SECRET=$(kubectl get noobaa noobaa -n my-noobaa-operator -o json | jq -r '.status.accounts.admin.secretRef.name' )
NOOBAA_MGMT=$(kubectl get noobaa noobaa -n my-noobaa-operator -o json | jq -r '.status.services.serviceMgmt.nodePorts[0]' )
NOOBAA_S3=$(kubectl get noobaa noobaa -n my-noobaa-operator -o json | jq -r '.status.services.serviceS3.nodePorts[0]' )
NOOBAA_ACCESS_KEY=$(kubectl get secret $NOOBAA_SECRET -n my-noobaa-operator -o json | jq -r '.data.AWS_ACCESS_KEY_ID|@base64d')
NOOBAA_SECRET_KEY=$(kubectl get secret $NOOBAA_SECRET -n my-noobaa-operator -o json | jq -r '.data.AWS_SECRET_ACCESS_KEY|@base64d')
```

### 5. Connect to Mgmt UI:
```
# show email/password from the secret:
kubectl get secret $NOOBAA_SECRET -n my-noobaa-operator -o json | jq '.data|map_values(@base64d)'

# open mgmt UI login:
open $NOOBAA_MGMT
```

### 6. Connect to S3 with aws-cli:
```
alias s3='AWS_ACCESS_KEY_ID=$NOOBAA_ACCESS_KEY AWS_SECRET_ACCESS_KEY=$NOOBAA_SECRET_KEY aws --endpoint $NOOBAA_S3 --no-verify-ssl s3'
s3 ls
s3 sync /var/log/ s3://first.bucket
s3 ls s3://first.bucket
```
