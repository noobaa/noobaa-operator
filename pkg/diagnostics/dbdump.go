package diagnostics

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"time"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/sirupsen/logrus"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/spf13/cobra"
)

const (

	// cnpgClusterLabelKey is set (to the cluster name) on every pod managed by a CNPG Cluster.
	cnpgClusterLabelKey = "cnpg.io/cluster=noobaa-db-pg-cluster"
	// cnpgInstanceRoleLabelKey and cnpgLegacyRoleLabelKey mark the primary/replica role
	// of a CNPG managed pod.
	cnpgInstanceRoleLabelKey = "cnpg.io/instanceRole"
	cnpgPrimaryRoleValue     = "primary"
	cnpgReplicaRoleValue     = "replica"
)

// CollectorDbDump configuration for diagnostics
type CollectorDbDump struct {
	folderName       string // Local folder to store the raw db dump
	dbDumpFolderPath string // Local path of the folder containing the raw dump file
	remoteTarPath    string // Remote path (at pod) of tar'd db dump
	kubeconfig       string
	kubeCommand      string
	dbDumpFrom       string // DB Preference. Auto-selecct 'replica' when value is empty
	log              *logrus.Entry
}

func newCollectorDbDump(kubeconfig string, dbDumpFrom string) *CollectorDbDump {
	c := new(CollectorDbDump)
	c.folderName = fmt.Sprintf("%s_%d", "noobaa_db_dump", time.Now().Unix())
	c.remoteTarPath = fmt.Sprintf("%s/%s%s", "/tmp", c.folderName, ".tar.gz")
	c.log = util.Logger()
	c.kubeconfig = kubeconfig
	c.kubeCommand = util.GetAvailabeKubeCli()
	c.dbDumpFrom = dbDumpFrom
	return c
}

// RunDump generates a tar file of a db dump
func RunDump(cmd *cobra.Command, args []string) {
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	destDir, _ := cmd.Flags().GetString("dir")
	dbDumpFrom, _ := cmd.Flags().GetString("dump-from")

	CollectDBDump(kubeconfig, destDir, dbDumpFrom)
}

// CollectDBDump exposes the functionality to the diagnostics collect mechanism
func CollectDBDump(kubeconfig string, destDir string, dbDumpFrom string) {
	c := newCollectorDbDump(kubeconfig, dbDumpFrom)

	c.log.Println("Collecting db dump")

	err := c.generateDBDump(destDir)
	if err == nil {
		err = c.exportDBDump(destDir)
		if err == nil {
			tarPath := fmt.Sprintf("%s.tar.gz", c.dbDumpFolderPath)
			c.log.Printf("✅ Generated db dump was saved in %s\n", tarPath)
			c.deleteRawResources()
		}
	}
}

// isPodReady returns true when the pod has a Ready condition with status True,
// meaning all containers have passed their readiness probes and the pod is
// capable of serving requests.
func isPodReady(pod *corev1.Pod) bool {
	for _, cond := range pod.Status.Conditions {
		if cond.Type == corev1.PodReady {
			return cond.Status == corev1.ConditionTrue
		}
	}
	return false
}

// getDBPodName resolves the name of the postgres pod to run pg_dumpall against.
// Ppod is auto-detected by looking for either:
//  1. The repplica pod of a CNPG managed cluster (e.g. "noobaa-db-pg-cluster-2"), which does not
//     follow a fixed naming convention and is instead identified via its CNPG labels
//     a. if the db from is empty or 'replica', fetch the dump from replica db
//     b. if the db from is primary, fetch the dump from primary db
func (c *CollectorDbDump) getDBPodName() (string, error) {

	primaryDBPodName := ""
	replicaDBPodName := ""
	podList := &corev1.PodList{}
	// Fall back to looking for the primary pod of a CNPG managed cluster
	cnpgSelector, err := labels.Parse(cnpgClusterLabelKey)
	if err != nil {
		return "", fmt.Errorf("failed to parse label selector %q: %v", cnpgClusterLabelKey, err)
	}
	podList = &corev1.PodList{}
	foundDBPod := util.KubeList(podList, &client.ListOptions{Namespace: options.Namespace, LabelSelector: cnpgSelector})
	if !foundDBPod || len(podList.Items) == 0 {
		return "", fmt.Errorf("Could not find noobaa DB pod in namespace %q, "+
			"Make sure your DB pods are up and running ", options.Namespace)
	}
	if foundDBPod {
		for _, pod := range podList.Items {
			role := pod.Labels[cnpgInstanceRoleLabelKey]
			if role == cnpgPrimaryRoleValue && isPodReady(&pod) {
				primaryDBPodName = pod.Name
			} else if role == cnpgReplicaRoleValue && isPodReady(&pod) {
				replicaDBPodName = pod.Name
			}
		}
	}

	switch c.dbDumpFrom {
	case "primary":
		return primaryDBPodName, nil
	case "replica":
		return replicaDBPodName, nil
	default:
		return "", fmt.Errorf("invalid --dump-from value %q; expected primary or replica", c.dbDumpFrom)
	}
}

// Create postgres dump
func (c *CollectorDbDump) generatePostgresDump(destDir string) error {
	dbPodName, err := c.getDBPodName()
	if err != nil {
		c.log.Printf(`❌ cannot find db pod: %v`, err)
		return err
	}
	if dbPodName == "" {
		c.log.Printf(`❌ Could not find db pod for %q db, make sure db pods are in Ready state`, c.dbDumpFrom)
		err = fmt.Errorf("Could not find db pod for %q db", c.dbDumpFrom)
		return err
	}
	c.log.Printf("Using db pod %q for db dump", dbPodName)

	// pg_dumpall's output is piped through gzip inside the pod to reduce the amount of data
	// transferred over the exec stream (important for large DBs). The gzip stream is
	// decompressed locally below so the resulting .sql file on disk is plain SQL text, as its
	// extension implies, rather than leaving it gzip-compressed.
	dumpScript := "set -o pipefail; pg_dumpall --no-role-passwords | gzip -c"
	cmd := exec.Command(c.kubeCommand, "exec", "-n", options.Namespace, "-i", fmt.Sprintf("pod/%s", dbPodName), "--", "bash", "-c", dumpScript)
	// handle custom path for kubeconfig file,
	// see --kubeconfig cli options
	if len(c.kubeconfig) > 0 {
		cmd.Env = append(cmd.Env, "KUBECONFIG="+c.kubeconfig)
	}

	// Compose the path of the folder containing the dump file
	if destDir != "" {
		c.dbDumpFolderPath = fmt.Sprintf("%s/%s", destDir, c.folderName)
	} else {
		c.dbDumpFolderPath = c.folderName
	}

	// Create the folder containing the dump file
	err = os.Mkdir(c.dbDumpFolderPath, os.ModePerm)
	if err != nil {
		c.log.Fatalf(`❌ Could not create directory %s, reason: %s`, c.folderName, err)
		return err
	}

	// Compose the path of the dump file
	dumpFilePath := fmt.Sprintf("%s/%s.sql", c.dbDumpFolderPath, c.folderName)

	// Create the dump file
	outfile, err := os.Create(dumpFilePath)
	if err != nil {
		c.log.Printf(`❌ can not create db dump at path %v: %v`, c.dbDumpFolderPath, err)
		return err
	}
	defer util.SafeClose(outfile, fmt.Sprintf("Failed to close dump file %s", dumpFilePath))
	// Attach to the command's stdout (the gzip-compressed dump) instead of redirecting it
	// straight to the local file, so it can be decompressed on the fly below
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		c.log.Printf(`❌ cannot attach to db dump command output: %v`, err)
		return err
	}
	var stderrBuf bytes.Buffer
	cmd.Stderr = &stderrBuf
	if err := cmd.Start(); err != nil {
		c.log.Printf(`❌ cannot start db dump command: %v`, err)
		return err
	}
	copyErr := decompressGzipStream(stdout, outfile)
	// Always wait for the remote command to finish (reaping the process), regardless of
	// whether reading/decompressing its output above succeeded
	waitErr := cmd.Wait()
	if copyErr != nil {
		c.log.Printf(`❌ cannot generate db dump: %v`, copyErr)
		c.log.Printf(`❌ DB dump command failed with error: %v`, stderrBuf.String())
		return copyErr
	}
	if waitErr != nil {
		c.log.Printf(`❌ cannot generate db dump: %v`, waitErr)
		c.log.Printf(`❌ DB dump command failed with error: %v`, stderrBuf.String())
		return waitErr
	}

	return nil
}

// decompressGzipStream reads a gzip-compressed stream from src and writes the
// decompressed content to dst
func decompressGzipStream(src io.Reader, dst io.Writer) error {
	gzReader, err := gzip.NewReader(src)
	if err != nil {
		return fmt.Errorf("failed reading gzip stream from db pod: %v", err)
	}
	defer util.SafeClose(gzReader, "Failed to close gzip reader")

	if _, err := io.Copy(dst, gzReader); err != nil {
		return fmt.Errorf("failed decompressing db dump: %v", err)
	}
	return nil
}

// Create the dump resources
func (c *CollectorDbDump) generateDBDump(destDir string) error {
	var err error

	c.log.Println("Generating db dump at pod")

	err = c.generatePostgresDump(destDir)

	return err
}

// Tar local dump resources
func (c *CollectorDbDump) exportPostgresDump(destDir string) error {
	// In case of a postgres, the dump was created as a local file, so all is left to
	// do is tar it
	var cmd *exec.Cmd

	// Compose the path of the tar output
	tarPath := fmt.Sprintf("%s.tar.gz", c.dbDumpFolderPath)

	// Genrate the tar command
	if destDir != "" {
		cmd = exec.Command("tar", "-C", destDir, "-cvzf", tarPath, c.folderName)
	} else {
		cmd = exec.Command("tar", "-cvzf", tarPath, c.folderName)
	}

	// Execute the tar command
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ failed to tar db dump`)
		return err
	}

	return nil
}

// Tar local resources (Postgres)
func (c *CollectorDbDump) exportDBDump(destDir string) error {
	var err error

	c.log.Println("Exporting db dump")

	err = c.exportPostgresDump(destDir)

	return err
}

// Delete dump on Postgres remote machine
func (c *CollectorDbDump) deletePostgresRawResources() {
	// In case of postgres, the db dump was created locall, so all is left
	// to do is remove the raw dump

	// Compose the deletion command
	cmd := exec.Command("rm", "-rf", c.dbDumpFolderPath)

	// Execute the deletion of raw resources
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ failed to tar remove dump folder`)
	}
}

// Remove raw resources from local or remote machines
func (c *CollectorDbDump) deleteRawResources() {
	c.deletePostgresRawResources()
}
