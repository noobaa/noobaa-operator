package obc_test

import (
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestObc(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Obc Suite")
}
