package diagnose

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/dbdump"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	secv1 "github.com/openshift/api/security/v1"
	"github.com/sirupsen/logrus"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

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
		Use:   "diagnose",
		Short: "Collect diagnostics",
		Run:   RunCollect,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().String("dir", "", "collect noobaa diagnose tar file into destination directory")
	cmd.Flags().Bool("db-dump", false, "collect db dump in addition to diagnostics")
	return cmd
}

// RunCollect runs a CLI command
func RunCollect(cmd *cobra.Command, args []string) {

	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	destDir, _ := cmd.Flags().GetString("dir")
	collectDBDump, _ := cmd.Flags().GetBool("db-dump")
	c := Collector{
		folderName: fmt.Sprintf("%s_%d", "noobaa_diagnostics", time.Now().Unix()),
		log:        util.Logger(),
		kubeconfig: kubeconfig,
	}

	c.log.Println("Running collection of diagnostics")

	err := os.Mkdir(c.folderName, os.ModePerm)
	if err != nil {
		c.log.Fatalf(`❌ Could not create directory %s, reason: %s`, c.folderName, err)
	}

	c.kubeCommand = util.GetAvailabeKubeCli();


	// Define to select only noobaa pods within the namespace
	podSelector, _ := labels.Parse("app=noobaa")
	listOptions := client.ListOptions{Namespace: options.Namespace, LabelSelector: podSelector}

	c.CollectCRs()
	c.CollectPodsLogs(listOptions)
	c.CollectPVs(listOptions)
	c.CollectPVCs(listOptions)
	c.CollectSCC()

	c.ExportDiagnostics(destDir)

	// Collects db dump in addition to diagnostics.
	// A separate tarball is created for diagnostics and db dump
	if collectDBDump {
		dbdump.CollectDBDump(kubeconfig, destDir)
	}
}

// CollectCR info
func (c *Collector) CollectCR(list client.ObjectList) {
	gvk := list.GetObjectKind().GroupVersionKind()

	if !util.KubeList(list, &client.ListOptions{Namespace: options.Namespace}) {
		c.log.Printf(`❌ Failed to list %s\n`, gvk.Kind)
		return
	}

	list.GetObjectKind().SetGroupVersionKind(gvk)

	targetFile := fmt.Sprintf("%s/%s_crs.yaml", c.folderName, gvk.Kind)
	err := util.SaveCRsToFile(list, targetFile)
	if err != nil {
		c.log.Printf("got error on util.SaveCRsToFile for %v: %v", targetFile, err)
	}
}

// CollectCRs collects the content of multiple CR types
func (c *Collector) CollectCRs() {
	c.CollectCR(&nbv1.BackingStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStoreList"},
	})

	c.CollectCR(&nbv1.NamespaceStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "NamespaceStoreList"},
	})

	c.CollectCR(&nbv1.BucketClassList{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClassList"},
	})

	c.CollectCR(&nbv1.NooBaaList{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaaList"},
	})

	c.CollectCR(&nbv1.NooBaaAccountList{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaaAccountList"},
	})
}

// CollectDescribe collects output of the "describe pod" of a single pod
func (c *Collector) CollectDescribe(Kind string, Name string) {
	cmd := exec.Command(c.kubeCommand, "describe", Kind, "-n", options.Namespace, Name)
	// handle custom path for kubeconfig file,
	// see --kubeconfig cli options
	if len(c.kubeconfig) > 0 {
		cmd.Env = append(cmd.Env, "KUBECONFIG="+c.kubeconfig)
	}

	// open the out file for writing
	fileName := c.folderName + "/" + Name + "-" + Kind + "-describe.txt"
	outfile, err := os.Create(fileName)
	if err != nil {
		c.log.Printf(`❌ cannot create file %v: %v`, fileName, err)
		return
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	// run kubectl describe
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ cannot describe %v %v in namespace %v: %v`, Kind, Name, options.Namespace, err)
	}
}

// CollectPodsLogs collects logs of all existing noobaa pods
func (c *Collector) CollectPodsLogs(listOptions client.ListOptions) {
	// List all pods and select only noobaa pods within the relevant namespace
	c.log.Println("Collecting pod logs")
	podList := &corev1.PodList{}
	if !util.KubeList(podList, &listOptions) {
		c.log.Printf(`❌ failed to get noobaa pod list within namespace %s\n`, options.Namespace)
		return
	}

	// Iterate the list of pods, collecting the logs of each
	for i := range podList.Items {
		pod := &podList.Items[i]

		c.CollectDescribe("pod", pod.Name)

		podLogs, _ := util.GetPodLogs(*pod)
		for containerName, containerLog := range podLogs {
			targetFile := fmt.Sprintf("%s/%s-%s.log", c.folderName, pod.Name, containerName)
			err := util.SaveStreamToFile(containerLog, targetFile)
			if err != nil {
				c.log.Printf("got error on util.SaveStreamToFile for %v: %v", targetFile, err)
			}
		}
	}
}

// CollectPVs collects describe of PVs
func (c *Collector) CollectPVs(listOptions client.ListOptions) {
	// List all PVs and select only noobaa PVs within the relevant namespace
	c.log.Println("Collecting PV logs")
	pvList := &corev1.PersistentVolumeList{}
	if !util.KubeList(pvList, &listOptions) {
		c.log.Printf(`❌ failed to get noobaa PV list within namespace %s\n`, options.Namespace)
		return
	}

	// Iterate the list of PVs, collecting the describe of each
	for i := range pvList.Items {
		pv := &pvList.Items[i]
		c.CollectDescribe("pv", pv.Name)
	}
}

// CollectPVCs collects describe of PVCs
func (c *Collector) CollectPVCs(listOptions client.ListOptions) {
	// List all PVCs and select only noobaa PVCs within the relevant namespace
	c.log.Println("Collecting PVC logs")
	pvcList := &corev1.PersistentVolumeClaimList{}
	if !util.KubeList(pvcList, &listOptions) {
		c.log.Printf(`❌ failed to get noobaa PVC list within namespace %s\n`, options.Namespace)
		return
	}

	// Iterate the list of PVCs, collecting the describe of each
	for i := range pvcList.Items {
		pvc := &pvcList.Items[i]
		c.CollectDescribe("pvc", pvc.Name)
	}
}

// CollectSCC collects the SCC
func (c *Collector) CollectSCC() {
	c.log.Println("Collecting SCC logs")
	for _, name := range []string{"noobaa", "noobaa-endpoint"} {
		scc := &secv1.SecurityContextConstraints{
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: options.Namespace,
			},
		}
		if util.KubeCheckOptional(scc) {
			c.CollectDescribe("scc", scc.Name)
		}
	}
}

// TODO: Use port forwarding (usePortForwarding in system.go)
// func collectSystemMetrics() {
// 	sys := getSystemObject()
// 	mgmtAddress := sys.Status.Services.ServiceMgmt.ExternalDNS[0]
// 	mgmtURL, err := url.Parse(mgmtAddress)
// 	if err != nil {
// 		log.Fatalf("failed to parse mgmt address %q. got error: %v", mgmtAddress, err)
// 	}

// 	targetAddress := fmt.Sprintf("%s/metrics/counter", mgmtURL.String())
// 	log.Printf("JENIA THIS IS THE URL %s", targetAddress)
// 	client := &http.Client{Transport: util.InsecureHTTPTransport}
// 	resp, err := client.Get(targetAddress)
// 	if err != nil {
// 		log.Printf(`%s`, err)
// 		log.Fatalf(`❌ JENIA ERROR REQUEST`)
// 		// handle error
// 	}
// 	targetFile := fmt.Sprintf("%s/NooBaa_metrics.txt", folderName)
// 	util.SaveStreamToFile(resp.Body, targetFile)
// }

// ExportDiagnostics info
func (c *Collector) ExportDiagnostics(destDir string) {
	targetFile := fmt.Sprintf("%s.tar.gz", c.folderName)
	if destDir != "" {
		if _, err := os.Stat(destDir); os.IsNotExist(err) {
			err := os.MkdirAll(destDir, os.ModePerm)
			if err != nil {
				c.log.Fatalf(`❌ Could not create directory %s, reason: %s`, destDir, err)
			}
		}
		targetFile = fmt.Sprintf("%s/%s", destDir, targetFile)
	}
	fileToWrite, err := os.Create(targetFile)
	if err != nil {
		c.log.Fatalf(`❌ Could not create target file %s, reason: %s`, targetFile, err)
	}

	err = util.Tar(c.folderName, fileToWrite)
	if err != nil {
		c.log.Fatalf(`❌ Could not compress and package diagnostics, reason: %s`, err)
	}

	err = os.RemoveAll(c.folderName)
	if err != nil {
		c.log.Fatalf(`❌ Could not delete diagnostics collecting folder %s, reason: %s`, c.folderName, err)
	}
}
