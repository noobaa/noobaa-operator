package backingstore

import (
	"github.com/spf13/cobra"
)

func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backing-store",
		Short: "Manage noobaa backing stores",
	}
	cmd.AddCommand(
		CmdList(),
		CmdCreateS3(),
		CmdDelete(),
	)
	return cmd
}

func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backing stores",
	}
	return cmd
}

func CmdCreateS3() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-s3",
		Short: "Create S3 backing store",
		Run:   RunCreateS3,
	}
	return cmd
}

func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete backing store",
		Run:   RunDelete,
	}
	return cmd
}

func RunCreateS3(cmd *cobra.Command, args []string) {
	panic("TODO implement RunCreateS3")
}

func RunDelete(cmd *cobra.Command, args []string) {
	panic("TODO implement RunDelete")
}
