package builder_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/builder/rule"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Apply", func() {
	Describe("ApplyRule", func() {
		It("should add a target", func() {
			node := &ast.Rule{}

			r := builder.ApplyRule(node, rule.WithTextTarget("target"))

			Expect(r).To(Equal(&ast.Rule{
				Targets: []ast.Expr{
					&ast.Text{Value: "target", ValuePos: token.Pos(1)},
				},
				Colon: token.Pos(7),
			}))
		})

		When("rule has existing targets", func() {
			It("should add a target", func() {
				node := &ast.Rule{Targets: []ast.Expr{
					&ast.Text{Value: "target", ValuePos: token.Pos(1)},
				}}

				r := builder.ApplyRule(node, rule.WithTextTarget("target2"))

				Expect(r).To(Equal(&ast.Rule{
					Targets: []ast.Expr{
						&ast.Text{Value: "target", ValuePos: token.Pos(1)},
						&ast.Text{Value: "target2", ValuePos: token.Pos(8)},
					},
					Colon: token.Pos(15),
				}))
			})

			It("should add a target with a different start", func() {
				node := &ast.Rule{Targets: []ast.Expr{
					&ast.Text{Value: "target", ValuePos: token.Pos(69)},
				}}

				r := builder.ApplyRule(node, rule.WithTextTarget("target2"))

				Expect(r).To(Equal(&ast.Rule{
					Targets: []ast.Expr{
						&ast.Text{Value: "target", ValuePos: token.Pos(69)},
						&ast.Text{Value: "target2", ValuePos: token.Pos(76)},
					},
					Colon: token.Pos(83),
				}))
			})
		})
	})
})
