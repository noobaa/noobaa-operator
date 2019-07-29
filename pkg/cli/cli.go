package cli

import (
	"context"
	"flag"
	"math/rand"
	"os"
	"time"

	"github.com/noobaa/noobaa-operator/pkg/apis"
	"github.com/noobaa/noobaa-operator/pkg/controller"
	"github.com/noobaa/noobaa-operator/pkg/system"
	"github.com/noobaa/noobaa-operator/pkg/util"
	"github.com/noobaa/noobaa-operator/version"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	apiextv1beta1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1beta1"
	"k8s.io/client-go/kubernetes/scheme"
	clientcmd "k8s.io/client-go/tools/clientcmd"
	"k8s.io/kubectl/pkg/util/templates"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

// CLI is the common tools used for most of this binary's commands.
// It creates the command structure and defines all the methods that
// implement the CLI commands.
type CLI struct {
	Client client.Client
	Ctx    context.Context
	Log    *logrus.Entry

	Namespace        string
	SystemName       string
	StorageClassName string
	NooBaaImage      string
	OperatorImage    string
	ImagePullSecret  string

	// Commands
	Cmd          *cobra.Command
	CmdOptions   *cobra.Command
	CmdVersion   *cobra.Command
	CmdInstall   *cobra.Command
	CmdUninstall *cobra.Command
	CmdStatus    *cobra.Command
	CmdBucket    *cobra.Command
	CmdCrd       *cobra.Command
	CmdOlmHub    *cobra.Command
	CmdOlmLocal  *cobra.Command
	CmdOperator  *cobra.Command
	CmdSystem    *cobra.Command
}

func New() *CLI {

	util.InitLogger()

	rand.Seed(time.Now().UTC().UnixNano())

	logo := ASCIILogo1
	if rand.Intn(2) == 0 { // 50% chance
		logo = ASCIILogo2
	}

	ctx := context.TODO()

	cli := &CLI{
		Ctx: ctx,
		Log: logrus.WithContext(ctx),

		Namespace:        "noobaa", //CurrentNamespace(),
		SystemName:       "noobaa",
		StorageClassName: "",
		NooBaaImage:      system.ContainerImage,
		OperatorImage:    "noobaa/noobaa-operator:" + version.Version,
		ImagePullSecret:  "",

		// Root command
		Cmd: &cobra.Command{
			Use:   "noobaa",
			Short: logo,
		},

		// Install Commands:
		CmdInstall: &cobra.Command{
			Use:   "install",
			Short: "Install the operator and create the noobaa system",
		},
		CmdUninstall: &cobra.Command{
			Use:   "uninstall",
			Short: "Uninstall the operator and delete the system",
		},
		CmdStatus: &cobra.Command{
			Use:   "status",
			Short: "Status of the operator and the system",
		},

		// Manage Commands:
		CmdBucket: &cobra.Command{
			Use:   "bucket",
			Short: "Manage noobaa buckets",
		},

		// Advanced Commands:

		CmdCrd: &cobra.Command{
			Use:   "crd",
			Short: "Deployment of CRDs",
		},
		CmdOlmHub: &cobra.Command{
			Use:   "olm-hub",
			Short: "Deployment using operatorhub.io",
		},
		CmdOlmLocal: &cobra.Command{
			Use:   "olm-local",
			Short: "Deployment using OLM",
		},
		CmdOperator: &cobra.Command{
			Use:   "operator",
			Short: "Deployment using operator",
		},
		CmdSystem: &cobra.Command{
			Use:   "system",
			Short: "Manage noobaa systems (create delete etc.)",
		},

		// Other Commands:
		CmdOptions: &cobra.Command{
			Use:   "options",
			Short: "Print the list of flags inherited by all commands",
		},
		CmdVersion: &cobra.Command{
			Use:   "version",
			Short: "Show version",
		},
	}

	cli.CmdVersion.Run = ToRunnable(cli.Version)
	cli.CmdInstall.Run = ToRunnable(cli.Install)
	cli.CmdUninstall.Run = ToRunnable(cli.Uninstall)
	cli.CmdStatus.Run = ToRunnable(cli.Status)
	cli.CmdOptions.Run = ToRunnable(func() {
		cli.CmdOptions.Usage()
	})

	cli.CmdBucket.AddCommand(
		&cobra.Command{
			Use:   "create",
			Short: "Create a NooBaa bucket",
			Run:   ToRunnableArgs(cli.BucketCreate),
		},
		&cobra.Command{
			Use:   "delete",
			Short: "Delete a NooBaa bucket",
			Run:   ToRunnableArgs(cli.BucketDelete),
		},
		&cobra.Command{
			Use:   "list",
			Short: "List NooBaa buckets",
			Run:   ToRunnable(cli.BucketList),
		},
	)

	cli.CmdCrd.AddCommand(
		&cobra.Command{
			Use:   "create",
			Short: "Create noobaa CRDs",
			Run:   ToRunnable(cli.CrdsCreate),
		},
		&cobra.Command{
			Use:   "delete",
			Short: "Delete noobaa CRDs",
			Run:   ToRunnable(cli.CrdsDelete),
		},
		&cobra.Command{
			Use:   "status",
			Short: "Status of noobaa CRDs",
			Run:   ToRunnable(cli.CrdsStatus),
		},
		&cobra.Command{
			Use:   "yaml",
			Short: "Show bundled CRDs",
			Run:   ToRunnable(cli.CrdsYaml),
		},
	)

	cli.CmdOlmHub.AddCommand(
		&cobra.Command{
			Use:   "install",
			Short: "Install noobaa-operator from operatorhub.io",
			Run:   ToRunnable(cli.HubInstall),
		},
		&cobra.Command{
			Use:   "uninstall",
			Short: "Uninstall noobaa-operator from operatorhub.io",
			Run:   ToRunnable(cli.HubUninstall),
		},
		&cobra.Command{
			Use:   "status",
			Short: "Status of noobaa-operator from operatorhub.io",
			Run:   ToRunnable(cli.HubStatus),
		},
	)

	cli.CmdOlmLocal.AddCommand(
		&cobra.Command{
			Use:   "install",
			Short: "Install noobaa-operator",
			Run:   ToRunnable(cli.OperatorInstall),
		},
		&cobra.Command{
			Use:   "uninstall",
			Short: "Uninstall noobaa-operator",
			Run:   ToRunnable(cli.OperatorUninstall),
		},
	)

	cli.CmdOperator.AddCommand(
		&cobra.Command{
			Use:   "install-local",
			Short: "Install the resources needed for local operator",
			Run:   ToRunnable(cli.OperatorLocalInstall),
		},
		&cobra.Command{
			Use:   "uninstall-local",
			Short: "Uninstall the resources needed for local operator",
			Run:   ToRunnable(cli.OperatorLocalUninstall),
		},
		&cobra.Command{
			Use:   "reconcile-local",
			Short: "Runs a reconcile attempt like noobaa-operator",
			Run:   ToRunnable(cli.OperatorLocalReconcile),
		},
		&cobra.Command{
			Use:   "install",
			Short: "Install noobaa-operator",
			Run:   ToRunnable(cli.OperatorInstall),
		},
		&cobra.Command{
			Use:   "uninstall",
			Short: "Uninstall noobaa-operator",
			Run:   ToRunnable(cli.OperatorUninstall),
		},
		&cobra.Command{
			Use:   "status",
			Short: "Status of a noobaa-operator",
			Run:   ToRunnable(cli.OperatorStatus),
		},
		&cobra.Command{
			Use:   "run",
			Short: "Runs the noobaa-operator",
			Run:   ToRunnable(controller.OperatorMain),
		},
		&cobra.Command{
			Use:   "yaml",
			Short: "Show bundled noobaa-operator yaml",
			Run:   ToRunnable(cli.OperatorYamls),
		},
	)

	cli.CmdSystem.AddCommand(
		&cobra.Command{
			Use:   "create",
			Short: "Create a noobaa system",
			Run:   ToRunnable(cli.SystemCreate),
		},
		&cobra.Command{
			Use:   "delete",
			Short: "Delete a noobaa system",
			Run:   ToRunnable(cli.SystemDelete),
		},
		&cobra.Command{
			Use:   "list",
			Short: "List noobaa systems",
			Run:   ToRunnable(cli.SystemList),
		},
		&cobra.Command{
			Use:   "status",
			Short: "Status of a noobaa system",
			Run:   ToRunnable(cli.SystemStatus),
		},
		&cobra.Command{
			Use:   "yaml",
			Short: "Show bundled noobaa yaml",
			Run:   ToRunnable(cli.SystemYaml),
		},
	)

	flagset := cli.Cmd.PersistentFlags()
	// flagset.AddFlagSet(zap.FlagSet())
	flagset.AddGoFlagSet(flag.CommandLine)
	flagset.StringVarP(
		&cli.Namespace, "namespace", "n",
		cli.Namespace, "Target namespace",
	)
	flagset.StringVarP(
		&cli.SystemName, "system-name", "N",
		cli.SystemName, "NooBaa system name",
	)
	flagset.StringVar(
		&cli.StorageClassName, "storage-class",
		cli.StorageClassName, "Storage class name",
	)
	flagset.StringVar(
		&cli.NooBaaImage, "noobaa-image",
		cli.NooBaaImage, "NooBaa image",
	)
	flagset.StringVar(
		&cli.OperatorImage, "operator-image",
		cli.OperatorImage, "Operator image",
	)
	flagset.StringVar(
		&cli.ImagePullSecret, "image-pull-secret",
		cli.ImagePullSecret, "Image pull secret (must be in same namespace)",
	)

	groups := templates.CommandGroups{
		{
			Message: "Install:",
			Commands: []*cobra.Command{
				cli.CmdInstall,
				cli.CmdStatus,
				cli.CmdUninstall,
			},
		},
		{
			Message: "Manage:",
			Commands: []*cobra.Command{
				cli.CmdBucket,
			},
		},
		{
			Message: "Advanced:",
			Commands: []*cobra.Command{
				cli.CmdCrd,
				cli.CmdOlmHub,
				cli.CmdOlmLocal,
				cli.CmdOperator,
				cli.CmdSystem,
			},
		},
	}

	groups.Add(cli.Cmd)
	cli.Cmd.AddCommand(
		cli.CmdVersion,
		cli.CmdOptions,
	)
	templates.ActsAsRootCommand(cli.Cmd, []string{}, groups...)
	templates.UseOptionsTemplates(cli.CmdOptions)
	return cli
}

// Run the CLI - the main
func (cli *CLI) Run() {
	util.Panic(apiextv1beta1.AddToScheme(scheme.Scheme))
	util.Panic(apis.AddToScheme(scheme.Scheme))
	cmd, args, err := cli.Cmd.Traverse(os.Args[1:])
	util.Panic(err)
	if cmd.Runnable() &&
		cmd != cli.CmdVersion &&
		cmd != cli.CmdOptions {
		cli.Client = util.KubeClient()
		cmd.Run(cmd, args)
	} else {
		cmd.Execute()
	}
}

func (cli *CLI) Version() {
	cli.Log.Printf("CLI version: %s\n", version.Version)
	cli.Log.Printf("noobaa-image: %s\n", cli.NooBaaImage)
	cli.Log.Printf("operator-image: %s\n", cli.OperatorImage)
}

func (cli *CLI) Install() {
	cli.Version()
	cli.Log.Printf("Namespace: %s", cli.Namespace)
	cli.Log.Printf("")
	cli.Log.Printf("CRD Create:")
	cli.CrdsCreate()
	cli.CrdsWaitReady()
	cli.Log.Printf("")
	cli.Log.Printf("Operator Install:")
	cli.OperatorInstall()
	cli.Log.Printf("")
	cli.Log.Printf("System Create:")
	cli.SystemCreate()
	cli.SystemWaitReady()
	cli.Status()
}

func (cli *CLI) Uninstall() {
	cli.Version()
	cli.Log.Printf("Namespace: %s", cli.Namespace)
	cli.Log.Printf("")
	cli.Log.Printf("System Delete:")
	cli.SystemDelete()
	cli.Log.Printf("")
	cli.Log.Printf("Operator Delete:")
	cli.OperatorUninstall()
	cli.Log.Printf("")
	cli.Log.Printf("CRD - Leaving Untouched - Status:")
	cli.Log.Printf("(use \"noobaa crd delete\" to force deletion)")
	cli.CrdsStatus()
}

func (cli *CLI) Status() {
	cli.Version()
	cli.Log.Printf("Namespace: %s", cli.Namespace)
	cli.Log.Printf("")
	cli.Log.Printf("CRD Status:")
	cli.CrdsStatus()
	cli.Log.Printf("")
	cli.Log.Printf("Operator Status:")
	cli.OperatorStatus()
	cli.Log.Printf("")
	cli.Log.Printf("System Status:")
	cli.SystemStatus()
}

func CurrentNamespace() string {
	loadingRules := clientcmd.NewDefaultClientConfigLoadingRules()
	configOverrides := &clientcmd.ConfigOverrides{}
	kubeConfig := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(loadingRules, configOverrides)
	ns, _, err := kubeConfig.Namespace()
	util.Panic(err)
	return ns
}

// Runnable is the type of cobra.Command.Run
type Runnable func(cmd *cobra.Command, args []string)

// ToRunnable adapts functions to Runnable
func ToRunnable(f func()) Runnable {
	return func(cmd *cobra.Command, args []string) {
		f()
	}
}

// ToRunnableArgs adapts functions to Runnable
func ToRunnableArgs(f func(args []string)) Runnable {
	return func(cmd *cobra.Command, args []string) {
		f(args)
	}
}

// ForEachCommand recursivly visits all subcommands
func ForEachCommand(cmd *cobra.Command, handler func(c *cobra.Command)) {
	for _, c := range cmd.Commands() {
		handler(c)
		ForEachCommand(c, handler)
	}
}
