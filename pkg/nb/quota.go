package nb

import "math/big"

// WarnIfQuotaCappedByFree emits a best-effort warning when the configured size quota exceeds
// the bucket's effective free capacity, meaning raising the quota will not increase Data Space Avail
func WarnIfQuotaCappedByFree(bucketName string, bucket *BucketInfo, nbClient Client, quota *QuotaConfig, warnf func(format string, args ...interface{})) {
	if quota == nil || quota.Size == nil || warnf == nil {
		return
	}
	requestedQuotaBytes, ok := QuotaSizeToBytes(quota.Size)
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

	cap := bucket.DataCapacity
	if bucket.BucketType == "NAMESPACE" || cap == nil || cap.Free == nil || cap.Size == nil {
		return
	}

	// warn if (quota - used) > free, meaning free space is the binding constraint
	// use big.Int to avoid int64 overflow for large capacity values
	quotaMinusUsed := new(big.Int).Sub(big.NewInt(requestedQuotaBytes), cap.Size.ToBig())
	if quotaMinusUsed.Cmp(cap.Free.ToBig()) > 0 {
		warnf("bucket %q size quota (%s) exceeds current effective free capacity (%s); Data Space Avail will remain capped (min(free, quota-used))",
			bucketName,
			IntToHumanBytes(requestedQuotaBytes),
			BigIntToHumanBytes(cap.Free),
		)
	}
}
