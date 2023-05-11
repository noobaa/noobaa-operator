// Package diagnostics implements the functions we use to save logs and status of our system.
package diagnostics

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

// Collector configuration for diagnostics
type Collector struct {
	folderName  string
	kubeconfig  string
	kubeCommand string
	log         *logrus.Entry
}

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "diagnostics",
		Short: "diagnostics of items in noobaa system",
	}
	cmd.AddCommand(
		CmdCollect(),
		CmdDbDump(),
		CmdAnalyze(),
	)
	return cmd
}

// CmdCollect returns a CLI command
func CmdCollect() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "collect",
		Short: "Collect diagnostics",
		Run:   RunCollect,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().String("dir", "", "collect noobaa diagnostics tar file into destination directory")
	cmd.Flags().Bool("db-dump", false, "collect db dump in addition to diagnostics")
	return cmd
}

// CmdDbDump returns a CLI command
func CmdDbDump() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db-dump",
		Short: "Collect db dump",
		Run:   RunDump,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().String("dir", "", "collect db dump file into destination directory")
	return cmd
}

// CmdAnalyze returns a CLI command
func CmdAnalyze() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "analyze",
		Short: "Analyze the resource status by running tests on it",
	}
	cmd.AddCommand(
		CmdAnalyzeBackingStore(),
		CmdAnalyzeNamespaceStore(),
		CmdAnalyzeResources(),
	)
	return cmd
}

// CmdAnalyzeBackingStore returns a CLI command
func CmdAnalyzeBackingStore() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backingstore <backing-store-name>",
		Short: "Analyze backingstore",
		Run:   RunAnalyzeBackingStore,
	}
	cmd.Flags().String(
		"bucket", "",
		"The bucket name on the cloud",
	)
	cmd.Flags().String("job-resources", "", "Analyze job resources JSON")
	cmd.Flags().String("dir", "", "collect analyze resource tar file into destination directory")
	return cmd
}

// CmdAnalyzeNamespaceStore returns a CLI command
func CmdAnalyzeNamespaceStore() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "namespacestore <namespace-store-name>",
		Short: "Analyze namespacestore",
		Run:   RunAnalyzeNamespaceStore,
	}
	cmd.Flags().String(
		"bucket", "",
		"The bucket name on the cloud",
	)
	cmd.Flags().String("job-resources", "", "Analyze job resources JSON")
	cmd.Flags().String("dir", "", "collect analyze resource tar file into destination directory")
	return cmd
}

// CmdAnalyzeResources returns a CLI command
func CmdAnalyzeResources() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resources",
		Short: "Analyze all resources (backingstores and namespacestores)",
		Run:   RunAnalyzeResources,
	}
	cmd.Flags().String("job-resources", "", "Analyze job resources JSON")
	cmd.Flags().String("dir", "", "collect analyze resource tar file into destination directory")
	return cmd
}

/////// Deprecated Functions ///////

// CmdDbDumpDeprecated returns a CLI command
func CmdDbDumpDeprecated() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "db-dump",
		Short:      "Collect db dump",
		Deprecated: "please use diagnostics db-dump",
		Run:        RunDump,
		Args:       cobra.NoArgs,
	}
	cmd.Flags().String("dir", "", "collect db dump file into destination directory")
	return cmd
}

// CmdDiagnoseDeprecated returns a CLI command
func CmdDiagnoseDeprecated() *cobra.Command {
	cmd := &cobra.Command{
		Use:        "diagnose",
		Short:      "Collect diagnostics",
		Deprecated: "please use diagnostics collect",
		Run:        RunCollect,
		Args:       cobra.NoArgs,
	}
	cmd.Flags().String("dir", "", "collect noobaa diagnose tar file into destination directory")
	cmd.Flags().Bool("db-dump", false, "collect db dump in addition to diagnostics")
	return cmd
}
