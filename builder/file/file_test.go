package file_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/file"
	"github.com/unmango/go-make/builder/rule"
)

var _ = Describe("File", func() {
	Describe("New", func() {
		It("should work", func() {
			f := file.New(1,
				file.Rule(rule.TextTarget("target")),
				file.Rule(rule.TextTarget("target2")),
			)

			Expect(f).To(Equal(&ast.File{
				FileStart: 1,
				Contents: []ast.Obj{
					rule.New(1, rule.TextTarget("target")),
					rule.New(9, rule.TextTarget("target2")),
				},
				FileEnd: 15,
			}))
		})
	})

	Describe("InsertRule", func() {
		It("should work", func() {
			f := file.New(1,
				file.Rule(rule.TextTarget("target")),
				file.Rule(rule.TextTarget("target3")),
			)

			file.InsertRule(f, 1, rule.TextTarget("target2"))

			Expect(f).To(Equal(&ast.File{
				FileStart: 1,
				Contents: []ast.Obj{
					&ast.Rule{
						Targets: []ast.Expr{
							&ast.Text{Value: "target", ValuePos: 1},
						},
						Colon: 7,
					},
					&ast.Rule{
						Targets: []ast.Expr{
							&ast.Text{Value: "target2", ValuePos: 9},
						},
						Colon: 15,
					},
					&ast.Rule{
						Targets: []ast.Expr{
							&ast.Text{Value: "target3", ValuePos: 17},
						},
						Colon: 24,
					},
				},
			}))
		})
	})
})
