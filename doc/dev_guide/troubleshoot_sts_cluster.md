# TroubleShoot AWS STS Cluster
Here are a couple of errors we saw during installations and how to investigate/fix them. The issues are happening when trying to install noobaa and the system is stuck in phase 'Configuring' while noobaa is trying to create the default backingstore that matches the AWS STS platform.

Please open the operator logs:

```bash
kubectl logs <operator-pod> -n <your-namespace> -f
```

Note: Changes in the printings of the logs attached here (in the logs you'll see your details):
- The role ARN with `<role-ARN>`.
- The request id with `<request-id>`.
- The namespace in the examples is "test1".

### Main issues:
#### 1) Wrong role:

```
time="2023-11-26T15:26:25Z" level=warning msg="⏳ Temporary Error: could not use AWS AssumeRoleWithWebIdentity with role name <role-ARN> and web identity token file /var/run/secrets/openshift/serviceaccount/token, AccessDenied: Not authorized to perform sts:AssumeRoleWithWebIdentity\n\tstatus code: 403, request id: <request-id>" sys=test1/noobaa
```

##### Solution:
Edit the role in the AWS console.
You can use the script in `scripts/create_aws_role.sh` to help you change the role.
Most of the time it is related to the "Condition" part under trusted entities.

You also need to see that the token is projected - for example in the operator pod:
```bash
MY_TOKEN_OPERATOR=$(kubectl exec $(kubectl get pods -n <your-namespace> | grep operator | awk '{ print $1}') -c noobaa-operator -n <your-namespace> -- cat /var/run/secrets/openshift/serviceaccount/token)
```

```bash
echo ${MY_TOKEN_OPERATOR}
```

And to verify that the issue is with the role please test it with assume-role-with-web-identity

```bash
aws sts assume-role-with-web-identity --role-arn <role-ARN> --role-session-name "test" --web-identity-token ${MY_TOKEN_OPERATOR}
```

You should see in the output the credentials (which includes the `AccessKeyId`, `SecretAccessKey`, and `SessionToken` - output example is in file `doc/dev_guide/create_aws_sts_setup_on_minikube.md`, but in case the role is wrong you'll see still `AccessDenied`, so you can create a new role with the script and test it.

Note: if this invalid role was sent as a part of OCP cluster you would need to update the subscription that have this environment variable: in odf-operator and mcg-operator subscriptions (in both of them - either in the UI or editing the YAMLs), search for the ROLEARN env name and update the value.

#### 2) Missing service account name in the role trusted policy
In case the role has a partial trust policy, for example - when this feature was launched in version 5.15 the service account name of stateful-set noobaa core was "noobaa" and in version 5.17 it was changed to "noobaa-core". This would cause `AccessDenied` error when the core pod tries assume-role-with-web-identity (hence any action would fail).

##### Solution:
The steps to trouble shoot this is to verify that every pod that sends a request using assume-role-with-web-identity (operator, core and endpoint) its mounted token has access - by running the mentioned steps in point number 1 (where it was demonstrated only in the operator pod) you should also run it in the endpoint and core pods.
For example in the core pod as well:

```bash
MY_TOKEN_CORE=$(kubectl exec $(kubectl get pods -n <your-namespace> | grep core | awk '{ print $1}') -n <your-namespace> -- cat /var/run/secrets/openshift/serviceaccount/token)
```

Notes:
- The next step as described in point number 1 it to run  instead of using `MY_TOKEN_OPERATOR` use `MY_TOKEN_CORE`.
- The token that is projected is different between the pods, and it is fine, what matters is the ability to check the assume-role-with-web-identity with each token. You can decoding the JWT tokens by running:

```bash
${MY_TOKEN} | cut -d '.' -f 2 | base64 -d | jq .
```

Partial output (when running on `MY_TOKEN_CORE` instead of `MY_TOKEN`):  
Note that *** signifies a redacted field.
In this output you can see that the token service account name is "noobaa-core".
```
{
  "aud": [
    "openshift"
  ],
  "exp": 1730137064,
  "iat": 1730133464,
  "iss": "https://kubernetes.default.svc",
  "jti": ***,
  "kubernetes.io": {
    "namespace": "openshift-storage",
    "node": {
      "name": ***,
      "uid": ***
    },
    "pod": {
      "name": "noobaa-core-0",
      "uid": ***
    },
    "serviceaccount": {
      "name": "noobaa-core",
      "uid": ***
    }
  },
  "nbf": ***,
  "sub": "system:serviceaccount:openshift-storage:noobaa-core"
}
```

In any case, better ask the user to attach the trust policy as a part of troubleshooting, and make sure to the needed lines of the service name:

```json
            "Action": "sts:AssumeRoleWithWebIdentity",
            "Condition": {
                "StringEquals": {
                    "${OIDC_PROVIDER}:sub": [
                        "system:serviceaccount:openshift-storage:noobaa",
                        "system:serviceaccount:openshift-storage:noobaa-core",
                        "system:serviceaccount:openshift-storage:noobaa-endpoint"
                    ]
                }
            }
```
where OIDC_PROVIDER will be filled according to the OIDC provider,  
and the `openshift-storage` is the namespace name (if it runs in a different namespace it would be different).

In case the system is already running and we need to update the trust policy, you would need to ask the user to update it (can be by simply editing the trust policy in the AWS console), then the check of assume-role-with-web-identity should return the credentials instead of `AccessDenied` error.

#### 3) Cluster configurations

```
time="2023-11-26T15:17:53Z" level=warning msg="⏳ Temporary Error: could not use AWS AssumeRoleWithWebIdentity with role <role-ARN> and web identity token file /var/run/secrets/openshift/serviceaccount/token, InvalidIdentityToken: No OpenIDConnect provider found in your account for https://kubernetes.default.svc\n\tstatus code: 400, request id: <request-id>" sys=test1/noobaa
```

##### Solution:
You need to make sure that the account issuer is set, try to run:

```bash
oc get authentication cluster -o jsonpath --template='{ .spec.serviceAccountIssuer }'
```

The structure of the output should be: 
1) In case the OIDC bucket configurations are in an S3 public bucket: `https://<oidc_bucket_name>.s3.<aws_region>.amazonaws.com`.
2) In case the OIDC bucket configurations are in an S3 private bucket (with a public CloudFront distribution URL): `d111111abcdef8.cloudfront.net` (this example it taken from [AWS docs](https://docs.aws.amazon.com/AmazonCloudFront/latest/DeveloperGuide/GettingStartedCreateDistribution.html))
Please follow the Openshift documentation.

#### 4) Wrong audience

```
time="2023-11-27T08:05:43Z" level=warning msg="⏳ Temporary Error: could not use AWS AssumeRoleWithWebIdentity with role name <role-ARN> and web identity token file /var/run/secrets/openshift/serviceaccount/token, InvalidIdentityToken: Incorrect token audience\n\tstatus code: 400, request id: <request-id>" sys=test1/noobaa
```

##### Solution:
Add the needed audience to match between the create role and the identity provider, for example:
- api - as we did in the local cluster example in `doc/dev_guide/create_aws_sts_setup_on_minikube.md`.
- openshift - as needed in the openshift cluster.

#### 5) Missing details:

```
time="2023-11-27T07:50:20Z" level=info msg="Secret noobaa-aws-cloud-creds-secret was created successfully by cloud-credentials operator" sys=test1/noobaa
time="2023-11-27T07:50:20Z" level=info msg="identified aws region us-east-2" sys=test1/noobaa
time="2023-11-27T07:50:20Z" level=info msg="Initiating a Session with AWS" sys=test1/noobaa
time="2023-11-27T07:50:20Z" level=info msg="AssumeRoleWithWebIdentityInput, roleARN =  webIdentityTokenPath = , " sys=test1/noobaa
time="2023-11-27T07:50:20Z" level=info msg="SetPhase: temporary error during phase \"Configuring\"" sys=test1/noobaa
time="2023-11-27T07:50:20Z" level=warning msg="⏳ Temporary Error: could not read WebIdentityToken from path , open : no such file or directory" sys=test1/noobaa
```

##### Solution:
The cloud credential operator (CCO) did not create the needed secret (it created a secret that matches AWS platform).
- Check that the secret contains the needed elements for AWS STS (role ARN and path for the token):

```bash
kubectl get secret noobaa-aws-cloud-creds-secret -n <your-namesapce> -o json | jq -r '.data.credentials' | base64 -d
```

You would see structure of:

```
[default]
aws_access_key_id = <access-key>
aws_secret_access_key = <secret-access-key>
```

instead of:

```
[default]
sts_regional_endpoints = regional
role_arn = <role-ARN>
web_identity_token_file = /var/run/secrets/openshift/serviceaccount/token
```

- Check that the credential request contains the needed elements (role ARN and path):

```bash
kubectl get credentialsrequest noobaa-aws-cloud-creds -n test1 -o json | grep -E 'stsIAMRoleARN|cloudTokenPath'
```

- Try to delete the credentialsrequest and the secret.

```bash
kubectl delete credentialsrequest noobaa-aws-cloud-creds -n <your-namespace>
```

```bash
kubectl delete secret noobaa-aws-cloud-creds-secret -n <your-namespace>
```

- If after noobaa operator creates a new credential request and we get the secret from the CCO and it still not match what we need, we need to investigate in the logs of the CCO.

```bash
kubectl logs $(kubectl get pod -n openshift-cloud-credential-operator | grep cloud-credential-operator | awk '{ print $1}') -c cloud-credential-operator -n openshift-cloud-credential-operator --tail 50 -f
```

#### 5) Other:

```
time="2023-12-20T09:46:59Z" level=info msg="AssumeRoleWithWebIdentityInput, roleARN = arn:aws:iam::<role-ARN>:role/<role-name> webIdentityTokenPath = /var/run/secrets/openshift/serviceaccount/token, " sys=openshift-storage/noobaa
time="2023-12-20T09:46:59Z" level=info msg="SetPhase: temporary error during phase \"Configuring\"" sys=openshift-storage/noobaa
time="2023-12-20T09:46:59Z" level=warning msg="⏳ Temporary Error: could not use AWS AssumeRoleWithWebIdentity with role name arn:aws:iam::<role-ARN>:role/<role-name> and web identity token file /var/run/secrets/openshift/serviceaccount/token, RequestError: send request failed\ncaused by: Post \"https://sts.amazonaws.com/\": tls: failed to verify certificate: x509: certificate signed by unknown authority" sys=openshift-storage/noobaa
```

In case you see this message and you checked that:
1. The token is projected (see part 1).
2. The AWS STS was successful in the test it with `assume-role-with-web-identity` (see part 1).
3. The credential request has role ARN (see part 4)
4. The secret has role ARN (see part 4).

Try to restart the noobaa pods: `kubectl delete pod <noobaa-pod>`.

Note: The above error message can be also if the role is not matched.
