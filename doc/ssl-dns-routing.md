[NooBaa Operator](../README.md) /
# SSL, Name Resolution and Routing

### NooBaa TLS/SSL support:
By default, NooBaa is using self-signed certificates while serving both S3 and management traffic.
This behavior can be changed by creating a `noobaa-mgmt-serving-cert` and a `noobaa-s3-serving-cert` secret in the same namespace where the NooBaa system is installed.
The certificates held by these secrets will be used by NooBaa to serve management and s3 traffic respectively.
Each one of the described secrets should contain the following keys under its data section:
- `tls.crt` to hold the public key in PEM format
- `tls.key` to hold the private key in PEM format

A simple way to create these secrets is using the `kubectl` CLI (or an equivalent CLI like `oc` in OpenShift environments) and running the following command:

```kubectl create secret generic secret_name --from-file=tls.crt --from-file=tls.key```

where `secret_name` is either `noobaa-mgmt-serving-cert` or `noobaa-s3-serving-cert` and `tls.crt` / `tls.key` are files that holds the public and secret keys for the appropriate certificate respectivly (please note the the  files names must be `tls.key` and `tls.crt`)

Alternatively you can change the secret's references inside the `noobaa-core` statefulset to point to diffrent secrets in the namespace by running the CLI command:

```kubectl patch sts noobaa-core -p '{"spec":{"template"{"spec"{"volumes":[{"name":"volume_name","secret":{"secretName": "secret_name"}}]}}}}'```

Where `volume_name` can be either `mgmt-secret` or `s3-secret` and `secret_name` is the name of the secret in the namespace that holds the certificate private and public key as described above.
To use the same secret (and consequently the same certificate) for both management and S3 traffic, both volumes can be pointed to the same secret.

NooBaa will monitor changes to the mounted secrets and the certificates held inside and will reload these certificates upon any change. This behavior is eventually consistent and may take some time before the change is reflected in the management and S3 endpoints traffic.

### TLS/SSL on OpenShift 4.2+
Every OpenShift 4.2+ cluster provides a service that can be used for the automatic issuing of certificates for services installed in the cluster.
NooBaa is taking advantage of this service in a way that any NooBaa installed using the NooBaa operator inside an OpenShift 4.2+ cluster will instruct this service to automatically issue both `mgmt-secret` and `s3-secret` and store them inside a   `noobaa-mgmt-serving-cert` and `noobaa-s3-serving-cert` secrets (as described in the previous section).
These certificates will hold the CN (Common Name) `noobaa-mgmt.namespace.svc` and `s3.namespace.svc` respectively (where `namespace` is the Kubernetes namespace where NooBaa is installed) and are issued by an `openshift-service-serving-signer` CA which has its CA certificate mounted on every pod in the cluster at `/var/run/secrets/kubernetes.io/serviceaccount/service-ca.crt`. The CA certificate should be added to the trusted root of any client installed on a pod inside the cluster to establish secure management and S3 connections.

The certificate issuing service will automatically override any changes to any secrets managed by it, including `noobaa-mgmt-serving-cert` and `noobaa-s3-serving-cert`.
To disable this behavior run the CLI command:

```kubectl annotate service servce_name service.beta.openshift.io/serving-cert-secret-name-```

where `service_name` can be either `noobaa-mgmt` or `s3` (please note the `-` sign in the end).
Alternatively, you can use the method described in the previous section to instruct NooBaa to load the certificates from a set of different secrets in the namespace.

### Name Resolution
NooBaa services can be accessed, from within the cluster, using the local names `noobaa-mgmt.namespace.svc` and `s3.namespace.svc` respectively, where `namespace` is the Kubernetes namespace where NooBaa is installed. Any certificate issued for these services should include at least the above local name as it's CN (Common Name) or as one of it's SAN (Subject Alternative Names).

### External Access, Public Names, and OpenShift Routes
OpenShift introduces the notion of routes. A route defines a public access point (hostname) and an optional TLS termination policy (in the case of secured routes) for a given Kubernetes resource. In most cases, the resource in question is a Kubernetes service.
The NooBaa operator automatically creates a secure route for every NooBaas service. The public hostnames configured for these routes are of the form `service_name-namespace.apps.cluster_domain_name` where:
- `service_name` in either `noobaa-mgmt` or `s3`
- `namespace` is the Kubernetes namespace where noobaa is installed
- `cluster_public_domain_name` is the cluste_domain_name as defined by the cluster's ingress operator.

The cluster's ingress operator issues a wildcard certificate with a CA (Common Name) of the form `*.apps.cluster_domain_name` that is used to verify any traffic that matches the above form. This includes the hostnames defined by the default NooBaa routes. The ingress operator uses the ingress CA certificate installed in the cluster to issue this certificate, by default the ingress operator CA certificate is self-signed but this can be changed (see https://docs.openshift.com/container-platform/4.2/networking/ingress-operator.html for more details)

The routes created by the NooBaa operator are configured to use a re-encrypt TLS termination policy, this configuration dictates that any secure traffic directed toward the routes' defined hostnames will be directed at one of the routers in the cluster and will be encrypted using the ingress operator certificate. The router will decrypt the packets then re-encrypt them using the certificate that matches the NooBaa service local name. This configuration allows for secure communication from clients both outside and inside the cluster.

In some cases, the re-encrypt termination process overhead may introduce some performance issues. To mitigate that the routes' TLS termination policy can be changed from re-encrypt to passthrough using the CLI command:

```kubectl patch route service_name -p '{"spec":{"tls":{"termination":"passthrough"}}}'```

where `service_name` is either `noobaa-mgmt` or `s3`.
A passthrough TLS termination dictates that the router should be used for name (and port) resolution only, any packets will pass through the router without any intervention from the side of the router. This means that the certificate that will be used for secure traffic will be the certificate installed in the NooBaa brain and not the ingress certificate.
To enable secure communication using this method, an admin should install a custom certificate that matches both the cluster local name for and route public hostname, for any given service, as either CN (Common Name) or SANs (Subject Alternate Names).

Please note that the certificates issued automatically by OpenShift 4.2+ (as described in previous sections) do not match these criteria and cannot be used for secure communication using the routes' hostnames when using passthrough TLS termination policy.
