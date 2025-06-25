package diagnostics

import (
	"fmt"
	"os"
	"os/exec"
	"time"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
)

// CollectorDbDump configuration for diagnostics
type CollectorDbDump struct {
	folderName       string // Local folder to store the raw db dump
	dbDumpFolderPath string // Local path of the folder containing the raw dump file
	remoteTarPath    string // Remote path (at pod) of tar'd db dump
	kubeconfig       string
	kubeCommand      string
	log              *logrus.Entry
}

func newCollectorDbDump(kubeconfig string) *CollectorDbDump {
	c := new(CollectorDbDump)
	c.folderName = fmt.Sprintf("%s_%d", "noobaa_db_dump", time.Now().Unix())
	c.remoteTarPath = fmt.Sprintf("%s/%s%s", "/tmp", c.folderName, ".tar.gz")
	c.log = util.Logger()
	c.kubeconfig = kubeconfig
	c.kubeCommand = util.GetAvailabeKubeCli()
	return c
}

// RunDump generates a tar file of a db dump
func RunDump(cmd *cobra.Command, args []string) {
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	destDir, _ := cmd.Flags().GetString("dir")

	CollectDBDump(kubeconfig, destDir)
}

// CollectDBDump exposes the functionality to the diagnostics collect mechanism
func CollectDBDump(kubeconfig string, destDir string) {
	c := newCollectorDbDump(kubeconfig)

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

// Create postgres dump
func (c *CollectorDbDump) generatePostgresDump(destDir string) error {
	// In case of a postgres db, the dump is a single file so in this case we can
	// redirect the output of the dump straight to a local file
	cmd := exec.Command(c.kubeCommand, "exec", "-n", options.Namespace, "-it", "pod/noobaa-db-pg-0", "--", "pg_dumpall")
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
	err := os.Mkdir(c.dbDumpFolderPath, os.ModePerm)
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

	// Redirect the command's output to the local dump file
	defer util.SafeClose(outfile, fmt.Sprintf("Failed to close dump file %s", dumpFilePath))
	cmd.Stdout = outfile

	// Execute the command, generating the dump file
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ cannot generate db dump: %v`, err)
		return err
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
