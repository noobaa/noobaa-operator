[NooBaa Operator](../README.md) /

# External Postgresql DB support

As part of a growing demand on supporting external Postgresql DB, we are now supporting a new basic option to use noobaa with an externally provided Postgresql DB. First, you will need to create a new empty DB with the needed collation by noobaa:
1. Validate you have a working Postgres that can be accessed from external pods.
2. Create a new DB that will be used by noobaa to save its metadata. Using the following SQL command:
```sql
CREATE DATABASE nbcore WITH LC_COLLATE = ‘C’ TEMPLATE template0;
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
    namespace: my-namespace
#...
```

Gaps:
1. We currently support only MD5 encryption in order to connect to the DB. So no Support for SSL/TLS.
2. We currently support only URL format for the connection details, we found it to be faster and easier. If demand will rise we will think of adding support for splitting the secret db_url key to host, port, db-name, user, and password keys.


