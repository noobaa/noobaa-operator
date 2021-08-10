package diagnose

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
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
	folderName string
	kubeconfig string
	log        *logrus.Entry
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
	return cmd
}

// RunCollect runs a CLI command
func RunCollect(cmd *cobra.Command, args []string) {

	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	destDir, _ := cmd.Flags().GetString("dir")
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

	c.CollectCR(&nbv1.BackingStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStoreList"},
	})

	c.CollectCR(&nbv1.BucketClassList{
		TypeMeta: metav1.TypeMeta{Kind: "BucketClassList"},
	})

	c.CollectCR(&nbv1.NooBaaList{
		TypeMeta: metav1.TypeMeta{Kind: "NooBaaList"},
	})

	corePodSelector, _ := labels.Parse("noobaa-core=" + options.SystemName)
	c.CollectPodLogs(corePodSelector)

	operatorPodSelector, _ := labels.Parse("noobaa-operator=deployment")
	c.CollectPodLogs(operatorPodSelector)

	endpointPodSelector, _ := labels.Parse("noobaa-s3=" + options.SystemName)
	c.CollectPodLogs(endpointPodSelector)

	dbPodSelector, _ := labels.Parse("noobaa-db=" + options.SystemName)
	if options.DBType == "postgres" {
		dbPodSelector, _ = labels.Parse("noobaa-db=" + options.DBType)
	}
	c.CollectPodLogs(dbPodSelector)

	// collectSystemMetrics()

	c.CollectSCC()

	c.ExportDiagnostics(destDir)
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

// collect output of the "describe pod"
func (c *Collector) collectPodDescribe(pod *corev1.Pod) {
	cmd := exec.Command("kubectl", "describe", "pod", "-n", pod.Namespace, pod.Name)
	// handle custom path for kubeconfig file,
	// see --kubeconfig cli options
	if len(c.kubeconfig) > 0 {
		cmd.Env = append(cmd.Env, "KUBECONFIG=" + c.kubeconfig)
	}

	// open the out file for writing
	fileName := c.folderName + "/" + pod.Name + "-describe.txt"
	outfile, err := os.Create(fileName)
	if err != nil {
		c.log.Printf(`❌ can not create file %v: %v`, fileName, err)
		return
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	// run kubectl describe
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ can not describe pod %v namespace %v: %v`, pod.Name, pod.Namespace, err)
	}
}

// CollectPodLogs info
func (c *Collector) CollectPodLogs(corePodSelector labels.Selector) {
	corePodList := &corev1.PodList{}
	currentPod := strings.Split(corePodSelector.String(), "=")[0]
	if !util.KubeList(corePodList, &client.ListOptions{Namespace: options.Namespace, LabelSelector: corePodSelector}) {
		return
	}
	if len(corePodList.Items) == 0 {
		c.log.Printf(`❌ No %s pods found\n`, currentPod)
		return
	}

	for i := range corePodList.Items {
		corePod := &corePodList.Items[i]

		c.collectPodDescribe(corePod)

		podLogs, _ := util.GetPodLogs(*corePod)
		for containerName, containerLog := range podLogs {
			targetFile := fmt.Sprintf("%s/%s-%s.log", c.folderName, corePod.Name, containerName)
			err := util.SaveStreamToFile(containerLog, targetFile)
			if err != nil {
				c.log.Printf("got error on util.SaveStreamToFile for %v: %v", targetFile, err)
			}

		}
	}
}

// collectSCCDescribe collect output of the "describe scc"
func (c *Collector) collectSCCDescribe(scc *secv1.SecurityContextConstraints) {
	cmd := exec.Command("kubectl", "describe", "scc", "-n", scc.Namespace, scc.Name)
	// handle custom path for kubeconfig file,
	// see --kubeconfig cli options
	if len(c.kubeconfig) > 0 {
		cmd.Env = append(cmd.Env, "KUBECONFIG=" + c.kubeconfig)
	}

	// open the out file for writing
	fileName := c.folderName + "/" + scc.Name + "-scc-describe.txt"
	outfile, err := os.Create(fileName)
	if err != nil {
		c.log.Printf(`❌ can not create file %v: %v`, fileName, err)
		return
	}
	defer outfile.Close()
	cmd.Stdout = outfile

	// run kubectl describe
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ can not describe scc %v namespace %v: %v`, scc.Name, scc.Namespace, err)
	}
}

// CollectSCC collects the SCC 
func (c *Collector) CollectSCC() {
	for _, name := range []string{"noobaa", "noobaa-endpoint"} {
		scc := &secv1.SecurityContextConstraints {
			ObjectMeta: metav1.ObjectMeta{
				Name:      name,
				Namespace: options.Namespace,
			},
		}
		if util.KubeCheckOptional(scc) {
			c.collectSCCDescribe(scc)
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
