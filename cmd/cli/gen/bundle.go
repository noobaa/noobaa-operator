package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/noobaa/noobaa-operator/version"
)

const (
	backtick        = "`"
	backtickReplace = "` + \"`\" + `"
	nameRE          = `[\\.,/?:;'"|\-+=~!@#$%^&*()<>{}\[\]]`
	root            = "../../"
)

var compiledNameRE = regexp.MustCompile(nameRE)

func main() {

	log.Printf("CLI GEN BUNDLE: Starting ...\n")

	files := []string{}
	err := filepath.Walk(root+"deploy/",
		func(path string, info os.FileInfo, err error) error {
			if err == nil && !info.IsDir() {
				files = append(files, path)
			}
			return err
		},
	)

	w, err := os.Create("bundle/bundle.go")
	fatal(err)
	write(w, "package bundle\n\n")
	writef(w, "const Version = \"%s\"\n\n", version.Version)

	for _, path := range files {
		name := compiledNameRE.ReplaceAllString(strings.TrimPrefix(path, root), "_")
		bytes, err := ioutil.ReadFile(path)
		fatal(err)
		sha256Bytes := sha256.Sum256(bytes)
		sha256Hex := hex.EncodeToString(sha256Bytes[:])
		log.Printf("CLI GEN BUNDLE: Adding name:%s size:%d sha256:%s\n",
			name, len(bytes), sha256Hex)
		writef(w, "const Sha256_%s = \"%s\"\n\n", name, sha256Hex)
		writef(w, "const File_%s = `", name)
		write(w, strings.ReplaceAll(string(bytes), backtick, backtickReplace))
		write(w, "`\n\n")
	}

	err = w.Close()
	fatal(err)
	log.Printf("CLI GEN BUNDLE: Done.\n")
}

func fatal(err error) {
	if err != nil {
		log.Fatalln(err)
	}
}

func write(w io.Writer, s string) {
	_, err := io.WriteString(w, s)
	fatal(err)
}

func writef(w io.Writer, format string, args ...interface{}) {
	write(w, fmt.Sprintf(format, args...))
}
