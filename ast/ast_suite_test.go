package ast_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestAst(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Ast Suite")
}
