package rule_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/rule"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Rule", func() {
	Describe("New", func() {
		It("should work", func() {
			r := rule.New(1)

			Expect(r).NotTo(BeNil())
		})

		It("should apply the given builders", func() {
			var expected *ast.Rule

			r := rule.New(1, func(p token.Pos, r *ast.Rule) {
				expected = r
			})

			Expect(r).To(BeIdenticalTo(expected))
		})
	})
})
