// Package main is the one created initially by operator-sdk new.
// We moved the scafold to pkg/controller/manager.go and call
// the CLI package to run the command tree.
// To run the operator itself use the CLI command args: `operator run`.
package main

//go:generate make gen

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/cli"
)

func main() {
	cli.Run()
}
