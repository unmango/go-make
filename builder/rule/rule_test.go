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

		It("should position the colon", func() {
			r := rule.New(1)

			Expect(r.Colon).To(Equal(token.Pos(1)))
		})
	})

	Describe("RePos", func() {
		It("should reposition a rule with a text target", func() {
			r := rule.New(1, rule.TextTarget("test"))
			rule.RePos(2, r)

			Expect(r.Pos()).To(Equal(token.Pos(2)))
			Expect(r.Targets[0].Pos()).To(Equal(token.Pos(2)))
			Expect(r.Colon).To(Equal(token.Pos(6)))
		})

		It("should reposition a rule with a text prereq", func() {
			r := rule.New(1, rule.TextTarget("test"), rule.TextPreReq("test"))
			rule.RePos(2, r)

			Expect(r.Pos()).To(Equal(token.Pos(2)))
			Expect(r.Targets[0].Pos()).To(Equal(token.Pos(2)))
		})
	})
})
