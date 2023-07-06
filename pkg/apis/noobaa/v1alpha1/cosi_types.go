package v1alpha1

import (
	cosiapi "sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage"
	cosiv1 "sigs.k8s.io/container-object-storage-interface-api/apis/objectstorage/v1alpha1"
)

// COSIBucketFinalizer is the name of the COSIBucket finalizer
const COSIBucketFinalizer = cosiapi.GroupName + "/finalizer"

// COSIBucketClaim is the API type for submitting bucket claims
type COSIBucketClaim = cosiv1.BucketClaim

// COSIBucketClaimList is a list of COSIBucketClaim
type COSIBucketClaimList = cosiv1.BucketClaimList

// COSIBucketClaimSpec defines the desired state of COSIBucketClaim
type COSIBucketClaimSpec = cosiv1.BucketClaimSpec

// COSIBucketClass is the API type for submitting bucket classes
type COSIBucketClass = cosiv1.BucketClass

// COSIBucketClassList is a list of COSIBucketClass
type COSIBucketClassList = cosiv1.BucketClassList

// COSIBucket is the API type for provisioners of buckets
type COSIBucket = cosiv1.Bucket

// COSIBucketList is a list of COSIBucket
type COSIBucketList = cosiv1.BucketList

// COSIBucketSpec defines the desired state of COSIBucket
type COSIBucketSpec = cosiv1.BucketSpec

// COSIProtocol is the API type for protocols of buckets
type COSIProtocol = cosiv1.Protocol

// COSIS3Protocol is a constant represents the s3 protocol of buckets
const COSIS3Protocol = cosiv1.ProtocolS3

// COSIDeletionPolicyRetain is a constant represents a retain deletion policy
const COSIDeletionPolicyRetain = cosiv1.DeletionPolicyRetain

// COSIDeletionPolicyDelete is  a constant represents a delete deletion policy
const COSIDeletionPolicyDelete = cosiv1.DeletionPolicyDelete
