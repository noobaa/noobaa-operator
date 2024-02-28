[NooBaa Operator](../README.md) /

# External Postgresql DB support

As part of a growing demand on supporting external Postgresql DB, we are now supporting a new basic option to use noobaa with an externally provided Postgresql DB. First, you will need to create a new empty DB with the needed collation by noobaa:
1. Validate you have a working Postgres that can be accessed from external pods.
2. Create a new DB that will be used by noobaa to save its metadata. Using the following SQL command:
```sql
CREATE DATABASE nbcore WITH LC_COLLATE = 'C' TEMPLATE template0;
```
3. Validate that you have the correct user and password in order to create new tables in this DB. Save the needed credentials and connection details in a URL format:
```
potgres://<user>:<password>@<full DNS name of external server>:<port>/<db-name>
```
for example:
```
postgres://user1:password1@externalserver.namespace.svc:5432/nbcore"
```
Now you have one of two options, depending on the way you are installing noobaa on your cluster:
1. Using Noobaa CLI: It will create a secret for you, where it will save the given URL (It will be called noobaa-external-pg-db)
This secret will be attached to the NooBaa system CRD when created during the installation. Run the following:
```bash
noobaa install --postgres-url="postgres://postgres:postgres@externalserver.namespace.svc:5432/nbcore"
```
2. You want to create your own noobaa instance by using yamls:
- Create a new secret saving the above URL:
```yaml
apiVersion: v1
kind: Secret
metadata:
  name: my-secret
  namespace: my-namespace
type: Opaque
stringData:
  db_url: postgres://postgres:postgres@externalserver.namespace.svc:5432/nbcore
```

- Create a new noobaa system, using the noobaa CRD. Add this to the spec:
```yaml
apiVersion: noobaa.io/v1alpha1
kind: NooBaa
metadata:
  name: noobaa
  namespace: my-namespace
spec:
  # ...
  externalPgSecret:
    name: my-secret
#...
```
## SSL support

Some external PG deployments for K8s require the clients to work with an encrypted SSL connection.

In order to allow working with an SSL connection do the following:

Server Side

If you want to force noobaa-core to communicate with the external server using an encrypted SSL connection, first make sure you have an external DB that can be connected using SSL, and then add the following options to the noobaa spec:
* externalPgSSLRequired -  will force the connection to be encrypted and validate the server certificate using the system-supported CAs. The default is false - no SSL.
* externalPgSSLUnauthorized - adding this option to the first one will force SSL, but will allow the server to use a self-signed certificate. The default is false - no self-signed certs allowed.
```yaml
spec:
  # ...
  externalPgSSLRequired: true
  externalPgSSLUnauthorized: false
```

Client Side

If as part of using SSL to communicate with the server, the server also demands that the user will use client-side certificate in order to authenticate itself, do the following:

create a new secret in the noobaa's namespace with the files provided to you, for example like this:

```bash
kubectl create secret generic secret_name --from-file=tls.crt --from-file=tls.key
```
make sure that the secret has two files in it:
1. tls.key - that will hold the client private key
2. tls.crt - that will hold the client public key

(please note the the file names must be tls.key and tls.crt)

Add a secret reference to this secret to the noobaa CR:
```yaml
spec:
  # ...
  externalPgSSLSecret:
    name: secret_name
```
NooBaa CLI also supports the following options to be used during install:
```bash
noob install --postgres-url="postgresql://postgres:noobaa@postgres-external.test.svc.cluster.local:5432/postgres"          --pg-ssl-required --pg-ssl-unauthorized --pg-ssl-key /certs/client.key --pg-ssl-cert /certs/client.crt
```
This will set SSL enabled with support of self-signed certs and with client certificate provided under local directory /certs/

Gaps:
We currently support only URL format for the connection details, we found it to be faster and easier. If demand rises, we will think of adding support for splitting the secret db_url key to host, port, db-name, user, and password keys.



