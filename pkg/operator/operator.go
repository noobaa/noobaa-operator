package operator

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"os"
	"strings"
	"time"

	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	"github.com/spf13/cobra"
	admissionv1 "k8s.io/api/admissionregistration/v1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/cli-runtime/pkg/printers"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operator",
		Short: "Deployment using operator",
	}
	cmd.AddCommand(
		CmdInstall(),
		CmdUninstall(),
		CmdStatus(),
		CmdYaml(),
		CmdRun(),
	)
	return cmd
}

// CmdInstall returns a CLI command
func CmdInstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install noobaa-operator",
		Run:   RunInstall,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().Bool("no-deploy", false, "Install only the needed resources but do not create the operator deployment")
	return cmd
}

// CmdUninstall returns a CLI command
func CmdUninstall() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall noobaa-operator",
		Run:   RunUninstall,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().Bool("cleanup", false, "Enable deletion of the Namespace")
	return cmd
}

// CmdStatus returns a CLI command
func CmdStatus() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Status of a noobaa-operator",
		Run:   RunStatus,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdYaml returns a CLI command
func CmdYaml() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "yaml",
		Short: "Show bundled noobaa-operator yaml",
		Run:   RunYaml,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// CmdRun returns a CLI command
func CmdRun() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Runs the noobaa-operator",
		Run:   RunOperator,
		Args:  cobra.NoArgs,
	}
	return cmd
}

// RunInstall runs a CLI command
func RunInstall(cmd *cobra.Command, args []string) {
	c := LoadOperatorConf(cmd)
	util.KubeCreateSkipExisting(c.NS)
	util.KubeCreateSkipExisting(c.SA)
	util.KubeCreateSkipExisting(c.SAEndpoint)
	util.KubeCreateSkipExisting(c.Role)
	util.KubeCreateSkipExisting(c.RoleEndpoint)
	util.KubeCreateSkipExisting(c.RoleBinding)
	util.KubeCreateSkipExisting(c.RoleBindingEndpoint)
	util.KubeCreateSkipExisting(c.ClusterRole)
	util.KubeCreateSkipExisting(c.ClusterRoleBinding)

	admission, _ := cmd.Flags().GetBool("admission")
	if admission {
		LoadAdmissionConf(c)
		AdmissionWebhookSetup(c)
		util.KubeCreateSkipExisting(c.WebhookConfiguration)
		util.KubeCreateSkipExisting(c.WebhookSecret)
		util.KubeCreateSkipExisting(c.WebhookService)
		operatorContainer := c.Deployment.Spec.Template.Spec.Containers[0]
		operatorContainer.Env = append(operatorContainer.Env, corev1.EnvVar{
			Name:  "ENABLE_NOOBAA_ADMISSION",
			Value: "true",
		})
		c.Deployment.Spec.Template.Spec.Containers[0].Env = operatorContainer.Env
	}

	noDeploy, _ := cmd.Flags().GetBool("no-deploy")
	if !noDeploy {
		operatorContainer := c.Deployment.Spec.Template.Spec.Containers[0]
		operatorContainer.Env = append(
			operatorContainer.Env,
			corev1.EnvVar{
				Name:  "NOOBAA_CLI_DEPLOYMENT",
				Value: "true",
			},
		)
		c.Deployment.Spec.Template.Spec.Containers[0].Env = operatorContainer.Env
		util.KubeCreateSkipExisting(c.Deployment)
	}
}

func waitForOperatorPodExit() {
	for {
		podsList := &corev1.PodList{}
		listRes := util.KubeList(podsList, client.InNamespace(options.Namespace), client.MatchingLabels{"noobaa-operator": "deployment"})

		// List failure
		if !listRes {
			log.Printf("❌ Can not list pods in %v namespace with noobaa-operator=deployment label, try again.", options.Namespace)
			// Exit condition, list succeded and no operator's pods are found
		} else if len(podsList.Items) == 0 {
			log.Printf("✅ NooBaa operator pod is not running, continue.")
			break
		}

		// Operator pod is still running
		log.Printf("⏳ Waiting for the operator's pod to exit, list pods result %v, items len %v", listRes, len(podsList.Items))
		time.Sleep(1 * time.Second)
	}
}

// RunUninstall runs a CLI command
func RunUninstall(cmd *cobra.Command, args []string) {
	c := LoadOperatorConf(cmd)
	LoadAdmissionConf(c)
	noDeploy, _ := cmd.Flags().GetBool("no-deploy")
	if !noDeploy {
		util.KubeDelete(c.Deployment)
		waitForOperatorPodExit()
		util.KubeDelete(c.ClusterRoleBinding)
		util.KubeDelete(c.ClusterRole)
		util.KubeDelete(c.RoleBindingEndpoint)
		util.KubeDelete(c.RoleBinding)
		util.KubeDelete(c.RoleEndpoint)
		util.KubeDelete(c.Role)
		util.KubeDelete(c.SAEndpoint)
		util.KubeDelete(c.SA)
	} else {
		log.Printf("Operator Delete: currently disabled with \"--no-deploy\" flag")
		log.Printf("Operator Deployment Status:")
		util.KubeCheck(c.Deployment)
	}

	util.KubeDelete(c.WebhookConfiguration)
	util.KubeDelete(c.WebhookSecret)
	util.KubeDelete(c.WebhookService)

	cleanup, _ := cmd.Flags().GetBool("cleanup")
	reservedNS := c.NS.Name == "default" ||
		strings.HasPrefix(c.NS.Name, "openshift-") ||
		strings.HasPrefix(c.NS.Name, "kubernetes-") ||
		strings.HasPrefix(c.NS.Name, "kube-")
	if reservedNS {
		log.Printf("Namespace Delete: disabled for reserved namespace")
		log.Printf("Namespace Status:")
		util.KubeCheck(c.NS)
	} else if !cleanup {
		log.Printf("Namespace Delete: currently disabled (enable with \"--cleanup\")")
		log.Printf("Namespace Status:")
		util.KubeCheck(c.NS)
	} else {
		log.Printf("Namespace Delete:")
		util.KubeDelete(c.NS)
	}
}

// RunStatus runs a CLI command
func RunStatus(cmd *cobra.Command, args []string) {
	c := LoadOperatorConf(cmd)
	LoadAdmissionConf(c)
	util.KubeCheck(c.NS)
	if util.KubeCheck(c.SA) && util.KubeCheck(c.SAEndpoint) {
		// in OLM deployment the roles and bindings have generated names
		// so we list and lookup bindings to our service account to discover the actual names
		DetectRole(c)
		DetectClusterRole(c)
	}
	util.KubeCheck(c.Role)
	util.KubeCheck(c.RoleBinding)
	util.KubeCheck(c.RoleEndpoint)
	util.KubeCheck(c.RoleBindingEndpoint)
	util.KubeCheck(c.ClusterRole)
	util.KubeCheck(c.ClusterRoleBinding)
	util.KubeCheckOptional(c.WebhookConfiguration)
	util.KubeCheckOptional(c.WebhookSecret)
	util.KubeCheckOptional(c.WebhookService)
	noDeploy, _ := cmd.Flags().GetBool("no-deploy")
	if !noDeploy {
		util.KubeCheck(c.Deployment)
	}
}

// RunYaml runs a CLI command
func RunYaml(cmd *cobra.Command, args []string) {
	c := LoadOperatorConf(cmd)
	p := printers.YAMLPrinter{}
	util.Panic(p.PrintObj(c.NS, os.Stdout))
	util.Panic(p.PrintObj(c.SA, os.Stdout))
	util.Panic(p.PrintObj(c.Role, os.Stdout))
	util.Panic(p.PrintObj(c.RoleBinding, os.Stdout))
	util.Panic(p.PrintObj(c.SAEndpoint, os.Stdout))
	util.Panic(p.PrintObj(c.RoleEndpoint, os.Stdout))
	util.Panic(p.PrintObj(c.RoleBindingEndpoint, os.Stdout))
	util.Panic(p.PrintObj(c.ClusterRole, os.Stdout))
	util.Panic(p.PrintObj(c.ClusterRoleBinding, os.Stdout))
	noDeploy, _ := cmd.Flags().GetBool("no-deploy")
	if !noDeploy {
		util.Panic(p.PrintObj(c.Deployment, os.Stdout))
	}
}

// Conf struct holds all the objects needed to install the operator
type Conf struct {
	NS                   *corev1.Namespace
	SA                   *corev1.ServiceAccount
	SAEndpoint           *corev1.ServiceAccount
	SAUI                 *corev1.ServiceAccount
	Role                 *rbacv1.Role
	RoleEndpoint         *rbacv1.Role
	RoleUI               *rbacv1.Role
	RoleBinding          *rbacv1.RoleBinding
	RoleBindingEndpoint  *rbacv1.RoleBinding
	ClusterRole          *rbacv1.ClusterRole
	ClusterRoleBinding   *rbacv1.ClusterRoleBinding
	Deployment           *appsv1.Deployment
	WebhookConfiguration *admissionv1.ValidatingWebhookConfiguration
	WebhookSecret        *corev1.Secret
	WebhookService       *corev1.Service
}

// LoadOperatorConf loads and initializes all the objects needed to install the operator
func LoadOperatorConf(cmd *cobra.Command) *Conf {
	c := &Conf{}

	c.NS = util.KubeObject(bundle.File_deploy_namespace_yaml).(*corev1.Namespace)
	c.SA = util.KubeObject(bundle.File_deploy_service_account_yaml).(*corev1.ServiceAccount)
	c.SAEndpoint = util.KubeObject(bundle.File_deploy_service_account_endpoint_yaml).(*corev1.ServiceAccount)
	c.SAUI = util.KubeObject(bundle.File_deploy_service_account_ui_yaml).(*corev1.ServiceAccount)
	c.Role = util.KubeObject(bundle.File_deploy_role_yaml).(*rbacv1.Role)
	c.RoleEndpoint = util.KubeObject(bundle.File_deploy_role_endpoint_yaml).(*rbacv1.Role)
	c.RoleUI = util.KubeObject(bundle.File_deploy_role_ui_yaml).(*rbacv1.Role)
	c.RoleBinding = util.KubeObject(bundle.File_deploy_role_binding_yaml).(*rbacv1.RoleBinding)
	c.RoleBindingEndpoint = util.KubeObject(bundle.File_deploy_role_binding_endpoint_yaml).(*rbacv1.RoleBinding)
	c.ClusterRole = util.KubeObject(bundle.File_deploy_cluster_role_yaml).(*rbacv1.ClusterRole)
	c.ClusterRoleBinding = util.KubeObject(bundle.File_deploy_cluster_role_binding_yaml).(*rbacv1.ClusterRoleBinding)
	c.Deployment = util.KubeObject(bundle.File_deploy_operator_yaml).(*appsv1.Deployment)

	c.NS.Name = options.Namespace
	c.SA.Namespace = options.Namespace
	c.SAEndpoint.Namespace = options.Namespace
	c.Role.Namespace = options.Namespace
	c.RoleEndpoint.Namespace = options.Namespace
	c.RoleBinding.Namespace = options.Namespace
	c.RoleBindingEndpoint.Namespace = options.Namespace
	c.ClusterRole.Namespace = options.Namespace
	c.Deployment.Namespace = options.Namespace

	configureClusterRole(c.ClusterRole)
	c.ClusterRoleBinding.Name = c.ClusterRole.Name
	c.ClusterRoleBinding.RoleRef.Name = c.ClusterRole.Name
	for i := range c.ClusterRoleBinding.Subjects {
		c.ClusterRoleBinding.Subjects[i].Namespace = options.Namespace
	}

	c.Deployment.Spec.Template.Spec.Containers[0].Image = options.OperatorImage
	if options.ImagePullSecret != "" {
		c.Deployment.Spec.Template.Spec.ImagePullSecrets =
			[]corev1.LocalObjectReference{{Name: options.ImagePullSecret}}
	}

	return c
}

// DetectRole looks up a role binding referencing our service account
func DetectRole(c *Conf) {
	roleBindings := &rbacv1.RoleBindingList{}
	selector := labels.SelectorFromSet(labels.Set{
		"olm.owner.kind":      "ClusterServiceVersion",
		"olm.owner.namespace": c.SA.Namespace,
	})
	util.KubeList(roleBindings, &client.ListOptions{
		Namespace:     c.SA.Namespace,
		LabelSelector: selector,
	})
	for i := range roleBindings.Items {
		b := &roleBindings.Items[i]
		for j := range b.Subjects {
			s := b.Subjects[j]
			if s.Kind == "ServiceAccount" &&
				s.Name == c.SA.Name &&
				s.Namespace == c.SA.Namespace {
				c.Role.Name = b.RoleRef.Name
				c.RoleBinding.Name = b.Name
			}
			if s.Kind == "ServiceAccount" &&
				s.Name == c.SAEndpoint.Name &&
				s.Namespace == c.SAEndpoint.Namespace {
				c.RoleEndpoint.Name = b.RoleRef.Name
				c.RoleBindingEndpoint.Name = b.Name
			}
		}
	}
}

// DetectClusterRole looks up a cluster role binding referencing our service account
func DetectClusterRole(c *Conf) {
	clusterRoleBindings := &rbacv1.ClusterRoleBindingList{}
	selector := labels.SelectorFromSet(labels.Set{
		"olm.owner.kind":      "ClusterServiceVersion",
		"olm.owner.namespace": c.SA.Namespace,
	})
	util.KubeList(clusterRoleBindings, &client.ListOptions{
		LabelSelector: selector,
	})
	for i := range clusterRoleBindings.Items {
		b := &clusterRoleBindings.Items[i]
		for j := range b.Subjects {
			s := b.Subjects[j]
			if s.Kind == "ServiceAccount" &&
				s.Name == c.SA.Name &&
				s.Namespace == c.SA.Namespace {
				c.ClusterRole.Name = b.RoleRef.Name
				c.ClusterRoleBinding.Name = b.Name
				return
			}
		}
	}
}

// LoadAdmissionConf loads and initializes all the objects needed to install the admission resources
func LoadAdmissionConf(c *Conf) {
	// Load admission resources yaml files
	c.WebhookConfiguration = util.KubeObject(bundle.File_deploy_internal_admission_webhook_yaml).(*admissionv1.ValidatingWebhookConfiguration)
	c.WebhookSecret = util.KubeObject(bundle.File_deploy_internal_secret_empty_yaml).(*corev1.Secret)
	c.WebhookService = util.KubeObject(bundle.File_deploy_internal_service_admission_webhook_yaml).(*corev1.Service)

	// Set resources Name and Namespace
	c.WebhookConfiguration.Namespace = options.Namespace
	c.WebhookSecret.Namespace = options.Namespace
	c.WebhookService.Namespace = options.Namespace
	c.WebhookConfiguration.Webhooks[0].ClientConfig.Service.Namespace = options.Namespace
	c.WebhookSecret.Name = "admission-webhook-secret"
}

// AdmissionWebhookSetup generate self-signed certificate and add volume mount to the operator deployment
func AdmissionWebhookSetup(c *Conf) {
	var caPEM, serverCertPEM, serverPrivKeyPEM *bytes.Buffer
	// CA config
	ca := &x509.Certificate{
		SerialNumber:          big.NewInt(2020),
		NotBefore:             time.Now(),
		NotAfter:              time.Now().AddDate(1, 0, 0),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}

	// CA private key
	caPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
	}

	// Self signed CA certificate
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	// PEM encode CA cert
	caPEM = new(bytes.Buffer)
	_ = pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	dnsNames := []string{c.WebhookService.Name, c.WebhookService.Name + "." + options.Namespace, c.WebhookService.Name + "." + options.Namespace + ".svc"}
	commonName := c.WebhookService.Name + "." + options.Namespace + ".svc"

	// server cert config
	cert := &x509.Certificate{
		DNSNames:     dnsNames,
		SerialNumber: big.NewInt(1658),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().AddDate(1, 0, 0),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	// server private key
	serverPrivKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		fmt.Println(err)
	}

	// sign the server cert
	serverCertBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, ca, &serverPrivKey.PublicKey, caPrivKey)
	if err != nil {
		fmt.Println(err)
	}

	// PEM encode the  server cert and key
	serverCertPEM = new(bytes.Buffer)
	_ = pem.Encode(serverCertPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: serverCertBytes,
	})

	serverPrivKeyPEM = new(bytes.Buffer)
	_ = pem.Encode(serverPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(serverPrivKey),
	})

	c.WebhookSecret.Data["tls.cert"] = serverCertPEM.Bytes()
	c.WebhookSecret.Data["tls.key"] = serverPrivKeyPEM.Bytes()
	c.WebhookConfiguration.Webhooks[0].ClientConfig.CABundle = caPEM.Bytes()

	volumeMounts := make([]corev1.VolumeMount, 1)
	volumeMounts[0] = corev1.VolumeMount{
		Name:      "webhook-certs",
		MountPath: "/etc/certs",
		ReadOnly:  true,
	}
	c.Deployment.Spec.Template.Spec.Containers[0].VolumeMounts = volumeMounts

	volumes := make([]corev1.Volume, 1)
	volumes[0] = corev1.Volume{
		Name: "webhook-certs",
		VolumeSource: corev1.VolumeSource{
			Secret: &corev1.SecretVolumeSource{
				SecretName: c.WebhookSecret.Name,
			},
		},
	}
	c.Deployment.Spec.Template.Spec.Volumes = volumes
}

func configureClusterRole(cr *rbacv1.ClusterRole) {
	cr.Name = options.SubDomainNS()
}
