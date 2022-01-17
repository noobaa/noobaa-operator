package dbdump

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

// Cmd returns a CLI command
func Cmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "db-dump",
		Short: "Collect db dump",
		Run:   RunDump,
		Args:  cobra.NoArgs,
	}
	cmd.Flags().String("dir", "", "collect db dump file into destination directory")
	return cmd
}

// Collector configuration for diagnostics
type Collector struct {
	folderName			string // Local folder to store the raw db dump
	dbDumpFolderPath	string // Local path of the folder containing the raw dump file
	remoteFolderPath	string // Remote folder (at pod) to store the raw db dump
	remoteTarPath		string // Remote path (at pod) of tar'd db dump
	kubeconfig 			string
	kubeCommand			string
	log					*logrus.Entry
}

func newCollector(kubeconfig string) *Collector {
	c := new(Collector)
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

// CollectDBDump exposes the functionality to the diagnose mechanism
func CollectDBDump(kubeconfig string, destDir string) {
	c := newCollector(kubeconfig)

	c.log.Println("Collecting db dump")

	err := c.generateDBDump(destDir)
	if(err == nil) {
		err = c.exportDBDump(destDir)
		if(err == nil) {
			c.deleteRawResources()
		}
	}
}

// Create postgres dump
func (c *Collector) generatePostgresDump(destDir string) error {
	// In case of a postgres db, the dump is a single file so in this case we can
	// redirect the output of the dump straight to a local file
	cmd := exec.Command(c.kubeCommand, "exec", "-it", "pod/noobaa-db-pg-0", "--", "pg_dumpall")
	// handle custom path for kubeconfig file,
	// see --kubeconfig cli options
	if len(c.kubeconfig) > 0 {
		cmd.Env = append(cmd.Env, "KUBECONFIG=" + c.kubeconfig)
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
	defer outfile.Close()
	cmd.Stdout = outfile

	// Execute the command, generating the dump file
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ cannot generate db dump: %v`, err)
		return err
	}

	return nil
}

// Create mongodb dump
func (c *Collector) generateMongoDBDump(destDir string) error {
	// In case of a mongo db, the dump is a dir containing multiple entries, in the case we
	// can't redirect the output to a local file, so we generate the dump on the remote machine
	// and tar it

	// Compose the path containing the dump folder at the remote machine
	c.remoteFolderPath = fmt.Sprintf("%s/%s", "/tmp", c.folderName)
	cmd := exec.Command(c.kubeCommand, "exec", "pod/noobaa-db-0", "--", "mongodump", "--db", "nbcore", "-o", c.remoteFolderPath)
	// handle custom path for kubeconfig file,
	// see --kubeconfig cli options
	if len(c.kubeconfig) > 0 {
		cmd.Env = append(cmd.Env, "KUBECONFIG=" + c.kubeconfig)
	}

	// Generate db dump folder
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ cannot generate db dump: %v`, err)
		return err
	}

	// Tar the raw db dump
	cmd = exec.Command(c.kubeCommand, "exec", "pod/noobaa-db-0", "--", "tar", "-C", "/tmp", "-cvzf", c.remoteTarPath, c.folderName)
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ failed to tar remote dump folder`)
		return err
	}

	return nil
}

// Create the dump resources
func (c *Collector) generateDBDump(destDir string) error {
	var err error

	c.log.Println("Generating db dump at pod")

	if options.DBType == "postgres" {
		err = c.generatePostgresDump(destDir)
	} else {
		err = c.generateMongoDBDump(destDir)
	}

	return err
}

// Tar local dump resources
func (c *Collector) exportPostgresDump(destDir string) error {
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

// Copy dump resources from remote machine 
func (c *Collector) exportMongoDBDump(destDir string) error {
	// In case of a mongo db, the db dump was generated and tar'd at the remote machine,
	// all is left to do is copy the tarball to the local machine
	var localPath string

	// Compose the local path to export the tarball
	if destDir != "" {
		localPath = fmt.Sprintf("%s/%s%s", destDir, c.folderName, ".tar.gz")
	} else {
		localPath = fmt.Sprintf("%s%s", c.folderName, ".tar.gz")
	}

	// Compose the full path to copy the tarball from
	fullTarPath := fmt.Sprintf("%s:%s", "noobaa-db-0", c.remoteTarPath)
		
	// Create the command to copy the tarball
	cmd := exec.Command(c.kubeCommand, "cp", fullTarPath, localPath)
	// handle custom path for kubeconfig file,
	// see --kubeconfig cli options
	if len(c.kubeconfig) > 0 {
		cmd.Env = append(cmd.Env, "KUBECONFIG=" + c.kubeconfig)
	}

	// Execute the command, copying the tarball
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ failed to copy remote tar file`)
		return err
	}

	return nil
}

// Tar local resources (Postgres) or copy from the remote machine (mongodb)
func (c *Collector) exportDBDump(destDir string) error {
	var err error

	c.log.Println("Exporting db dump")

	if options.DBType == "postgres" {
		err = c.exportPostgresDump(destDir)
	} else {
		err = c.exportMongoDBDump(destDir)
	}

	return err
}

// Delete dump on Postgres remote machine
func (c *Collector) deletePostgresRawResources() {
	// In case of postgres, the db dump was created locall, so all is left
	// to do is remove the raw dump

	// Compose the deletion command
	cmd := exec.Command("rm", "-rf", c.dbDumpFolderPath)

	// Execute the deletion of raw resources
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ failed to tar remove dump folder`)
	}
}

// Delete dump on mongodb remote machine
func (c *Collector) deleteMongoDBRawResources() {
	// In case of mongo db, the dump was created and tar'd at the remote machine,
	// all is left to do is clean the remote machine

	// Compose the deletion command of the raw resources
	cmd := exec.Command(c.kubeCommand, "exec", "pod/noobaa-db-0", "--", "rm", "-rf", c.remoteFolderPath)

	// Execute deletion of the raw resources
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ failed to remove remote dump folder`)
	}

	// Compose the deletion command of the tar'd resources
	cmd = exec.Command(c.kubeCommand, "exec", "pod/noobaa-db-0", "--", "rm", c.remoteTarPath)

	// Execute deletion of the tar'd resources	
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ failed to tar remove dump tar file`)
	}
}

// Remove raw resources from local or remote machines
func (c *Collector) deleteRawResources() {
	if options.DBType == "postgres" {
		c.deletePostgresRawResources()
	} else {
		c.deleteMongoDBRawResources()
	}
}
