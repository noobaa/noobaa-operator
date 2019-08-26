NAMESPACES="marketplace olm operators my-noobaa-operator"
NOOBAA_NAMES=$(kubectl api-resources --api-group=noobaa.io -o name)
OBJECTBUCKET_NAMES=$(kubectl api-resources --api-group=objectbucket.io -o name)
OPERATORS_NAMES=$(kubectl api-resources --api-group=operators.coreos.com -o name)

CLUSTER_ROLES=$(kubectl get clusterrole -o name | egrep "olm|marketplace|noobaa")
CLUSTER_ROLE_BINDINGS=$(kubectl get clusterrolebinding -o name | egrep "olm|marketplace|noobaa")
STORAGE_CLASSES=$(kubectl get storageclass -o name | grep "noobaa")

echo "Delete namespaced resources ..."
for r in $NOOBAA_NAMES $OBJECTBUCKET_NAMES $OPERATORS_NAMES
do
    kubectl delete --all -A $r
done

echo "Delete namespaces ..."
kubectl delete ns $NAMESPACES

echo "Delete cluster-wide resources (ClusterRole, ClusterRoleBinding, StorageClass) ..."
[ -n "$CLUSTER_ROLES" ]         && kubectl delete $CLUSTER_ROLES
[ -n "$CLUSTER_ROLE_BINDINGS" ] && kubectl delete $CLUSTER_ROLE_BINDINGS
[ -n "$STORAGE_CLASSES" ]       && kubectl delete $STORAGE_CLASSES

echo "Delete CRDs ..."
[ -n "$NOOBAA_NAMES" ]          && kubectl delete crd $NOOBAA_NAMES
[ -n "$OBJECTBUCKET_NAMES" ]    && kubectl delete crd $OBJECTBUCKET_NAMES
[ -n "$OPERATORS_NAMES" ]       && kubectl delete crd $OPERATORS_NAMES

echo "Done. "