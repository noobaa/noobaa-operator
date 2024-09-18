package obc_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestObc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Obc Suite")
}
