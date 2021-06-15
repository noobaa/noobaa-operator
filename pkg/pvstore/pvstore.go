package pvstore

import (
	"fmt"

	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/spf13/cobra"
)

// Cmd creates a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pvstore",
		Short: "Manage noobaa pv store",
	}
	cmd.AddCommand(
		CmdCreate(),
		CmdList(),
		CmdDelete(),
	)
	return cmd
}

// CmdCreate creates a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <store-name>",
		Short: "Create a NooBaa PV Store",
		Run:   RunCreate,
	}

	cmd.Flags().Uint32(
		"num-volumes", 0,
		`Number of volumes in the store`,
	)
	cmd.Flags().Uint32(
		"pv-size-gb", 0,
		`PV size of each volume in the store`,
	)

	return cmd
}

// CmdDelete creates a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <store-name>",
		Short: "Deletes a NooBaa PV Store",
		Run:   RunDelete,
	}
	return cmd
}

// CmdList creates a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List NooBaa PV stores",
		Run:   RunList,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <pv-store-name> %s`, cmd.UsageString())
	}

	nbClient := system.GetNBClient()

	poolName := args[0]
	numVolumes, _ := cmd.Flags().GetUint32("num-volumes")
	pvSizeGB, _ := cmd.Flags().GetUint32("pv-size-gb")

	if numVolumes == 0 {
		fmt.Printf("Enter number of volumes: ")
		_, err := fmt.Scan(&numVolumes)
		util.Panic(err)
		if numVolumes == 0 {
			log.Fatalf(`❌ Missing number of volumes %s`, cmd.UsageString())
		}
	}
	if numVolumes > 20 {
		log.Fatalf(`❌ Number of volumes seems to be too large %d %s`, numVolumes, cmd.UsageString())
	}

	if pvSizeGB == 0 {
		fmt.Printf("Enter PV size (GB): ")
		_, err := fmt.Scan(&pvSizeGB)
		util.Panic(err)
		if pvSizeGB == 0 {
			log.Fatalf(`❌ Missing PV size (GB) %s`, cmd.UsageString())
		}
	}
	if pvSizeGB > 1024 {
		log.Fatalf(`❌ PV size seems to be too large %d %s`, pvSizeGB, cmd.UsageString())
	}

	gbsize := int64(pvSizeGB) * 1024 * 1024 * 1024
	_, err := nbClient.CreateHostsPoolAPI(nb.CreateHostsPoolParams{
		Name:       poolName,
		IsManaged:  true,
		HostCount:  int(numVolumes),
		HostConfig: nb.PoolHostsInfo{VolumeSize: gbsize},
	})
	util.Panic(err)
}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <pv-store-name> %s`, cmd.UsageString())
	}

	nbClient := system.GetNBClient()

	poolName := args[0]

	err := nbClient.DeletePoolAPI(nb.DeletePoolParams{
		Name: poolName,
	})
	if rpcErr, isRPCErr := err.(*nb.RPCError); isRPCErr {
		switch rpcErr.RPCCode {
		case "NO_SUCH_POOL":
			log.Fatalf(`❌ Can't delete pv-store %s - pool does not exist`, poolName)
		case "DEFAULT_RESOURCE":
			log.Fatalf(`❌ Can't delete pv-store %s - pool is one or more accounts default resource`, poolName)
		case "IN_USE":
			log.Fatalf(`❌ Can't delete pv-store %s - pool is used by one or more buckets`, poolName)
		}
	}
	util.Panic(err)
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	nbClient := system.GetNBClient()
	res, err := nbClient.ReadSystemAPI()
	util.Panic(err)

	table := (&util.PrintTable{}).AddRow("POOL-NAME", "NUM-VOLUMES", "PV-SIZE-GB")
	empty := true
	for i := range res.Pools {
		p := &res.Pools[i]
		if p.ResourceType == "HOSTS" {
			empty = false
			table.AddRow(
				p.Name,
				fmt.Sprintf("%d", p.Hosts.ConfiguredCount),
				fmt.Sprintf("%d", p.HostInfo.VolumeSize/1024/1024/1024),
			)
		}
	}

	if empty {
		fmt.Printf("No PV stores found.\n")
		return
	}

	fmt.Print(table.String())
}
