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

// CollectorDbStatData configuration for diagnostics
type CollectorDbStatData struct {
	folderName           string // Local folder to store the raw db stat
	dbStatDataFolderPath string // Local path of the folder containing the raw stat file
	remoteTarPath        string // Remote path (at pod) of tar'd db stat
	kubeconfig           string
	kubeCommand          string
	log                  *logrus.Entry
}

func newCollectorDbStatData(kubeconfig string) *CollectorDbStatData {
	c := new(CollectorDbStatData)
	c.folderName = fmt.Sprintf("%s_%d", "noobaa_db_Statdata", time.Now().Unix())
	c.remoteTarPath = fmt.Sprintf("%s/%s%s", "/tmp", c.folderName, ".tar.gz")
	c.log = util.Logger()
	c.kubeconfig = kubeconfig
	c.kubeCommand = util.GetAvailabeKubeCli()
	return c
}

// RunDBStat generates a tar file of a db Statdata
func RunDBStat(cmd *cobra.Command, args []string) {
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	destDir, _ := cmd.Flags().GetString("dir")

	CollectStatData(kubeconfig, destDir)
}

// RunDBStatPrepare generates a tar file of a db Statdata
func RunDBStatPrepare(cmd *cobra.Command, args []string) {
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")

	PrepareStatData(kubeconfig)
}

func PrepareStatData(kubeconfig string) {
	c := newCollectorDbStatData(kubeconfig)

	err := c.prepareDBStatData()
	if err == nil {

	}
}

// CollectStatData exposes the functionality to the diagnostics collect mechanism
func CollectStatData(kubeconfig string, destDir string) {
	c := newCollectorDbStatData(kubeconfig)

	c.log.Printf("Collecting db Statdate destDir : %s", destDir)
	err := c.generateDBStatData(destDir)
	if err == nil {
		err = c.exportDBStatdata(destDir)
		if err == nil {
			tarPath := fmt.Sprintf("%s.tar.gz", c.dbStatDataFolderPath)
			c.log.Printf("✅ Generated db stat was saved in %s\n", tarPath)
			c.deleteDbRawResources()
		}
	}
}

// prepareDBStatData create the stat resources
func (c *CollectorDbStatData) prepareDBStatData() error {

	c.log.Println("Prepare db Statdata at pod")

	// create pg_stat_statements extension and reset
	cmd := exec.Command(c.kubeCommand, "exec", "-n", options.Namespace, "pod/noobaa-db-pg-0", "--", "psql", "-c",
		"create extension pg_stat_statements", "-c", "select pg_stat_reset()", "-c", "select pg_stat_statements_reset()")
	// handle custom path for kubeconfig file,
	// see --kubeconfig cli options
	if len(c.kubeconfig) > 0 {
		cmd.Env = append(cmd.Env, "KUBECONFIG="+c.kubeconfig)
	}

	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ cannot generate db Statdata: %v`, err)
		return err
	}

	return nil
}

// generateDBStatData create the stat resources
func (c *CollectorDbStatData) generateDBStatData(destDir string) error {

	c.log.Println("Generating db Stat data at pod, destDir : ", destDir)

	// In case of a postgres db, the stat is a single file so in this case we can
	// redirect the output of the stat straight to a local file

	// Compose the path of the folder containing the stat file
	if destDir != "" {
		c.dbStatDataFolderPath = fmt.Sprintf("%s/%s", destDir, c.folderName)
	} else {
		c.dbStatDataFolderPath = c.folderName
	}

	// Create the folder containing the stat file
	err := os.Mkdir(c.dbStatDataFolderPath, os.ModePerm)
	if err != nil {
		c.log.Fatalf(`❌ Could not create directory %s, reason: %s`, c.folderName, err)
		return err
	}

	statFilePath := fmt.Sprintf("%s/%s.sql", c.dbStatDataFolderPath, c.folderName)

	// Create the stat file
	outfile, err := os.Create(statFilePath)
	if err != nil {
		c.log.Printf(`❌ can not create db stat at path %v: %v`, c.dbStatDataFolderPath, err)
		return err
	}

	// Redirect the command's output to the local stat file
	defer outfile.Close()

	cmd := exec.Command(c.kubeCommand, "exec", "-n", options.Namespace, "pod/noobaa-db-pg-0", "--", "psql", "-c", "SELECT total_plan_time*'1ms'::interval AS total, mean_plan_time*'1ms'::interval AS avg, calls, query from pg_stat_statements ORDER BY total_plan_time DESC LIMIT 10")
	if len(c.kubeconfig) > 0 {
		cmd.Env = append(cmd.Env, "KUBECONFIG="+c.kubeconfig)
	}
	cmd.Stdout = outfile
	// Execute the command, generating the stat file
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ cannot generate db Statdata: %v`, err)
		return err
	}

	return nil
}

// Tar local stat resources
func (c *CollectorDbStatData) exportDBStatdata(destDir string) error {
	// In case of a postgres, the stat was created as a local file, so all is left to
	// do is tar it
	var cmd *exec.Cmd

	// Compose the path of the tar output
	tarPath := fmt.Sprintf("%s.tar.gz", c.dbStatDataFolderPath)

	// Genrate the tar command
	if destDir != "" {
		cmd = exec.Command("tar", "-C", destDir, "-cvzf", tarPath, c.folderName)
	} else {
		cmd = exec.Command("tar", "-cvzf", tarPath, c.folderName)
	}
	c.log.Printf("tarPath %s, cmd : %v", tarPath, cmd)
	// Execute the tar command
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ failed to tar db stat`)
		return err
	}
	return nil
}

// deleteDbRawResources stat on Postgres remote machine
func (c *CollectorDbStatData) deleteDbRawResources() {

	// Compose the deletion command
	cmd := exec.Command("rm", "-rf", c.dbStatDataFolderPath)

	// Execute the deletion of raw resources
	if err := cmd.Run(); err != nil {
		c.log.Printf(`❌ failed to tar remove stat folder`)
	}
	//time.Sleep(2 * time.Second)

	cmd_select_drop := exec.Command(c.kubeCommand, "exec", "-n", options.Namespace, "pod/noobaa-db-pg-0", "--", "psql", "-c", "drop extension pg_stat_statements")
	if len(c.kubeconfig) > 0 {
		cmd_select_drop.Env = append(cmd_select_drop.Env, "KUBECONFIG="+c.kubeconfig)
	}
	// Execute the command, delete the extension
	if err := cmd_select_drop.Run(); err != nil {
		c.log.Printf(`❌ cannot generate db Statdata: %v`, err)
	}
}
