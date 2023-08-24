package diagnostics

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	nbv1 "github.com/noobaa/noobaa-operator/v5/pkg/apis/noobaa/v1alpha1"
	"github.com/noobaa/noobaa-operator/v5/pkg/backingstore"
	"github.com/noobaa/noobaa-operator/v5/pkg/bundle"
	"github.com/noobaa/noobaa-operator/v5/pkg/namespacestore"
	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/util/wait"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spf13/cobra"
)

var ctx = context.TODO()

// RunAnalyzeBackingStore runs a CLI command
func RunAnalyzeBackingStore(cmd *cobra.Command, args []string) {
	log := util.Logger()

	backingStore := backingstore.GetBackingStoreFromArgs(cmd, args)
	if checkIfBackingStoreTypeIsSupported(backingStore) {
		log.Printf("⏳ Running this command will run tests on BackingStore %s...\n", backingStore.Name)
		collector := collectorInstance(cmd)
		makeDirForLogs(collector.folderName)
		analyzeBackingstore(cmd, backingStore, collector)
		destDir, _ := cmd.Flags().GetString("dir")
		err := printTestsSummary(collector.folderName)
		if err != nil {
			log.Errorln("❌ Could not print tests summary")
		}
		collector.ExportDiagnostics(destDir)
	} else {
		log.Printf("BackingStore %s type is not supported in analyze\n", backingStore.Name)
	}
}

// RunAnalyzeNamespaceStore runs a CLI command
func RunAnalyzeNamespaceStore(cmd *cobra.Command, args []string) {
	log := util.Logger()

	namespaceStore := namespacestore.GetNamespaceStoreFromArgs(cmd, args)
	if checkIfNamespaceStoreTypeIsSupported(namespaceStore) {
		log.Printf("⏳ Running this command will run tests on NamespaceStore %s...\n", namespaceStore.Name)
		collector := collectorInstance(cmd)
		makeDirForLogs(collector.folderName)
		analyzeNamespaceStore(cmd, namespaceStore, collector)
		destDir, _ := cmd.Flags().GetString("dir")
		err := printTestsSummary(collector.folderName)
		if err != nil {
			log.Errorln("❌ Could not print tests summary")
		}
		collector.ExportDiagnostics(destDir)
	} else {
		log.Printf("NamespaceStore %s type is not supported in analyze\n", namespaceStore.Name)
	}
}

// RunAnalyzeResources runs a CLI command
func RunAnalyzeResources(cmd *cobra.Command, args []string) {
	collector := collectorInstance(cmd)
	makeDirForLogs(collector.folderName)
	collector.log.Println("⏳ Running this command will run tests on all your resources")
	collector.log.Println("Iterating over all BackingStores...")
	foundBackingStore := analyzeAllBackingStores(cmd, collector)
	collector.log.Println("Iterating over all NamespaceStore...")
	foundNamespaceStore := analyzeAllNamespaceStores(cmd, collector)
	if foundBackingStore || foundNamespaceStore {
		destDir, _ := cmd.Flags().GetString("dir")
		err := printTestsSummary(collector.folderName)
		if err != nil {
			collector.log.Errorln("❌ Could not print tests summary")
		}
		collector.ExportDiagnostics(destDir)
	}
}

func analyzeAllBackingStores(cmd *cobra.Command, collector *Collector) bool {
	list := &nbv1.BackingStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "BackingStoreList"},
	}
	if !util.KubeList(list, &client.ListOptions{Namespace: options.Namespace}) {
		collector.log.Println("Could not get BackingStoreList.")
		return false
	}
	if len(list.Items) == 0 {
		collector.log.Println("No BackingStores found.")
		return false
	}

	foundBackingStore := false
	for _, backingStore := range list.Items {
		if checkIfBackingStoreTypeIsSupported(&backingStore) {
			foundBackingStore = true
			collector.log.Printf("BackingStore %s:\n", backingStore.Name)
			analyzeBackingstore(cmd, &backingStore, collector)
		} else {
			collector.log.Printf("BackingStore %s type is not supported in analyze\n", backingStore.Name)
		}
	}
	return foundBackingStore
}

func analyzeAllNamespaceStores(cmd *cobra.Command, collector *Collector) bool {
	list := &nbv1.NamespaceStoreList{
		TypeMeta: metav1.TypeMeta{Kind: "NamespaceStoreList"},
	}
	if !util.KubeList(list, &client.ListOptions{Namespace: options.Namespace}) {
		collector.log.Println("Could not get NamespaceStoreList.")
		return false
	}
	if len(list.Items) == 0 {
		collector.log.Println("No NamespaceStores found.")
		return false
	}

	foundNamespaceStore := false
	for _, namespaceStore := range list.Items {
		if checkIfNamespaceStoreTypeIsSupported(&namespaceStore) {
			foundNamespaceStore = true
			collector.log.Printf("NamespaceStores %s:\n", namespaceStore.Name)
			analyzeNamespaceStore(cmd, &namespaceStore, collector)
		} else {
			collector.log.Printf("NamespaceStore %s type is not supported in analyze\n", namespaceStore.Name)
		}
	}
	return foundNamespaceStore
}

func checkIfBackingStoreTypeIsSupported(backingStore *nbv1.BackingStore) bool {
	if util.IsSTSClusterBS(backingStore) || backingStore.Spec.Type == nbv1.StoreTypePVPool {
		return false
	}
	return true
}

func checkIfNamespaceStoreTypeIsSupported(namespaceStore *nbv1.NamespaceStore) bool {
	if util.IsSTSClusterNS(namespaceStore) || namespaceStore.Spec.Type == nbv1.NSStoreTypeNSFS {
		return false
	}
	return true
}

func analyzeBackingstore(cmd *cobra.Command, backingStore *nbv1.BackingStore, collector *Collector) {
	analyzeResourceJob := loadAnalyzeResourceJob()
	setImageInJob(analyzeResourceJob)
	setNetworkEnvsInJob(analyzeResourceJob)
	setBackingStoreDetailsInJob(backingStore, cmd, analyzeResourceJob)
	setJobAnalyzeResource(cmd, analyzeResourceJob, collector)
}

func analyzeNamespaceStore(cmd *cobra.Command, namespaceStore *nbv1.NamespaceStore, collector *Collector) {
	analyzeResourceJob := loadAnalyzeResourceJob()
	setImageInJob(analyzeResourceJob)
	setNetworkEnvsInJob(analyzeResourceJob)
	setNamespacetoreDetailsInJob(namespaceStore, cmd, analyzeResourceJob)
	setJobAnalyzeResource(cmd, analyzeResourceJob, collector)
}

func loadAnalyzeResourceJob() *batchv1.Job {
	analyzeResourceJob := util.KubeObject(bundle.File_deploy_job_analyze_resource_yml).(*batchv1.Job)
	analyzeResourceJob.Namespace = options.Namespace
	return analyzeResourceJob
}

func collectorInstance(cmd *cobra.Command) *Collector {
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	collector := &Collector{
		folderName:  fmt.Sprintf("%s_%d", "noobaa_analyze_resource", time.Now().Unix()),
		log:         util.Logger(),
		kubeconfig:  kubeconfig,
		kubeCommand: util.GetAvailabeKubeCli(),
	}
	return collector
}

func makeDirForLogs(folderName string) {
	err := os.Mkdir(folderName, os.ModePerm)
	if err != nil {
		util.Logger().Fatalf(`❌ Could not create directory %s, reason: %s`, folderName, err)
	}
}

func setImageInJob(analyzeResourceJob *batchv1.Job) {
	noobaa := util.KubeObject(bundle.File_deploy_crds_noobaa_io_v1alpha1_noobaa_cr_yaml).(*nbv1.NooBaa)
	noobaa.Namespace = options.Namespace
	if !util.KubeCheck(noobaa) {
		util.Logger().Fatalf(`❌ Could not get noobaa %q in Namespace %q`,
			noobaa.Name, noobaa.Namespace)
	}
	analyzeResourceJob.Spec.Template.Spec.Containers[0].Image = noobaa.Status.ActualImage
}

func setNetworkEnvsInJob(analyzeResourceJob *batchv1.Job) {
	log := util.Logger()

	coreApp := util.KubeObject(bundle.File_deploy_internal_statefulset_core_yaml).(*appsv1.StatefulSet)
	coreApp.Namespace = options.Namespace
	if !util.KubeCheck(coreApp) {
		util.Logger().Fatalf(`❌ Could not get core StatefulSet %q in Namespace %q`,
			coreApp.Name, coreApp.Namespace)
	}

	for _, networkEnvName := range []string{"HTTP_PROXY", "HTTPS_PROXY", "NO_PROXY", "NODE_EXTRA_CA_CERTS"} {
		networkEnvInCoreApp := util.GetEnvVariable(&coreApp.Spec.Template.Spec.Containers[0].Env, networkEnvName)
		if networkEnvInCoreApp != nil && networkEnvInCoreApp.Value != "" {
			networkEnvInJob := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, networkEnvName)
			if networkEnvInJob == nil {
				log.Fatalf("❌ Could not get %s environment variable in %s %q in Namespace %q",
					networkEnvName, analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
			}
			networkEnvInJob.Value = networkEnvInCoreApp.Value
		}
	}
}

func setBackingStoreDetailsInJob(backingStore *nbv1.BackingStore, cmd *cobra.Command, analyzeResourceJob *batchv1.Job) {
	log := util.Logger()

	// type of backingStore
	storeType := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "RESOURCE_TYPE")
	if storeType == nil {
		log.Fatalf("❌ Could not get resource type in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	storeType.Value = string(backingStore.Spec.Type)

	// name of backingStore
	storeName := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "RESOURCE_NAME")
	if storeName == nil {
		log.Fatalf("❌ Could not get resource name in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	storeName.Value = backingStore.Name

	// bucket to analyze (target bucket by default)
	bucket := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "BUCKET")
	if bucket == nil {
		log.Fatalf("❌ Could not get bucket in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	var err error
	bucket.Value, err = util.GetBackingStoreTargetBucket(backingStore)
	if err != nil {
		log.Fatalf("❌ Could not get target bucket of %s %q in Namespace %q",
			backingStore.Kind, backingStore.Name, backingStore.Namespace)
	}

	// endpoint url
	endpoint := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "ENDPOINT")
	if endpoint == nil {
		log.Fatalf("❌ Could not get endpoint in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	endpoint.Value, err = util.GetEndpointByBackingStoreType(backingStore)
	if err != nil {
		log.Fatalf("❌ Could not get endpoint in %s %q in Namespace %q",
			backingStore.Kind, backingStore.Name, backingStore.Namespace)
	}

	// in case it is aws endpoint, we would pass the region
	if backingStore.Spec.Type == nbv1.StoreTypeAWSS3 {
		region := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "REGION")
		if region == nil {
			log.Fatalf("❌ Could not get region in %s %q in Namespace %q",
				analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
		}
		region.Value = backingStore.Spec.AWSS3.Region
	}

	// signature version
	signatureVersion := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "S3_SIGNATURE_VERSION")
	if signatureVersion == nil {
		log.Fatalf("❌ Could not get signatureVersion in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	switch backingStore.Spec.Type {
	case nbv1.StoreTypeS3Compatible:
		signatureVersion.Value = string(backingStore.Spec.S3Compatible.SignatureVersion)
	case nbv1.StoreTypeIBMCos:
		signatureVersion.Value = string(backingStore.Spec.IBMCos.SignatureVersion)
	default:
		signatureVersion.Value = "NOT_DEFINED"
	}

	// SecretName in the job yaml
	secretRef, err := util.GetBackingStoreSecret(backingStore)
	if err != nil {
		log.Fatalf("❌ Could not get backing store secret in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	analyzeResourceJob.Spec.Template.Spec.Volumes[0].Secret.SecretName = secretRef.Name
}

func setNamespacetoreDetailsInJob(namespaceStore *nbv1.NamespaceStore, cmd *cobra.Command, analyzeResourceJob *batchv1.Job) {
	log := util.Logger()

	// type of namespaceStore
	storeType := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "RESOURCE_TYPE")
	if storeType == nil {
		log.Fatalf("❌ Could not get resource type in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	storeType.Value = string(namespaceStore.Spec.Type)

	// name of namespaceStore
	storeName := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "RESOURCE_NAME")
	if storeName == nil {
		log.Fatalf("❌ Could not get resource name in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	storeName.Value = namespaceStore.Name

	// bucket to analyze (target bucket by default)
	bucket := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "BUCKET")
	if bucket == nil {
		log.Fatalf("❌ Could not get bucket in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	var err error
	bucket.Value, err = util.GetNamespaceStoreTargetBucket(namespaceStore)
	if err != nil {
		log.Fatalf("❌ Could not get target bucket of %s %q in Namespace %q",
			namespaceStore.Kind, namespaceStore.Name, namespaceStore.Namespace)
	}

	// endpoint url
	endpoint := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "ENDPOINT")
	if endpoint == nil {
		log.Fatalf("❌ Could not get endpoint  in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	endpoint.Value, err = util.GetEndpointByNamespaceStoreType(namespaceStore)
	if err != nil {
		log.Fatalf("❌ Could not get endpoint  in %s %q in Namespace %q",
			namespaceStore.Kind, namespaceStore.Name, namespaceStore.Namespace)
	}

	// in case it is aws endpoint, we would pass the region
	if namespaceStore.Spec.Type == nbv1.NSStoreTypeAWSS3 {
		region := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "REGION")
		if region == nil {
			log.Fatalf("❌ Could not get region in %s %q in Namespace %q",
				analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
		}
		region.Value = namespaceStore.Spec.AWSS3.Region
	}

	// signature version
	signatureVersion := util.GetEnvVariable(&analyzeResourceJob.Spec.Template.Spec.Containers[0].Env, "S3_SIGNATURE_VERSION")
	if signatureVersion == nil {
		log.Fatalf("❌ Could not get signatureVersion  in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	switch namespaceStore.Spec.Type {
	case nbv1.NSStoreTypeS3Compatible:
		signatureVersion.Value = string(namespaceStore.Spec.S3Compatible.SignatureVersion)
	case nbv1.NSStoreTypeIBMCos:
		signatureVersion.Value = string(namespaceStore.Spec.IBMCos.SignatureVersion)
	default:
		signatureVersion.Value = "NOT_DEFINED"
	}

	// SecretName in the job yaml
	secretRef, err := util.GetNamespaceStoreSecret(namespaceStore)
	if err != nil {
		log.Fatalf("❌ Could not get namespace store secret  in %s %q in Namespace %q",
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
	analyzeResourceJob.Spec.Template.Spec.Volumes[0].Secret.SecretName = secretRef.Name
}

func setJobAnalyzeResource(cmd *cobra.Command, analyzeResourceJob *batchv1.Job, collector *Collector) {
	// change job resources (cpu, memory)
	JobResourcesJSON, _ := cmd.Flags().GetString("job-resources")
	if JobResourcesJSON != "" {
		util.Panic(json.Unmarshal([]byte(JobResourcesJSON), &analyzeResourceJob.Spec.Template.Spec.Containers[0].Resources))
	}

	// create the job!
	if !util.KubeCreateFailExisting(analyzeResourceJob) {
		collector.log.Fatalf(`❌ Could not create %s %q in Namespace %q`,
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}

	// wait for the job
	waitFinish(analyzeResourceJob)

	// save the logs
	podSelector, _ := labels.Parse(fmt.Sprintf(`job-name=%s`, analyzeResourceJob.Name))
	listOptions := client.ListOptions{Namespace: options.Namespace, LabelSelector: podSelector}
	collector.CollectPodsLogs(listOptions)

	// delete the job and its pod
	propagationPolicy := metav1.DeletePropagationForeground
	deleteOpts := client.DeleteOptions{PropagationPolicy: &propagationPolicy}
	if !util.KubeDelete(analyzeResourceJob, &deleteOpts) {
		collector.log.Fatalf(`❌ Could not delete %s %q in Namespace %q`,
			analyzeResourceJob.Kind, analyzeResourceJob.Name, analyzeResourceJob.Namespace)
	}
}

func waitFinish(job *batchv1.Job) bool {
	klient := util.KubeClient()
	interval := time.Duration(3)
	
	err := wait.PollUntilContextCancel(ctx, interval*time.Second, true, func(ctx context.Context) (bool, error) {
		err := klient.Get(util.Context(), util.ObjectKey(job), job)
		if err != nil {
			util.Logger().Printf("⏳ Failed to get Job: %s", err)
			return false, nil
		}
		checkJobStatus(job)
		if job.Status.Failed > 0 {
			return false, fmt.Errorf("JobFailed")
		}
		if job.Status.Succeeded == 0 {
			return false, nil
		}
		return true, nil
	})
	return (err == nil)
}

func checkJobStatus(job *batchv1.Job) {
	log := util.Logger()
	if job.Status.Succeeded > 0 {
		log.Printf("✅ Job %q is Completed", job.Name)
	} else if job.Status.Failed > 0 {
		log.Errorf("❌ Job %q is Failed", job.Name)
	} else {
		log.Printf("⏳ Job %q is Unfinished, waiting...", job.Name)
	}
}

func printTestsSummary(folderName string) error {
	log := util.Logger()

	dirEntries, err := os.ReadDir(folderName)
	if err != nil {
		log.Errorln("❌ Could not get files, error: ", err)
		return err
	}

	log.Println("Summary:")
	for _, dirEntry := range dirEntries {
		ext := filepath.Ext(dirEntry.Name())
		if ext == ".log" {
			path := fmt.Sprintf("%s/%s", folderName, dirEntry.Name())

			_, err := os.Stat(path)
			if os.IsNotExist(err) {
				log.Errorf("❌ File does not exist in path: %v, error: %v", path, err)
				return err
			}

			file, err := os.Open(path)
			if err != nil {
				log.Errorf("❌ Could not open file: %v", err)
				return err
			}
			defer file.Close()

			scanner := bufio.NewScanner(file)
			shouldPrint := false
			// this is a printing that appears if it fails configuring or after the tests run (in the core repo)
			prefixToSearchFailed := "Test Diagnose Resource Failed"
			prefixToSearchSummary := "Analyze Resource Tests Result"
			var indexToExtractLine int
			for scanner.Scan() {
				line := scanner.Text()
				// We find the prefix in the file
				if strings.Contains(line, prefixToSearchFailed) {
					shouldPrint = true
					indexToExtractLine = strings.Index(line, prefixToSearchFailed)
				} else if strings.Contains(line, prefixToSearchSummary) {
					shouldPrint = true
					indexToExtractLine = strings.Index(line, prefixToSearchSummary)
				}
				// Once we found the prefix we will print all the log lines without the dbg part
				if shouldPrint {
					substring := line[indexToExtractLine:]
					log.Println(substring)
				}
			}

			if err := scanner.Err(); err != nil {
				return err
			}

		}
	}
	return nil
}
