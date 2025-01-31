package make_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/internal/testing"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Print", func() {
	Describe("Fprint", func() {
		It("should print text", func() {
			buf := &bytes.Buffer{}
			l := &ast.Text{Value: "target"}

			err := make.Fprint(buf, l)

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target"))
		})

		It("should print a recipe", func() {
			buf := &bytes.Buffer{}
			r := &ast.Recipe{
				Prefix: token.TAB,
				Text:   "recipe",
			}

			err := make.Fprint(buf, r)

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("\trecipe\n"))
		})

		It("should print a rule", func() {
			buf := &bytes.Buffer{}
			r := &ast.Rule{
				Targets: []ast.Expr{
					&ast.Text{Value: "target"},
				},
				PreReqs: []ast.Expr{
					&ast.Text{Value: "prereq"},
				},
				Recipes: []*ast.Recipe{{
					Prefix: token.TAB,
					Text:   "recipe",
				}},
			}

			err := make.Fprint(buf, r)

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target: prereq\n\trecipe\n"))
		})

		It("should print a file", func() {
			buf := &bytes.Buffer{}
			r := &ast.Rule{
				Targets: []ast.Expr{
					&ast.Text{Value: "target"},
				},
				PreReqs: []ast.Expr{
					&ast.Text{Value: "prereq"},
				},
				Recipes: []*ast.Recipe{{
					Prefix: token.TAB,
					Text:   "recipe",
				}},
			}
			f := &ast.File{Decls: []ast.Decl{r}}

			err := make.Fprint(buf, f)

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target: prereq\n\trecipe\n"))
		})

		DescribeTable("should surface write errors",
			Entry("target", 1),
			Entry("colon", 2),
			Entry("space", 3),
			Entry("prereq", 4),
			Entry("newline", 5),
			Entry("tab", 6),
			func(position int) {
				w := testing.NewErrAfterWriter(position)
				r := &ast.Rule{
					Targets: []ast.Expr{
						&ast.Text{Value: "target"},
					},
					PreReqs: []ast.Expr{
						&ast.Text{Value: "prereq"},
					},
					Recipes: []*ast.Recipe{{
						Prefix: token.TAB,
						Text:   "recipe",
					}},
				}

				err := make.Fprint(w, r)

				Expect(err).To(MatchError(fmt.Sprintf("write err: %d", position)))
			},
		)
	})
})
