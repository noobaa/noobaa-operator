package nb

import (
	"strconv"
)

// WarnIfQuotaCappedByFree emits a best-effort warning when the configured size quota exceeds
// the bucket's effective free capacity, meaning raising the quota will not increase Data Space Avail
func WarnIfQuotaCappedByFree(bucketName string, bucket *BucketInfo, nbClient Client, quota *QuotaConfig, warnf func(format string, args ...interface{})) {
	if quota == nil || quota.Size == nil || warnf == nil {
		return
	}
	requestedQuotaBytes, ok := SizeQuotaToBytes(quota.Size)
	if !ok {
		return
	}

	if bucket == nil {
		if nbClient == nil {
			return
		}
		b, err := nbClient.ReadBucketAPI(ReadBucketParams{Name: bucketName})
		if err != nil {
			return
		}
		bucket = &b
	}

	if bucket.BucketType == "NAMESPACE" || bucket.DataCapacity == nil || bucket.DataCapacity.Free == nil || bucket.DataCapacity.Size == nil {
		return
	}
	freeBytes, err := strconv.ParseInt(bucket.DataCapacity.Free.ToString(), 10, 64)
	if err != nil {
		return
	}
	usedBytes, err := strconv.ParseInt(bucket.DataCapacity.Size.ToString(), 10, 64)
	if err != nil {
		return
	}

	if requestedQuotaBytes-usedBytes > freeBytes {
		warnf("bucket %q size quota (%s) exceeds current effective free capacity (%s); Data Space Avail will remain capped (min(free, quota-used))",
			bucketName,
			IntToHumanBytes(requestedQuotaBytes),
			BigIntToHumanBytes(bucket.DataCapacity.Free),
		)
	}
}
