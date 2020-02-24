nb install
nb status
kubectl get noobaa
kubectl describe noobaa

# S3-comaptible
nb bucket create second.bucket
nb backingstore create s3-compatible nb1 --target-bucket first.bucket --endpoint 127.0.0.1:6443 --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY
nb backingstore create s3-compatible nb2 --target-bucket second.bucket --endpoint 127.0.0.1:6443 --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY
nb backingstore status nb1
nb backingstore status nb2
nb backingstore list
nb status
kubectl get backingstore
kubectl describe backingstore

# IBM-Cos
nb bucket create second.bucket
nb backingstore create ibm-cos nb1 --target-bucket first.bucket --endpoint 127.0.0.1:6443 --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY
nb backingstore create ibm-cos nb2 --target-bucket second.bucket --endpoint 127.0.0.1:6443 --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY
nb backingstore status nb1
nb backingstore status nb2
nb backingstore list
nb status
kubectl get backingstore
kubectl describe backingstore

# AWS-S3
nb backingstore create aws-s3 aws1 --target-bucket znoobaa --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY
nb backingstore create aws-s3 aws2 --target-bucket noobaa-qa --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY
nb backingstore status aws1
nb backingstore status aws2
nb backingstore list
nb status
kubectl get backingstore
kubectl describe backingstore

# Azure-Blob
nb backingstore create azure-blob blob1 --target-blob-container jacky-container --account-name $AZURE_ACCOUNT_NAME --account-key $AZURE_ACCOUNT_KEY

# Google
nb backingstore create google-cloud-storage google1 --target-bucket jacky-bucket --private-key-json-file ~/Downloads/noobaa-test-1-d462775d1e1a.json

# BucketClass
nb bucketclass create class1 --backingstores nb1
nb bucketclass create class2 --placement Mirror --backingstores nb1,aws1
nb bucketclass create class3 --placement Spread --backingstores aws1,aws2
nb bucketclass create class4 --backingstores nb1,nb2
nb bucketclass status class1
nb bucketclass status class2
nb bucketclass list
nb status
kubectl get bucketclass
kubectl describe bucketclass

# OBC
nb obc create buck1 --bucketclass class1
nb obc create buck2 --bucketclass class2
nb obc create buck3 --bucketclass class3 --app-namespace default
nb obc create buck4 --bucketclass class4
nb obc list
# nb obc status buck1
# nb obc status buck2
# nb obc status buck3
kubectl get obc
kubectl describe obc
kubectl get obc,ob,secret,cm -l noobaa-obc

AWS_ACCESS_KEY_ID=XXX AWS_SECRET_ACCESS_KEY=YYY aws s3 --endpoint-url XXX ls BUCKETNAME

nb obc delete buck1
nb obc delete buck2
nb obc delete buck3
nb bucketclass delete class1
nb bucketclass delete class2
nb backingstore delete aws1
nb backingstore delete aws2
nb backingstore delete nb1
nb backingstore delete nb2

nb uninstall
