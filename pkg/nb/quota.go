package nb

import (
	"fmt"
	"math/big"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

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

// ValidateQuotaAgainstBucketUsage ensures quota limits are not below current bucket usage .
func ValidateQuotaAgainstBucketUsage(bucket *BucketInfo, quota *QuotaConfig) error {
	if bucket == nil || quota == nil {
		return nil
	}
	var sizeBytes, objects int64
	if quota.Size != nil {
		if b, ok := QuotaSizeToBytes(quota.Size); ok {
			sizeBytes = b
		}
	}
	if quota.Quantity != nil && quota.Quantity.Value > 0 {
		objects = int64(quota.Quantity.Value)
	}
	if sizeBytes == 0 && objects == 0 {
		return nil
	}
	if bucket.BucketType == "NAMESPACE" {
		return nil
	}
	var currentNumObjects int64
	if bucket.NumObjects != nil {
		currentNumObjects = bucket.NumObjects.Value
	}
	if objects > 0 && currentNumObjects > objects {
		return util.ValidationError{
			Msg: fmt.Sprintf("bucket %q validation error: maxObjects quota %d is below current usage of %d objects",
				bucket.Name, objects, currentNumObjects),
		}
	}
	if sizeBytes > 0 {
		var currentUsageBytes *big.Int
		if bucket.DataCapacity != nil {
			currentUsageBytes = bucket.DataCapacity.Size.ToBig()
		} else {
			currentUsageBytes = big.NewInt(0)
		}
		if currentUsageBytes.Cmp(big.NewInt(sizeBytes)) > 0 {
			return util.ValidationError{
				Msg: fmt.Sprintf("bucket %q validation error: maxSize quota %s is below current usage of %s",
					bucket.Name,
					IntToHumanBytes(sizeBytes),
					BigIntToHumanBytes(bucket.DataCapacity.Size)),
			}
		}
	}
	return nil
}
