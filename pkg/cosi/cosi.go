package cosi

import (
	"github.com/spf13/cobra"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cosi",
		Short: "Manage cosi resources",
	}
	cmd.AddCommand(
		CmdCOSIBucketClaim(),
		CmdCOSIBucketClass(),
	)
	return cmd
}
