package variable_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestVariable(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Variable Suite")
}
