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
		CmdReport(),
		CmdLogAnalysis(),
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

// CmdReport returns a CLI command
func CmdReport() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "report",
		Short: "Run reports of the status and setup",
		Run:   RunReport,
		Args:  cobra.NoArgs,
	}
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

// CmdLogAnalysis returns a CLI command
func CmdLogAnalysis() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "log-analysis",
		Short: "Run the log analyzer on all NooBaa pods",
		Run:   RunLogAnalysis,
	}
	cmd.Flags().BoolP("verbose", "v", false, "Print every matching log line")
	cmd.Flags().Bool("fuzzy", false, "(Experimental) Toggle fuzzy matching for the search string")
	cmd.Flags().Int64("tail", 1000, "Number of lines to tail from the logs, minimum 1, default 1000")
	cmd.Flags().BoolP("case-insensitive", "i", false, "Toggle search-string case insensitivity (similar to grep's -i flag)")
	cmd.Flags().BoolP("whole-string", "w", false, "Match the whole search string as a single word (similar to grep's -w flag)")
	cmd.Flags().Bool("prefer-line", false, "Prefer to print the line containing the search string when it doesn't contain a timestamp")
	cmd.Flags().Bool("noobaa-time", false, "Use NooBaa-provided timestamps instead of the Kubernetes ones (~10ms earlier than Kubernetes)")


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
