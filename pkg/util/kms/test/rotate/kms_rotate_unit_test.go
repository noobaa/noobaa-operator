package kmsrotatetest

import (
	"fmt"
	"time"

	"github.com/noobaa/noobaa-operator/v5/pkg/util/kms"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

var _ = Describe("KMS - K8S Key Rotate - Unit", func() {
	now := time.Now()

	active_key := fmt.Sprintf("key-%v", now.UnixNano())

	Context("Verify Remove old keys", func() {
		tests := []struct {
			name     string
			givenTime   time.Time
			minKeys  int
			expected map[string]string
			deleteKeys int
		}{
		{
			name: "No keys to remove - all keys are out of range",
			givenTime:  now.AddDate(0, -9, 0),
			minKeys: 1,
			expected: map[string]string{
				"active_root_key": active_key,
				active_key:        "latest_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -1, 0).UnixNano()):           "1_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -2, 0).UnixNano()):           "2_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -3, 0).UnixNano()):           "3_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -4, 0).UnixNano()):           "4_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -5, 0).UnixNano()):           "5_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -6, 0).UnixNano()):           "6_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -7, 0).UnixNano()):           "7_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -8, 0).UnixNano()):           "8_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -9, 0).UnixNano()):           "9_month_old_key",
			},
			deleteKeys: 0,
		},
		{
			name: "Remove keys older than 6 months",
			givenTime:  now.AddDate(0, -6, 0),
			minKeys: 1,
			expected: map[string]string{
				"active_root_key": active_key,
				active_key:        "latest_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -1, 0).UnixNano()):           "1_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -2, 0).UnixNano()):           "2_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -3, 0).UnixNano()):           "3_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -4, 0).UnixNano()):           "4_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -5, 0).UnixNano()):           "5_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -6, 0).UnixNano()):           "6_month_old_key",
			},
			deleteKeys: 3,
		},
		{
			name: "Remove keys older than 15 days - Retain at least min_keys keys",
			givenTime:  now.AddDate(0, 0, -15),
			minKeys: 3,
			expected: map[string]string{
				"active_root_key": active_key,
				active_key:        "latest_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -1, 0).UnixNano()):           "1_month_old_key",
				fmt.Sprintf("key-%v", now.AddDate(0, -2, 0).UnixNano()):           "2_month_old_key",
			},
			deleteKeys: 7,
		},
		}
		for _, tt := range tests {
			Specify(tt.name, func() {
				secret := map[string]string{}
				for i := 1; i < 10; i++ {
					secret[fmt.Sprintf("key-%v", now.AddDate(0, -i, 0).UnixNano())] = fmt.Sprintf("%v_month_old_key", i)
				}
				secret["active_root_key"] = active_key
				secret[active_key] = "latest_key"
				deleted_keys, err := kms.RemoveOldKeysFromSecret(secret, tt.givenTime, tt.minKeys)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(deleted_keys).To(Equal(tt.deleteKeys))
				Expect(secret).To(Equal(tt.expected))
			})
		}
	})
})
