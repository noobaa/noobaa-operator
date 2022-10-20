package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/noobaa/noobaa-operator/v5/pkg/util"
	"github.com/noobaa/noobaa-operator/v5/version"
	"github.com/sirupsen/logrus"
)

const (
	backtick        = "`"
	backtickReplace = "` + \"`\" + `"
	nameRE          = `[\\.,/?:;'"|\-+=~!@#$%^&*()<>{}\[\]]`
)

var compiledNameRE = regexp.MustCompile(nameRE)

func main() {
	util.InitLogger(logrus.DebugLevel)

	src := os.Args[1]
	out := os.Args[2]
	logrus.Printf("bundle files in %s writing to %s\n", src, out)

	w, err := os.Create(out)
	fatal(err)
	write(w, "package bundle\n\n")
	writef(w, "const Version = %q\n\n", version.Version)

	err = filepath.Walk(src,
		func(path string, info os.FileInfo, err error) error {
			fatal(err)
			if info.IsDir() {
				return nil
			}
			name := compiledNameRE.ReplaceAllString(path, "_")
			bytes, err := os.ReadFile(filepath.Clean(path))
			fatal(err)
			sha256Bytes := sha256.Sum256(bytes)
			sha256Hex := hex.EncodeToString(sha256Bytes[:])
			logrus.Printf("bundle file %s size:%d sha256:%s\n",
				path, len(bytes), sha256Hex)
			writef(w, "const Sha256_%s = %q\n\n", name, sha256Hex)
			writef(w, "const File_%s = `", name)
			write(w, strings.ReplaceAll(string(bytes), backtick, backtickReplace))
			write(w, "`\n\n")
			return nil
		},
	)
	fatal(err)

	err = w.Close()
	fatal(err)
	logrus.Printf("bundle - done.\n")
}

func fatal(err error) {
	if err != nil {
		logrus.Fatalln(err)
	}
}

func write(w io.Writer, s string) {
	_, err := io.WriteString(w, s)
	fatal(err)
}

func writef(w io.Writer, format string, args ...interface{}) {
	write(w, fmt.Sprintf(format, args...))
}
