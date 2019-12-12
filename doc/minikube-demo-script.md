# Minikube Demo Script

### Install minikube and noobaa CLI
```bash
brew install minikube
brew install docker-machine-driver-hyperkit
brew install noobaa/noobaa/noobaa
```

### Configure and start minikube
```bash
minikube config set cpus 4
minikube config set memory 4096
minikube config set vm-driver hyperkit
minikube config view
minikube start
```

### Setup environment
```bash
. <(noobaa completion)
eval $(minikube docker-env)
```

### Create namespace
```bash
kubectl create ns noobaa
kubectl config set-context --current --namespace noobaa
```

### Install noobaa
```bash
noobaa install
noobaa status
```

### Create backingstore and bucketclass
```bash
noobaa backingstore create aws-s3 s3-store1
noobaa bucketclass create class1 --backingstores=s3-store1
```

### Create an ObjectBucketClaim (OBC) for an application
```bash
kubectl create ns app1
noobaa obc create bucket1 --app-namespace app1 --bucketclass class1
```

### Get the bucket information for the claim
```bash
noobaa obc status bucket1 --app-namespace app1
kubectl get obc,cm,secret -n app1
```

### Run an S3 client
```bash
# take this alias from obc status output
alias s3=... 

# the obc account has access only to a single bucket
s3 ls 2>/dev/null

# save the generated bucket name
BKT=$(s3 ls 2>/dev/null | cut -d' ' -f3)

# write some files
s3 cp --recursive /bin s3://$BKT/bin 2>/dev/null

# list of uploaded files
s3 ls --recursive --summarize $BKT 2>/dev/null

# read the files back
s3 sync s3://$BKT/bin /tmp/$BKT 2>/dev/null
diff -r /bin /tmp/$BKT/bin
```

### Delete the OBC

WARNING: This will delete the bucket and all the files since the default storage class specifies RetainPolicy: Delete

```bash
noobaa obc delete bucket1 --app-namespace app1
```
