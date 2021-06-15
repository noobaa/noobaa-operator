package version

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/version"

	"github.com/spf13/cobra"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version",
		Run:   RunVersion,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunVersion runs a CLI command
func RunVersion(cmd *cobra.Command, args []string) {
	log := util.Logger()
	log.Printf("CLI version: %s\n", version.Version)
	log.Printf("noobaa-image: %s\n", options.NooBaaImage)
	log.Printf("operator-image: %s\n", options.OperatorImage)
}
