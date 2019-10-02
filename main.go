// Package main is the top level package in the noobaa-operator project
// which means that running bare go commands like `go generate` and `go build`
// will refer to this main package.
package main

//go:generate make gen

import (
	"github.com/noobaa/noobaa-operator/v2/pkg/cli"
	"github.com/noobaa/noobaa-operator/v2/pkg/util"
)

func main() {
	util.IgnoreError(cli.Cmd().Execute())
}
