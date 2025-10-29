package plana_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestArona(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Arona Suite")
}
