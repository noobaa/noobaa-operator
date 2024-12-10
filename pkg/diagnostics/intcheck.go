package diagnostics

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"

	"github.com/spf13/cobra"

	"github.com/noobaa/noobaa-operator/v5/pkg/options"
	"github.com/noobaa/noobaa-operator/v5/pkg/util"
)

const defaultDBName = "nbcore"

func RunIntCheck(cmd *cobra.Command, args []string) {
	kubeconfig, _ := cmd.Flags().GetString("kubeconfig")
	dataDump, _ := cmd.Flags().GetBool("dump-data-map")

	dbname := os.Getenv("NOOBAA_DB")
	if dbname == "" {
		dbname = defaultDBName
	}

	runIntCheck(kubeconfig, dbname, dataDump)
}

func runIntCheck(kubeconfig, dbname string, dataDump bool) {
	kubecommand := util.GetAvailabeKubeCli()
	objectmds, err := getObjectMds(kubecommand, dbname, 0)
	if err != nil {
		log.Fatalln(err)
	}

	partsIntegrityIssues := 0
	for _, objectmd := range objectmds {
		parts_data, err := getObjectParts(kubecommand, dbname, objectmd["_id"].(string))
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
			chunksdata, err := getChunkData(kubecommand, dbname, objpart["chunk"].(string))
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

				realFragData, err := getFragsData(kubecommand, dbname, fragIDs)
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
		json.NewEncoder(os.Stdout).Encode(objectmds)
	}
}

func getObjectMds(kubecommand, dbname string, page int) ([]map[string]any, error) {
	sql := fmt.Sprintf("SELECT json_agg(data) FROM objectmds LIMIT 1000 OFFSET %d;", page*1000)
	return getJSONDataFromDB(kubecommand, dbname, sql)
}

func getObjectParts(kubecommand, dbname, objectID string) ([]map[string]any, error) {
	sql := fmt.Sprintf("SELECT json_agg(data) FROM objectparts WHERE data->>'obj' = '%s';", objectID)
	return getJSONDataFromDB(kubecommand, dbname, sql)
}

func getChunkData(kubecommand, dbname, chunkID string) ([]map[string]any, error) {
	sql := fmt.Sprintf("SELECT json_agg(data) FROM datachunks WHERE data->>'_id' = '%s';", chunkID)
	return getJSONDataFromDB(kubecommand, dbname, sql)
}

func getFragsData(kubecommand, dbname string, frags []string) ([]map[string]any, error) {
	ids := ""
	for idx, frag := range frags {
		ids += ("'" + frag + "'")
		if idx != len(frags)-1 {
			ids += ","
		}
	}

	sql := fmt.Sprintf("SELECT json_agg(data) FROM datablocks WHERE data->>'frag' IN (%s);", ids)
	return getJSONDataFromDB(kubecommand, dbname, sql)
}

func getJSONDataFromDB(kubecommand, dbname, sql string) ([]map[string]any, error) {
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
