package builder_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/builder/rule"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Mut", func() {
	Describe("AtRule", func() {
		It("should add a target", func() {
			node := &ast.Rule{}

			r := builder.AtRule(node, rule.WithTextTarget("target"))

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

				r := builder.AtRule(node, rule.WithTextTarget("target2"))

				Expect(r).To(Equal(&ast.Rule{
					Targets: []ast.Expr{
						&ast.Text{Value: "target", ValuePos: token.Pos(1)},
						&ast.Text{Value: "target2", ValuePos: token.Pos(8)},
					},
					Colon: token.Pos(15),
				}))
			})
		})
	})
})
