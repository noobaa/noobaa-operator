# Bucket Notifications

## NooBaa CRD configuration

NooBaa CRD contains a field named 'bucketNotifications' (under "spec" field) that manages bucket notifications.

bucketNotifications value is an object:

	{
		enabled: Boolean,
		pvc: String,
		connections: [SecretReference]
	}

- enabled: This field must be set to true to enable bucket notifications.
- pvc: Bucket notifications uses a PVC to store pending notifications.
This field is used to specify an existing PVC to use.
The PVC must support RWX access mode.
If CephFS is available, and this field is left empty, a PVC would be allocated from CephFS.
- connections: An array of secrets that contain the data on how to connect to the external notification server.
When refering to a connection in the notification's configuration's TopicArn field, use
`<secret name>/<file used to created the secret>`
Eg if secret was created with:
`kubectl create secret generic notif-secret --from_file connect.json`
Then the TopicArn should be
`notif-secret/connect.json`

## NooBaa CRD configuration example

bucketNotifications:
connections:
- name: notif-secret
	namespace: my-namespace
- name: notif-secret2
	namespace: my-namespace
enabled: true
pvc: notif-pvc

## Connection secrets
The structure of a connection secret is the same as connection file.
Please refer to the Bucket Notification document in [noobaa-core](https://github.com/noobaa/noobaa-core/blob/master/docs/bucket-notifications.md).
The operator will mount the secrets as files in the core pod at /etc/notif_connect/<secret name>.
