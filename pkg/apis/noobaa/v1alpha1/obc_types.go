package v1alpha1

import (
	obv1 "github.com/kube-object-storage/lib-bucket-provisioner/pkg/apis/objectbucket.io/v1alpha1"
	"github.com/kube-object-storage/lib-bucket-provisioner/pkg/provisioner/api"
)

// ObjectBucketFinalizer is the name of the ObjectBucket finalizer
const ObjectBucketFinalizer = api.Domain + "/finalizer"

// ObjectBucketClaim is the API type for submitting bucket claims
type ObjectBucketClaim = obv1.ObjectBucketClaim

// ObjectBucketClaimList is a list of ObjectBucketClaim
type ObjectBucketClaimList = obv1.ObjectBucketClaimList

// ObjectBucketClaimSpec defines the desired state of ObjectBucketClaim
type ObjectBucketClaimSpec = obv1.ObjectBucketClaimSpec

// ObjectBucket is the API type for provisioners of buckets
type ObjectBucket = obv1.ObjectBucket

// ObjectBucketList is a list of ObjectBucket
type ObjectBucketList = obv1.ObjectBucketList

// ObjectBucketSpec defines the desired state of ObjectBucket
type ObjectBucketSpec = obv1.ObjectBucketSpec

// ObjectBucketConnection is the internal API type that Provision() should populate
type ObjectBucketConnection = obv1.Connection

// ObjectBucketEndpoint is the info needed for apps to access the bucket (besides auth credentials)
type ObjectBucketEndpoint = obv1.Endpoint

// ObjectBucketAuthentication is the auth credentials info needed for apps to access the bucket
type ObjectBucketAuthentication = obv1.Authentication

// ObjectBucketAccessKeys is the access keys inside ObjectBucketAuthentication
type ObjectBucketAccessKeys = obv1.AccessKeys
