package diagnostics

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"time"

	"github.com/spf13/cobra"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

const defaultDBName = "nbcore"

var dbStatDataFolderPath = ""

func RunIntCheck(cmd *cobra.Command, args []string) {
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	dataDump, _ := cmd.Flags().GetBool("dump-data-map")
	destDir, _ := cmd.Flags().GetString("dir")

	dbname := os.Getenv("NOOBAA_DB")
	if dbname == "" {
		dbname = defaultDBName
	}

	runIntCheck(kubeconfig, dbname, dataDump, destDir)
}

func runIntCheck(kubeconfig, dbname string, dataDump bool, destDir string) {
	kubecommand := util.GetAvailabeKubeCli()

	var folderName = fmt.Sprintf("%s_%d", "noobaa_db_int_check", time.Now().Unix())
	if destDir != "" {
		dbStatDataFolderPath = fmt.Sprintf("%s/%s", destDir, folderName)
	} else {
		dbStatDataFolderPath = folderName
	}

	// Create the folder containing the stat file
	err := os.Mkdir(dbStatDataFolderPath, os.ModePerm)
	if err != nil {
		log.Fatalf(`❌ Could not create directory %s, reason: %s`, folderName, err)
	}

	statFilePath := fmt.Sprintf("%s/%s.sql", dbStatDataFolderPath, folderName)

	// Create the stat file
	outfile, err := os.Create(statFilePath)
	if err != nil {
		log.Printf(`❌ can not create db stat at path %v: %v`, dbStatDataFolderPath, err)
	}

	// Redirect the command's output to the local stat file
	defer outfile.Close()

	objectmds, err := getObjectMds(kubeconfig, kubecommand, dbname, 0)
	if err != nil {
		log.Fatalln(err)
	}

	partsIntegrityIssues := 0
	for _, objectmd := range objectmds {
		parts_data, err := getObjectParts(kubeconfig, kubecommand, dbname, objectmd["_id"].(string))
		if err != nil {
			log.Fatalln("failed to get object part data for object:", objectmd)
		}

		if len(parts_data) != int(objectmd["num_parts"].(float64)) {
			partsIntegrityIssues += 1
			log.Printf(
				"invalid object parts found for object: %s - expected: %d, got: %d\n",
				objectmd["_id"].(string),
				int(objectmd["num_parts"].(float64)),
				len(parts_data),
			)
		}

		objectmd["parts_data"] = parts_data
	}

	log.Println("Object Parts integrity issues - ", partsIntegrityIssues)

	chunkIntegrityIssues := 0
	for _, objectmd := range objectmds {
		for _, objpart := range objectmd["parts_data"].([]map[string]any) {
			chunksdata, err := getChunkData(kubeconfig, kubecommand, dbname, objpart["chunk"].(string))
			if err != nil {
				log.Fatalf(
					"failed to get chunk data for part %s of object %s\n",
					objpart["_id"].(string),
					objectmd["_id"].(string),
				)
			}

			if len(chunksdata) == 0 {
				chunkIntegrityIssues += 1
				log.Println("unexpected chunks 0 for part id -", objpart["_id"].(string))
			}

			fragsIntegrityIssues := 0
			objpart["chunks_data"] = chunksdata

			for _, chunkdata := range chunksdata {
				frags, ok := chunkdata["frags"].([]any)
				if !ok {
					fragsIntegrityIssues += 1
					log.Println("invalid fragment data found for chunk -", chunkdata["_id"])
				}

				fragIDs := []string{}
				for _, frag := range frags {
					frag := frag.(map[string]any)
					fragIDs = append(fragIDs, frag["_id"].(string))
				}

				realFragData, err := getFragsData(kubeconfig, kubecommand, dbname, fragIDs)
				if err != nil {
					fragsIntegrityIssues += 1
					log.Println("failed to fragment data for chunk - ", chunkdata["_id"])
					continue
				}

				if len(realFragData) != len(frags) {
					chunkIntegrityIssues += 1
					log.Printf("unexpected frags %v for part id - %s", len(realFragData), objpart["_id"].(string))
					continue
				}
			}

			log.Println("Object Fragments integrity issues - ", fragsIntegrityIssues)
		}
	}

	log.Println("Object Chunks integrity issues - ", chunkIntegrityIssues)

	if dataDump {
		json.NewEncoder(outfile).Encode(objectmds)
	}
}

func getObjectMds(kubeconfig, kubecommand, dbname string, page int) ([]map[string]any, error) {
	sql := fmt.Sprintf("SELECT json_agg(data) FROM objectmds LIMIT 1000 OFFSET %d;", page*1000)
	return getJSONDataFromDB(kubeconfig, kubecommand, dbname, sql)
}

func getObjectParts(kubeconfig, kubecommand, dbname, objectID string) ([]map[string]any, error) {
	sql := fmt.Sprintf("SELECT json_agg(data) FROM objectparts WHERE data->>'obj' = '%s';", objectID)
	return getJSONDataFromDB(kubeconfig, kubecommand, dbname, sql)
}

func getChunkData(kubeconfig, kubecommand, dbname, chunkID string) ([]map[string]any, error) {
	sql := fmt.Sprintf("SELECT json_agg(data) FROM datachunks WHERE data->>'_id' = '%s';", chunkID)
	return getJSONDataFromDB(kubeconfig, kubecommand, dbname, sql)
}

func getFragsData(kubeconfig, kubecommand, dbname string, frags []string) ([]map[string]any, error) {
	ids := ""
	for idx, frag := range frags {
		ids += ("'" + frag + "'")
		if idx != len(frags)-1 {
			ids += ","
		}
	}

	sql := fmt.Sprintf("SELECT json_agg(data) FROM datablocks WHERE data->>'frag' IN (%s);", ids)
	return getJSONDataFromDB(kubeconfig, kubecommand, dbname, sql)
}

func getJSONDataFromDB(kubeconfig, kubecommand, dbname, sql string) ([]map[string]any, error) {
	cmd := exec.Command(
		kubecommand,
		"exec",
		"-n", options.Namespace,
		"pod/noobaa-db-pg-0",
		"--",
		"psql",
		"-d",
		dbname,
		"-c",
		sql,
		"-q",
		"-t",
	)
	if len(kubeconfig) > 0 {
		cmd.Env = append(cmd.Env, "KUBECONFIG="+kubeconfig)
	}

	output, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	parsed := []map[string]any{}
	if err := json.Unmarshal(output, &parsed); err != nil {
		return nil, err
	}

	return parsed, err
}
