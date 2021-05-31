// Package main is the top level package in the noobaa-operator project
// which means that running bare go commands like `go generate` and `go build`
// will refer to this main package.
package main

//go:generate make gen

import (
	"github.com/noobaa/noobaa-operator/v5/pkg/cli"
)

func main() {
	cli.Run()
}
