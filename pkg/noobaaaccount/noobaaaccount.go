package noobaaaccount

import (
	"fmt"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/nb"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/system"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	sigyaml "sigs.k8s.io/yaml"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Manage noobaa accounts",
	}
	cmd.AddCommand(
		CmdCreate(),
		CmdRegenerate(),
		CmdPasswd(),
		CmdDelete(),
		CmdStatus(),
		CmdList(),
		CmdReconcile(),
	)
	return cmd
}

// CmdCreate returns a CLI command
func CmdCreate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create <noobaa-account-name>",
		Short: "Create noobaa account",
		Run:   RunCreate,
	}
	cmd.Flags().Bool("allow_bucket_create", true,
		"Should this account be allowed to create new buckets")
	cmd.Flags().Bool("full_permission", false,
		"Should this account be allowed to access all the buckets (including future ones)")
	cmd.Flags().StringSlice("allowed_buckets", nil,
		"Set the user allowed buckets list (use commas or multiple flags)")
	cmd.Flags().String("default_resource", "", "Set the default resource, on which new buckets will be created")
	cmd.Flags().Bool("nsfs_account_config", false, "This flag is for creating nsfs account")
	cmd.Flags().Int("uid", -1, "Set the nsfs uid")
	cmd.Flags().Int("gid", -1, "Set the nsfs gid")
	cmd.Flags().String("new_buckets_path", "/", "Change the path where new buckets will be created")
	cmd.Flags().Bool("nsfs_only", true, "Set if this account is used only for nsfs")
	return cmd
}

// CmdRegenerate returns a CLI command
func CmdRegenerate() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regenerate <noobaa-account-name>",
		Short: "Regenerate S3 Credentials",
		Run:   RunRegenerate,
	}
	return cmd
}

// CmdPasswd returns a CLI command
func CmdPasswd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "passwd <noobaa-account-name>",
		Short: "reset password for noobaa account",
		Run:   RunPasswd,
	}
	cmd.Flags().String(
		"old-password", "",
		`Old Password for authentication - the best practice is to **omit this flag**, in that case the CLI will prompt to prompt and read it securely from the terminal to avoid leaking secrets in the shell history`,
	)
	cmd.Flags().String(
		"new-password", "",
		`New Password for authentication - the best practice is to **omit this flag**, in that case the CLI will prompt to prompt and read it securely from the terminal to avoid leaking secrets in the shell history`,
	)
	cmd.Flags().String(
		"retype-new-password", "",
		`Retype new Password for authentication - the best practice is to **omit this flag**, in that case the CLI will prompt to prompt and read it securely from the terminal to avoid leaking secrets in the shell history`,
	)
	return cmd
}

// CmdDelete returns a CLI command
func CmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete <noobaa-account-name>",
		Short: "Delete noobaa account",
		Run:   RunDelete,
	}
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status <noobaa-account-name>",
		Short: "Status noobaa account",
		Run:   RunStatus,
	}
	return cmd
}

// CmdList returns a CLI command
func CmdList() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List noobaa accounts",
		Run:   RunList,
	}
	return cmd
}

// CmdReconcile returns a CLI command
func CmdReconcile() *cobra.Command {
	cmd := &cobra.Command{
		Hidden: true,
		Use:    "reconcile",
		Short:  "Runs a reconcile attempt like noobaa-operator",
		Run:    RunReconcile,
	}
	return cmd
}

// RunCreate runs a CLI command
func RunCreate(cmd *cobra.Command, args []string) {

	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <noobaa-account-name> %s`, cmd.UsageString())
	}
	name := args[0]

	allowedBuckets := []string{}
	fullPermission, _ := cmd.Flags().GetBool("full_permission")
	bucketList, _ := cmd.Flags().GetStringSlice("allowed_buckets")

	allowedBuckets = append(allowedBuckets, bucketList...)
	if !fullPermission && len(allowedBuckets) == 0 {
		log.Fatalf(`❌ Must provide at least one allowed buckets, or full_permission`)
	}
	if len(allowedBuckets) > 0 && fullPermission {
		log.Fatalf(`❌ Can't provide both full_permission and an allowed buckets list`)
	}

	allowBucketCreate, _ := cmd.Flags().GetBool("allow_bucket_create")
	defaultResource, _ := cmd.Flags().GetString("default_resource")

	nsfsAccountConfig, _ := cmd.Flags().GetBool("nsfs_account_config")

	newBucketsPath, _ := cmd.Flags().GetString("new_buckets_path")
	nsfsOnly, _ := cmd.Flags().GetBool("nsfs_only")

	// Check and get system
	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml)
	sys := o.(*nbv1.NooBaa)
	sys.Name = options.SystemName
	sys.Namespace = options.Namespace

	o = util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml)
	noobaaAccount := o.(*nbv1.NooBaaAccount)
	noobaaAccount.Name = name
	noobaaAccount.Namespace = options.Namespace
	noobaaAccount.Spec.AllowBucketCreate = allowBucketCreate
	noobaaAccount.Spec.AllowedBuckets.FullPermission = fullPermission
	noobaaAccount.Spec.AllowedBuckets.PermissionList = allowedBuckets

	if nsfsAccountConfig {
		nsfsUID := util.GetFlagIntOrPrompt(cmd, "uid")
		if nsfsUID < 0 {
			log.Fatalf(`❌  uid must be a whole positive number`)
		}

		nsfsGID := util.GetFlagIntOrPrompt(cmd, "gid")
		if nsfsGID < 0 {
			log.Fatalf(`❌  gid must be a whole positive number`)
		}

		noobaaAccount.Spec.NsfsAccountConfig = &nbv1.AccountNsfsConfig{
			UID:            nsfsUID,
			GID:            nsfsGID,
			NewBucketsPath: newBucketsPath,
			NsfsOnly:       nsfsOnly,
		}
	}

	if !util.KubeCheck(sys) {
		log.Fatalf(`❌ Could not find NooBaa system %q in namespace %q`, sys.Name, sys.Namespace)
	}

	if defaultResource == "" { // if user doesn't provide default resource we will use the default backingstore
		defaultResource = sys.Name + "-default-backing-store"
	}

	isResourceBackingStore := checkResourceBackingStore(defaultResource)
	isResourceNamespaceStore := checkResourceNamespaceStore(defaultResource)

	if isResourceBackingStore && isResourceNamespaceStore {
		log.Fatalf(`❌  got BackingStore and NamespaceStore %q in namespace %q`,
			defaultResource, options.Namespace)
	} else if !isResourceBackingStore && !isResourceNamespaceStore {
		log.Fatalf(`❌ Could not get BackingStore or NamespaceStore %q in namespace %q`,
			defaultResource, options.Namespace)
	}

	noobaaAccount.Spec.DefaultResource = defaultResource

	err := util.KubeClient().Get(util.Context(), util.ObjectKey(noobaaAccount), noobaaAccount)
	if err == nil {
		log.Fatalf(`❌ noobaaAccount %q already exists in namespace %q`, noobaaAccount.Name, noobaaAccount.Namespace)
	}

	// Create noobaa account CR
	util.Panic(controllerutil.SetControllerReference(sys, noobaaAccount, scheme.Scheme))
	if !util.KubeCreateSkipExisting(noobaaAccount) {
		log.Fatalf(`❌ Could not create noobaaAccount %q in Namespace %q (conflict)`, noobaaAccount.Name, noobaaAccount.Namespace)
	}

	log.Printf("")
	util.PrintThisNoteWhenFinishedApplyingAndStartWaitLoop()
	log.Printf("")
	log.Printf("NooBaaAccount Wait Ready:")
	if WaitReady(noobaaAccount) {
		log.Printf("")
		log.Printf("")
		RunStatus(cmd, args)
	}
}

// RunRegenerate runs a CLI command
func RunRegenerate(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <noobaa-account-name> %s`, cmd.UsageString())
	}

	var decision string
	log.Printf("You are about to regenerate an account's security credentials.")
	log.Printf("This will invalidate all connections between S3 clients and NooBaa which are connected using the current credentials.")
	log.Printf("are you sure? y/n")

	for {
		fmt.Scanln(&decision)
		if decision == "y" {
			break
		} else if decision == "n" {
			return
		}
	}

	name := args[0]

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml)
	noobaaAccount := o.(*nbv1.NooBaaAccount)
	noobaaAccount.Name = name
	noobaaAccount.Namespace = options.Namespace

	if !util.KubeCheck(noobaaAccount) && (name != "admin@noobaa.io") {
		err := GenerateNonCrdAccountKeys(name)
		if err != nil {
			log.Fatalf(`❌ Could not regenerate credentials for %q: %v`, name, err)
		}
	} else {
		err := GenerateAccountKeys(name)
		if err != nil {
			log.Fatalf(`❌ Could not regenerate credentials for %q: %v`, name, err)
		}

		RunStatus(cmd, args)
	}

}

// RunPasswd runs a CLI command
func RunPasswd(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <noobaa-account-name> %s`, cmd.UsageString())
	}

	name := args[0]

	oldPassword := util.GetFlagStringOrPromptPassword(cmd, "old-password")
	newPassword := util.GetFlagStringOrPromptPassword(cmd, "new-password")
	retypeNewPassword := util.GetFlagStringOrPromptPassword(cmd, "retype-new-password")

	secret := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)

	if name == "admin@noobaa.io" {
		secret.Name = "noobaa-admin"
	} else {
		secret.Name = fmt.Sprintf("noobaa-account-%s", name)
	}
	secret.Namespace = options.Namespace
	if !util.KubeCheck(secret) {
		log.Fatalf(`❌  Could not find secret: %s, will not reset password`, secret.Name)
	}

	if oldPassword != secret.StringData["password"] {
		log.Fatalf(`❌  Password is incorrect, aborting.`)
	}

	err := ResetPassword(name, oldPassword, newPassword, retypeNewPassword)
	if err != nil {
		log.Fatalf(`❌ Could not reset password for %q: %v`, name, err)
	}

	secret.StringData = map[string]string{}
	secret.StringData["password"] = newPassword

	//If we will not be able to update the secret we will print the credentials as they allready been changed by the RPC
	if !util.KubeUpdate(secret) {
		log.Fatalf(`❌  Failed to update the secret %s with the new password, please write it down.`, secret.Name)
	}

	log.Printf("✅ Successfully reset the password for the account %q", name)

}

// RunDelete runs a CLI command
func RunDelete(cmd *cobra.Command, args []string) {

	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <bucket-class-name> %s`, cmd.UsageString())
	}

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml)
	noobaaAccount := o.(*nbv1.NooBaaAccount)
	noobaaAccount.Name = args[0]
	noobaaAccount.Namespace = options.Namespace

	if !util.KubeDelete(noobaaAccount) {
		log.Fatalf(`❌ Could not delete NooBaaAccount %q in namespace %q`,
			noobaaAccount.Name, noobaaAccount.Namespace)
	}
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	log := util.Logger()

	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`❌ Missing expected arguments: <noobaa-account-name> %s`, cmd.UsageString())
	}

	name := args[0]

	o := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaaaccount_cr_yaml)
	noobaaAccount := o.(*nbv1.NooBaaAccount)
	secret := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)

	noobaaAccount.Name = name
	secret.Name = fmt.Sprintf("noobaa-account-%s", name)
	noobaaAccount.Namespace = options.Namespace
	secret.Namespace = options.Namespace

	if !util.KubeCheck(noobaaAccount) && (name != "admin@noobaa.io") {
		log.Fatalf(`❌ Could not get NooBaaAccount %q in namespace %q`,
			noobaaAccount.Name, noobaaAccount.Namespace)
	} else if name == "admin@noobaa.io" {
		secret.Name = "noobaa-admin"
	} else {
		CheckPhase(noobaaAccount)

		fmt.Println()
		fmt.Println("# NooBaaAccount spec:")
		output, err := sigyaml.Marshal(noobaaAccount.Spec)
		util.Panic(err)
		fmt.Print(string(output))
		fmt.Println()
	}

	util.KubeCheck(secret)

	fmt.Printf("Connection info:\n")
	credsEnv := ""
	for k, v := range secret.StringData {
		if v != "" {
			//In admin secret there is also the password, email and system that we do not want to print
			if k == "AWS_ACCESS_KEY_ID" || k == "AWS_SECRET_ACCESS_KEY" {
				if options.ShowSecrets {
					fmt.Printf("  %-22s : %s\n", k, v)
				} else {
					fmt.Printf("  %-22s : %s\n", k, nb.MaskedString(v))
				}
				credsEnv += k + "=" + v + " "
			}
		}
	}
	fmt.Println()
}

// WaitReady waits until the system phase changes to ready by the operator
func WaitReady(noobaaAccount *nbv1.NooBaaAccount) bool {
	log := util.Logger()
	klient := util.KubeClient()

	intervalSec := time.Duration(3)

	err := wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		err := klient.Get(util.Context(), util.ObjectKey(noobaaAccount), noobaaAccount)
		if err != nil {
			log.Printf("⏳ Failed to get NooBaaAccount: %s", err)
			return false, nil
		}
		CheckPhase(noobaaAccount)
		if noobaaAccount.Status.Phase == nbv1.NooBaaAccountPhaseRejected {
			return false, fmt.Errorf("NooBaaAccountPhaseRejected")
		}
		if noobaaAccount.Status.Phase != nbv1.NooBaaAccountPhaseReady {
			return false, nil
		}
		return true, nil
	})
	return err == nil
}

// CheckPhase prints the phase and reason for it
func CheckPhase(noobaaAccount *nbv1.NooBaaAccount) {
	log := util.Logger()

	reason := "waiting..."
	for _, c := range noobaaAccount.Status.Conditions {
		if c.Type == "Available" {
			reason = fmt.Sprintf("%s %s", c.Reason, c.Message)
		}
	}

	switch noobaaAccount.Status.Phase {

	case nbv1.NooBaaAccountPhaseReady:
		log.Printf("✅ NooBaaAccount %q Phase is Ready", noobaaAccount.Name)

	case nbv1.NooBaaAccountPhaseRejected:
		log.Errorf("❌ NooBaaAccount %q Phase is %q: %s", noobaaAccount.Name, noobaaAccount.Status.Phase, reason)

	case nbv1.NooBaaAccountPhaseVerifying:
		fallthrough
	case nbv1.NooBaaAccountPhaseDeleting:
		fallthrough
	default:
		log.Printf("⏳ NooBaaAccount %q Phase is %q: %s", noobaaAccount.Name, noobaaAccount.Status.Phase, reason)
	}
}

// RunList runs a CLI command
func RunList(cmd *cobra.Command, args []string) {
	list := &nbv1.NooBaaAccountList{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaaAccountList"},
	}
	if !util.KubeList(list, &client.ListOptions{Namespace: options.Namespace}) {
		return
	}
	if len(list.Items) == 0 {
		fmt.Printf("No noobaa accounts found.\n")
		return
	}
	table := (&util.PrintTable{}).AddRow(
		"NAME",
		"ALLOWED_BUCKETS",
		"DEFAULT_RESOURCE",
		"PHASE",
		"AGE",
	)
	for i := range list.Items {
		na := &list.Items[i]
		var allowedBuckets []string
		if na.Spec.AllowedBuckets.FullPermission {
			allowedBuckets = append(allowedBuckets, "*")
		} else {
			allowedBuckets = na.Spec.AllowedBuckets.PermissionList
		}
		defaultResource := na.Spec.DefaultResource
		if !na.Spec.AllowBucketCreate {
			defaultResource = "-NO-BUCKET-CREATION-"
		}
		table.AddRow(
			na.Name,
			fmt.Sprintf("%+v", allowedBuckets),
			defaultResource,
			string(na.Status.Phase),
			time.Since(na.CreationTimestamp.Time).Round(time.Second).String(),
		)
	}
	fmt.Print(table.String())
}

// RunReconcile runs a CLI command
func RunReconcile(cmd *cobra.Command, args []string) {
	log := util.Logger()
	if len(args) != 1 || args[0] == "" {
		log.Fatalf(`Missing expected arguments: <bucket-name> %s`, cmd.UsageString())
	}
	noobaaAccountName := args[0]
	klient := util.KubeClient()
	intervalSec := time.Duration(3)
	util.Panic(wait.PollImmediateInfinite(intervalSec*time.Second, func() (bool, error) {
		req := reconcile.Request{
			NamespacedName: types.NamespacedName{
				Namespace: options.Namespace,
				Name:      noobaaAccountName,
			},
		}
		res, err := NewReconciler(req.NamespacedName, klient, scheme.Scheme, nil).Reconcile()
		if err != nil {
			return false, err
		}
		if res.Requeue || res.RequeueAfter != 0 {
			log.Printf("\nRetrying in %d seconds\n", intervalSec)
			return false, nil
		}
		return true, nil
	}))
}

// GenerateAccountKeys regenerate noobaa account (CRD based) S3 keys
func GenerateAccountKeys(name string) error {
	log := util.Logger()

	var accessKeys nb.S3AccessKeys

	sysClient, err := system.Connect(true)
	if err != nil {
		return err
	}

	// Checking that we can find the secret before we are calling the RPC to change the credentials.
	secret := util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)
	secret.Namespace = options.Namespace
	// Handling a special case when the account is "admin@noobaa.io" we don't have CRD but have a secret
	if name == "admin@noobaa.io" {
		secret.Name = "noobaa-admin"
	} else {
		secret.Name = fmt.Sprintf("noobaa-account-%s", name)
	}
	if !util.KubeCheckQuiet(secret) {
		log.Fatalf(`❌  Could not find secret: %s, will not regenerate keys.`, secret.Name)
	}

	err = sysClient.NBClient.GenerateAccountKeysAPI(nb.GenerateAccountKeysParams{
		Email: name,
	})
	if err != nil {
		return err
	}

	// GenerateAccountKeysAPI have no replay so we need to read the account in order to get the new credentials
	accountInfo, err := sysClient.NBClient.ReadAccountAPI(nb.ReadAccountParams{
		Email: name,
	})
	if err != nil {
		return err
	}

	accessKeys = accountInfo.AccessKeys[0]

	secret.StringData = map[string]string{}
	secret.StringData["AWS_ACCESS_KEY_ID"] = accessKeys.AccessKey
	secret.StringData["AWS_SECRET_ACCESS_KEY"] = accessKeys.SecretKey

	//If we will not be able to update the secret we will print the credentials as they allready been changed by the RPC
	if !util.KubeUpdate(secret) {
		log.Printf(`❌  Please write the new credentials for account %s:`, name)
		fmt.Printf("\nAWS_ACCESS_KEY_ID     : %s\n", accessKeys.AccessKey)
		fmt.Printf("AWS_SECRET_ACCESS_KEY : %s\n\n", accessKeys.SecretKey)
		log.Fatalf(`❌  Failed to update the secret %s with the new accessKeys`, secret.Name)
	}

	log.Printf("✅ Successfully reganerate s3 credentials for the account %q", name)
	return nil
}

// GenerateNonCrdAccountKeys regenerate noobaa account (none CRD based) S3 keys
func GenerateNonCrdAccountKeys(name string) error {
	log := util.Logger()

	var accessKeys nb.S3AccessKeys

	sysClient, err := system.Connect(true)
	if err != nil {
		return err
	}

	err = sysClient.NBClient.GenerateAccountKeysAPI(nb.GenerateAccountKeysParams{
		Email: name,
	})
	if err != nil {
		if nbErr, ok := err.(*nb.RPCError); ok && nbErr.RPCCode == "NO_SUCH_ACCOUNT" {
			log.Fatalf(`❌  Could not find the account: %s, will not regenerate keys.`, name)
		}
		return err
	}

	// GenerateAccountKeysAPI have no replay so we need to read the account in order to get the new credentials
	accountInfo, err := sysClient.NBClient.ReadAccountAPI(nb.ReadAccountParams{
		Email: name,
	})
	if err != nil {
		log.Fatalf(`❌  Could not read account: %s, keys were allready regenerated, please read the account to get the keys`, name)
	}

	accessKeys = accountInfo.AccessKeys[0]

	log.Printf("✅ Successfully reganerate s3 credentials for the account %q", name)
	log.Printf(`✅  Please write the new credentials for account %s:`, name)
	fmt.Printf("\nAWS_ACCESS_KEY_ID     : %s\n", accessKeys.AccessKey)
	fmt.Printf("AWS_SECRET_ACCESS_KEY : %s\n\n", accessKeys.SecretKey)

	return nil
}

// ResetPassword reset noobaa account password
func ResetPassword(name string, oldPassword string, newPassword string, retypeNewPassword string) error {
	sysClient, err := system.Connect(true)
	if err != nil {
		return err
	}

	PasswordResstrictions(oldPassword, newPassword, retypeNewPassword)

	err = sysClient.NBClient.ResetPasswordAPI(nb.ResetPasswordParams{
		Email:                name,
		VerificationPassword: nb.MaskedString(oldPassword),
		Password:             nb.MaskedString(newPassword),
	})
	if err != nil {
		return err
	}

	return nil
}

// PasswordResstrictions checks for all kind of password restrictions
func PasswordResstrictions(oldPassword string, newPassword string, retypeNewPassword string) {
	log := util.Logger()

	//Checking that we did not get the same password as the old one
	if newPassword == oldPassword {
		log.Fatalf(`❌  The password cannot match the old password, aborting.`)
	}

	//Checking that we got the same password twice
	if newPassword != retypeNewPassword {
		log.Fatalf(`❌  The password and is not matching the retype, aborting.`)
	}

	//TODO... This is the place for adding more restrictions
	// length of password
	// charecters

}

// checkResourceBackingStore checks if a resourceName exists and if BackingStore
func checkResourceBackingStore(resourceName string) bool {
	// check that a backing store exists
	resourceBackingStore := &nbv1.BackingStore{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStore"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: options.Namespace,
		},
	}

	return util.KubeCheckQuiet(resourceBackingStore)
}

// checkResourceNamespaceStore checks if a resourceName exists and if NamespaceStore
func checkResourceNamespaceStore(resourceName string) bool {
	// check that a namespace store exists
	resourceNamespaceStore := &nbv1.NamespaceStore{
		TypeMeta: metav1.TypeMeta{Kind: "NamespaceStore"},
		ObjectMeta: metav1.ObjectMeta{
			Name:      resourceName,
			Namespace: options.Namespace,
		},
	}

	return util.KubeCheckQuiet(resourceNamespaceStore)
}
