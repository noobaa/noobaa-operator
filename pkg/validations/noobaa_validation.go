package validations

import (
	"fmt"
	"net"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

// ValidateNoobaaDeletion check the existence of AllowNoobaaDeletion in Noobaa CR
func ValidateNoobaaDeletion(nb nbv1.NooBaa) error {
	if !nb.Spec.CleanupPolicy.AllowNoobaaDeletion {
		return util.ValidationError{
			Msg: "Noobaa cleanup policy is not set, blocking Noobaa deletion",
		}
	}
	return nil
}

// ValidateNoobaaUpdate validates that the updated Noobaa CR is valid
func ValidateNoobaaUpdate(nb nbv1.NooBaa) error {
	return validateNoobaaLoadbalancerSourceSubnets(nb)
}

// ValidateNoobaaCreation validates that the created Noobaa CR is valid
func ValidateNoobaaCreation(nb nbv1.NooBaa) error {
	return validateNoobaaLoadbalancerSourceSubnets(nb)
}

// validateNoobaaLoadbalancerSourceSubnets validates that the Noobaa CR has
// a valid loadbalancer source subnet
func validateNoobaaLoadbalancerSourceSubnets(nb nbv1.NooBaa) error {
	for _, subnet := range nb.Spec.LoadBalancerSourceSubnets.S3 {
		if err := validateCIDR(subnet); err != nil {
			return util.ValidationError{
				Msg: fmt.Sprintf("Invalid S3 loadbalancer source subnet %s: %s", subnet, err),
			}
		}
	}

	for _, subnet := range nb.Spec.LoadBalancerSourceSubnets.STS {
		if err := validateCIDR(subnet); err != nil {
			return util.ValidationError{
				Msg: fmt.Sprintf("Invalid STS loadbalancer source subnet %s: %s", subnet, err),
			}
		}
	}

	return nil
}

func validateCIDR(cidr string) error {
	_, _, err := net.ParseCIDR(cidr)
	if err != nil {
		return err
	}

	return nil
}
