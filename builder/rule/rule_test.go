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

		It("should position the colon after the last target", func() {
			r := rule.New(1, rule.TextTarget("test"))

			Expect(r.Colon).To(Equal(token.Pos(5)))
		})
	})

	Describe("Copy", func() {
		It("should copy an empty rule", func() {
			r := rule.New(1)

			actual := rule.Copy(2, r)

			Expect(actual).To(Equal(&ast.Rule{
				Colon: 2,
			}))
		})

		It("should copy a rule with a target", func() {
			r := rule.New(1, rule.TextTarget("test"))

			actual := rule.Copy(2, r)

			Expect(actual).To(Equal(&ast.Rule{
				Targets: []ast.Expr{
					&ast.Text{Value: "test", ValuePos: 2},
				},
				Colon: 6,
			}))
		})

		It("should copy a rule with multiple targets", func() {
			r := rule.New(1,
				rule.TextTarget("test"),
				rule.TextTarget("test2"),
			)

			actual := rule.Copy(2, r)

			Expect(actual).To(Equal(&ast.Rule{
				Targets: []ast.Expr{
					&ast.Text{Value: "test", ValuePos: 2},
					&ast.Text{Value: "test2", ValuePos: 7},
				},
				Colon: 12,
			}))
		})

		It("should copy a rule with a pre-requisite", func() {
			r := rule.New(1, rule.TextPreReq("test"))

			actual := rule.Copy(2, r)

			Expect(actual).To(Equal(&ast.Rule{
				Colon: 2,
				PreReqs: []ast.Expr{
					&ast.Text{Value: "test", ValuePos: 4},
				},
			}))
		})

		It("should copy a rule with multiple pre-requisites", func() {
			r := rule.New(1,
				rule.TextPreReq("test"),
				rule.TextPreReq("test2"),
			)

			actual := rule.Copy(2, r)

			Expect(actual).To(Equal(&ast.Rule{
				Colon: 2,
				PreReqs: []ast.Expr{
					&ast.Text{Value: "test", ValuePos: 4},
					&ast.Text{Value: "test2", ValuePos: 9},
				},
			}))
		})
	})

	Describe("RePos", Pending, func() {
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
