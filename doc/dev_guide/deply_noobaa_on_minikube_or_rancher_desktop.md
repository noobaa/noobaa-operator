# Deploy NooBaa on Minikube / Rancher Desktop
There are many ways to deploy NooBaa on k8s on you local machine, this guide refers to 2 options (the difference is in the configurations):
1. Docker Desktop + Minikube
2. Rancher Desktop

The guide includes the steps:
1. Configuration (Docker Desktop + Minikube or Rancher Desktop)
2. Build images (noobaa-core and noobaa-operator) and optional step to push to a container registry (Quay or Docker Hub).
3. Install NooBaa using NooBaa CLI.

## 1) Configuration (First Time Only)

### Docker Desktop + Minikube
#### Change Your Docker Resource Preferences
Click on docker logo -> preference… -> Resource (left choice in the menu): CPU: 6, Memory: 10 GB, SWAP: 1 GB, Disk: 80 GB

#### Setting Your Minikube Configurations
a. Start your cluster:
```bash
minikube start
```
b. Set configurations:
```bash
minikube config set memory 8000
minikube config set cpus 5
```
##### Tips for minikube users:
c. Optional - You can also use minikube dashboard:
```bash
minikube dashboard --url
```
d. Optional - before starting you cluster (not talking about first time installation) it is better to run this command so the user won't get the previous cluster (sometimes minikube won't get deleted even after the system restart):
```bash
minikube delete --all
```
### Rancher Desktop
Inside the application -> preferences -> Virtual Machine (left tab) -> Memory: 10 GB, CPU: 6

## General Comment on Working in the Terminal
We will run the commands in the terminal, you may work with at least two tabs:
1. For noobaa-core repository
2. For noobaa-operator repository

In each step, it is mentioned what tab you should use.

## 2) Before Building The Images (Noobaa-Core Tab)
1) Change the working directory to the local repository (in this example the local repository is inside `SourceCode` directory):
```bash
 cd ./SourceCode/noobaa-core
```
2) (Minikube only): It is recommended to build all images on the minikube docker daemon. Configure your docker client to use minikube's docker run:
```bash
eval $(minikube docker-env)
```
Note: You can build on your local machine (it would be faster, since n minikube it would build the image from scratch) - but minikube doesn't have access to it, so you will need to push the image to a remote repository Docker Hub or Quay (and then pull it).

## 3) Before Building The Images (Noobaa-Operator Tab)
Change the working directory to the local repository (in this example the local repository is inside `SourceCode` directory).
```bash
 cd ./SourceCode/noobaa-operator
```
In order to build the CLI and the operator image run the following:
```bash
. ./devenv.sh
```
Notes: 
1) Minikube only: the file `devenv.sh` contains the command `eval $(minikube docker-env)`. We run the command `eval $(minikube docker-env)` prior to an image build (whether from noobaa-core repository or noobaa-operator repository).
2) When using Rancher Desktop you'll see: `WARNING: minikube is not started - cannot change docker-env` since the we don't use minikube. The `devenv.sh` script is setting an alias `nb` to run the local build of the CLI, so we will run it anyway.

### 4) Build Operator Images (Noobaa-Operator Tab)
```bash
make all
```
This will build the following:
* noobaa-operator image with tag `noobaa/noobaa-operator:<major.minor.patch>` (for example: `noobaa/noobaa-operator:5.13.0`). this tag is used by default when installing with the CLI.
* noobaa CLI. The `devenv.sh` script is setting an alias `nb` to run the local build of the CLI.

### 5) Push Operator Images (Noobaa-Operator Tab) - Optional
If we would like to save the image we can push it to a remote repository Docker Hub or Quay, you'll need an account, login to it in the terminal and create repositories noobaa-core and noobaa-operator. It is useful if you don't have any changes an the code.

Using Quay (your user in Quay):
```bash
docker tag noobaa-operator:<tag-name> quay.io/<your-user>/noobaa-operator:<tag-name>
docker push quay.io/<your-user>/noobaa-operator:<tag-name>
```

Using Docker Hub (your user in Docker Hub):
```bash
docker tag noobaa-core:<tag-name> <your-user>/noobaa-core:<tag-name>
docker push <your-user>/noobaa-core:<tag-name>
```

### 6) Build Core Images (Noobaa-Core Tab)
Run the following to build noobaa-core image with the desired tag to build the image:
```bash
make noobaa
```
Change the tag name  `noobaa:latest noobaa-core:<tag-name>`, for example: 
```bash
docker tag noobaa:latest noobaa-core:my-deploy
```
Tip: You can use this option instead of running the two commands:
```bash
make noobaa NOOBAA_TAG=noobaa-core:<tag-name>
```
### 7) Push Core Images (Noobaa-Core Tab) - Optional
Please refer to [Push Operator Images (Noobaa-Operator Tab) - Optional](#5-push-operator-images-noobaa-operator-tab---optional), the only change is instead of `noobaa-operator` use `noobaa-core`.

Tip: You can use this option instead of running the two commands (using Quay, your user in Quay):
```bash
make noobaa NOOBAA_TAG=quay.io/<your-user>/noobaa-core:<tag-name>
```
### 8) Deploy Noobaa (Noobaa-Operator Tab)
Deploy noobaa and you the image you created and tagged:
```bash
nb install --dev --noobaa-image='noobaa-core:my-deploy'
```
_Note: We have the alias to `nb` from the step 'Build Operator'._

In case you chose to use remote images (images that were pushed), for example using Quay:
```bash
nb install --dev --noobaa-image='quay.io/<your-user>/noobaa-core:<tag-name>' --operator-image='quay.io/<your-user>/noobaa-operator:<tag-name> -n noobaa
```

The installation should take 5-10 minutes.
Once noobaa is installed please notice that the phase is Ready, you will see it in the CLI logs:

✅ System Phase is "Ready".

You can see something similar to this when getting the pods:
```
> kubectl get pods
NAME                                               READY   STATUS    RESTARTS   AGE
noobaa-core-0                                      1/1     Running   0          51m
noobaa-db-pg-0                                     1/1     Running   0          51m
noobaa-default-backing-store-noobaa-pod-a586c55b   1/1     Running   0          47m
noobaa-endpoint-6cf5cccfc6-rmdrd                   1/1     Running   0          47m
noobaa-operator-5c959d5564-qzgqb                   2/2     Running   0          51m
```

### 9) Wait For Default Backingstore to Be Ready (Noobaa-Operator Tab)
Note that the default backing store might not be up as soon as the noobaa installation completes. For this reason it is advised to run `kubectl get pods` to make sure the default backing store is up. In case its not, wait for it to be up. If you run kubectl wait on the backing store before its up, the command will fail.

In case you use the default backingstore pod, we need it to be in phase Ready, run:
```bash
kubectl wait --for=condition=available backingstore/noobaa-default-backing-store --timeout=6m
```

