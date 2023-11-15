package sts

import (
	"encoding/json"
	"log"

	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/spf13/cobra"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sts",
		Short: "Manage the NooBaa Security Token Service",
		Long: "Manage the NooBaa Security Token Service by assigning, updating or removing a NooBaa account's role config.\n" +
			"The role config object must contain the keys 'role_name' and 'assume_role_policy', with their respective values.",
	}
	cmd.AddCommand(
		CmdAssignRole(),
		CmdRemoveRole(),
	)
	return cmd
}

// CmdAssignRole returns a CLI command
func CmdAssignRole() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "assign-role <noobaa-account-name> <role-config>",
		Short: "Assign a role config to a NooBaa account - note that this will override the existing role config",
		Run:   RunAssign,
	}
	cmd.Flags().String("email", "", "The email of the account that will be updated")
	err := cmd.MarkFlagRequired("email")
	if err != nil {
		log.Fatalf(`❌ Failed to mark email flag as required - %s`, err)
	}
	cmd.Flags().String("role_config", "", "The new value that the account's role_config should be set to")
	err = cmd.MarkFlagRequired("role_config")
	if err != nil {
		log.Fatalf(`❌ Failed to mark role_config flag as required - %s`, err)
	}
	return cmd
}

// CmdRemoveRole returns a CLI command
func CmdRemoveRole() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove-role <noobaa-account-name>",
		Short: "Remove a NooBaa account's role config",
		Run:   RunRemove,
	}
	cmd.Flags().String("email", "", "The email of the account that will be updated")
	err := cmd.MarkFlagRequired("email")
	if err != nil {
		log.Fatalf(`❌ Failed to mark email flag as required - %s`, err)
	}
	return cmd
}

// RunAssign runs a CLI command
func RunAssign(cmd *cobra.Command, args []string) {
	log := util.Logger()
	email, _ := cmd.Flags().GetString("email")
	roleConfig, _ := cmd.Flags().GetString("role_config")

	if !json.Valid([]byte(roleConfig)) {
		log.Fatalf(`❌ The provided role configuration is not valid JSON`)
	}

	sysClient, err := system.Connect(true)
	if err != nil {
		log.Fatalf(`❌ Failed to create RPC client %s`, err)
	}
	NBClient := sysClient.NBClient

	var roleConfigObject interface{}
	err = json.Unmarshal([]byte(roleConfig), &roleConfigObject)
	if err != nil {
		log.Fatalf("❌ Failed to parse role config - %s", err)
	}
	if err != nil {
		log.Fatalf(`❌ Failed to read account - %s`, err)
	}
	UpdateAccountParams := nb.UpdateAccountParams{
		Email:      email,
		RoleConfig: roleConfigObject,
	}

	err = NBClient.UpdateAccount(UpdateAccountParams)
	if err != nil {
		log.Fatalf(`❌ Failed to update account - %s`, err)
	}
}

// RunRemove runs a CLI command
func RunRemove(cmd *cobra.Command, args []string) {
	email, _ := cmd.Flags().GetString("email")

	sysClient, err := system.Connect(true)
	if err != nil {
		log.Fatalf(`❌ Failed to create RPC client %s`, err)
	}
	NBClient := sysClient.NBClient

	UpdateAccountParams := nb.UpdateAccountParams{
		Email:            email,
		RemoveRoleConfig: true,
	}

	err = NBClient.UpdateAccount(UpdateAccountParams)
	if err != nil {
		log.Fatalf(`❌ Failed to remove the requested role config - %s`, err)
	}
}
