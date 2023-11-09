package cli

import (
	"flag"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/noobaa/noobaa-operator/v5/pkg/backingstore"
	"github.com/noobaa/noobaa-operator/v5/pkg/bucket"
	"github.com/noobaa/noobaa-operator/v5/pkg/bucketclass"
	"github.com/noobaa/noobaa-operator/v5/pkg/cosi"
	"github.com/noobaa/noobaa-operator/v5/pkg/crd"
	"github.com/noobaa/noobaa-operator/v5/pkg/diagnostics"
	"github.com/noobaa/noobaa-operator/v5/pkg/install"
	"github.com/noobaa/noobaa-operator/v5/pkg/namespacestore"
	"github.com/noobaa/noobaa-operator/v5/pkg/noobaaaccount"
	"github.com/noobaa/noobaa-operator/v5/pkg/obc"
	"github.com/noobaa/noobaa-operator/v5/pkg/olm"
	"github.com/noobaa/noobaa-operator/v5/pkg/operator"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/pvstore"
	"github.com/noobaa/noobaa-operator/v5/pkg/sts"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/pkg/version"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"k8s.io/kubectl/pkg/util/templates"
)

// ASCIILogo1 is an ascii logo of noobaa
const ASCIILogo1 = `
._   _            ______              
| \ | |           | ___ \             
|  \| | ___   ___ | |_/ / __ _  __ _  
| . \ |/ _ \ / _ \| ___ \/ _\ |/ _\ | 
| |\  | (_) | (_) | |_/ / (_| | (_| | 
\_| \_/\___/ \___/\____/ \__,_|\__,_| 
`

// ASCIILogo2 is an ascii logo of noobaa
const ASCIILogo2 = `
#                       # 
#    /~~\___~___/~~\    # 
#   |               |   # 
#    \~~|\     /|~~/    # 
#        \|   |/        # 
#         |   |         # 
#         \~~~/         # 
#                       # 
#      N O O B A A      # 
`

// Run runs
func Run() {
	err := Cmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

// Cmd returns a CLI command
func Cmd() *cobra.Command {

	util.InitLogger(logrus.DebugLevel)

	r := rand.New(rand.NewSource(time.Now().UTC().UnixNano()))

	logo := ASCIILogo1
	if r.Intn(2) == 0 { // 50% chance
		logo = ASCIILogo2
	}

	// Root command
	rootCmd := &cobra.Command{
		Use:   "noobaa",
		Short: logo,
	}

	rootCmd.PersistentFlags().AddFlagSet(options.FlagSet)
	rootCmd.PersistentFlags().AddGoFlagSet(flag.CommandLine)

	viperSetup(options.FlagSet)

	optionsCmd := options.Cmd()

	completionCmd := &cobra.Command{
		Use:   "completion",
		Short: "Generates bash completion scripts",
		Long: fmt.Sprintf(`
Load noobaa completion to bash:
(add to your ~/.bashrc and ~/.bash_profile to auto load)

. <(%s completion)
`, rootCmd.Name()),
		Run: func(cmd *cobra.Command, args []string) {
			alias, _ := cmd.Flags().GetString("alias")
			if alias != "" {
				rootCmd.Use = alias
			}
			err := rootCmd.GenBashCompletion(os.Stdout)
			if err != nil {
				fmt.Printf("got error on GenBashCompletion. %v", err)
			}

		},
		Args: cobra.NoArgs,
	}
	completionCmd.Flags().String("alias", "", "Custom alias name to generate the completion for")

	groups := templates.CommandGroups{{
		Message: "Install:",
		Commands: []*cobra.Command{
			install.CmdInstall(),
			install.CmdUpgrade(),
			install.CmdUninstall(),
			install.CmdStatus(),
		},
	}, {
		Message: "Manage:",
		Commands: []*cobra.Command{
			backingstore.Cmd(),
			namespacestore.Cmd(),
			bucketclass.Cmd(),
			noobaaaccount.Cmd(),
			obc.Cmd(),
			cosi.Cmd(),
			diagnostics.CmdDiagnoseDeprecated(),
			diagnostics.CmdDbDumpDeprecated(),
			diagnostics.Cmd(),
			sts.Cmd(),
		},
	}, {
		Message: "Advanced:",
		Commands: []*cobra.Command{
			operator.Cmd(),
			system.Cmd(),
			system.CmdAPICall(),
			bucket.Cmd(),
			pvstore.Cmd(),
			crd.Cmd(),
			olm.Cmd(),
		},
	}}

	groups.Add(rootCmd)

	rootCmd.AddCommand(
		version.Cmd(),
		optionsCmd,
		completionCmd,
	)

	templates.ActsAsRootCommand(rootCmd, []string{}, groups...)
	templates.UseOptionsTemplates(optionsCmd)

	return rootCmd
}

func viperSetup(flagsets ...*pflag.FlagSet) {
	viper.SetConfigName("noobaa.cfg")
	viper.SetConfigType("yaml")
	viper.AddConfigPath("/etc/noobaa/")
	viper.AddConfigPath("$HOME/.noobaa")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			logrus.Warn("failed to read config:", err)
		}
	} else {
		logrus.Info("Using config file:", viper.ConfigFileUsed())
	}

	for _, flagset := range flagsets {
		if err := viper.BindPFlags(flagset); err != nil {
			logrus.Warn("failed to bind flags:", err)
			continue
		}

		flagset.VisitAll(func(flag *pflag.Flag) {
			// Instead of using viper.Get interfaces throughout the codebases
			// we set the value of the flag to the value from viper, so we can use the flag.Value
			// everywhere.
			//
			// # Safety
			// viper.GetString will not panic even if the flag value is not a string because the internal
			// casting is type aware.
			if viper.IsSet(flag.Name) {
				if err := flag.Value.Set(viper.GetString(flag.Name)); err != nil {
					logrus.Warn("failed to set flag value:", err)
				}
			}
		})
	}
}
