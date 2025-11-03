package plana_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestPlana(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Plana Suite")
}
