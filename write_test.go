package make_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/internal/testing"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Write", func() {
	It("should write a line", func() {
		buf := &bytes.Buffer{}
		w := make.NewWriter(buf)

		n, err := w.WriteLine()

		Expect(err).NotTo(HaveOccurred())
		Expect(buf.String()).To(Equal("\n"))
		Expect(n).To(Equal(1))
	})

	It("should write multiple targets", func() {
		buf := &bytes.Buffer{}
		w := make.NewWriter(buf)

		n, err := make.WriteTargetList(w, &ast.TargetList{
			List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{Name: "target"}},
				&ast.LiteralFileName{Name: &ast.Ident{Name: "target2"}},
			},
		})

		Expect(err).NotTo(HaveOccurred())
		Expect(buf.String()).To(Equal("target target2:"))
		Expect(n).To(Equal(15))
	})

	Describe("WriteRule", func() {
		DescribeTable("Rules",
			Entry("target",
				&ast.Rule{Targets: &ast.TargetList{List: []ast.FileName{
					&ast.LiteralFileName{Name: &ast.Ident{Name: "target"}},
				}}},
				"target:\n",
			),
			Entry("multiple targets",
				&ast.Rule{Targets: &ast.TargetList{List: []ast.FileName{
					&ast.LiteralFileName{Name: &ast.Ident{Name: "target"}},
					&ast.LiteralFileName{Name: &ast.Ident{Name: "target2"}},
				}}},
				"target target2:\n",
			),
			Entry("target with prereq",
				&ast.Rule{
					Targets: &ast.TargetList{List: []ast.FileName{
						&ast.LiteralFileName{Name: &ast.Ident{Name: "target"}},
					}},
					PreReqs: &ast.PreReqList{List: []ast.FileName{
						&ast.LiteralFileName{Name: &ast.Ident{Name: "prereq"}},
					}},
				},
				"target: prereq\n",
			),
			Entry("target, prereq, and recipe",
				&ast.Rule{
					Targets: &ast.TargetList{List: []ast.FileName{
						&ast.LiteralFileName{Name: &ast.Ident{Name: "target"}},
					}},
					PreReqs: &ast.PreReqList{List: []ast.FileName{
						&ast.LiteralFileName{Name: &ast.Ident{Name: "prereq"}},
					}},
					Recipes: []*ast.Recipe{{
						Tok:  token.TAB,
						Text: "curl https://example.com",
					}},
				},
				"target: prereq\n\tcurl https://example.com\n",
			),
			Entry("target with recipe",
				&ast.Rule{
					Targets: &ast.TargetList{List: []ast.FileName{
						&ast.LiteralFileName{Name: &ast.Ident{Name: "target"}},
					}},
					Recipes: []*ast.Recipe{{
						Tok:  token.TAB,
						Text: "curl https://example.com",
					}},
				},
				"target:\n\tcurl https://example.com\n",
			),
			func(r *ast.Rule, expected string) {
				buf := &bytes.Buffer{}
				w := make.NewWriter(buf)

				n, err := make.WriteRule(w, r)

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal(expected))
				Expect(n).To(Equal(len(expected)))
			},
		)

		It("should write multiple rules", func() {
			buf := &bytes.Buffer{}
			w := make.NewWriter(buf)

			_, err := make.WriteRule(w, &ast.Rule{Targets: &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{Name: "target"}},
			}}})
			Expect(err).NotTo(HaveOccurred())
			_, err = make.WriteRule(w, &ast.Rule{Targets: &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{Name: "target2"}},
			}}})
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target:\ntarget2:\n"))
		})

		DescribeTable("should error when rule has no targets",
			Entry("empty rule", &ast.Rule{}),
			Entry("with prereqs", &ast.Rule{PreReqs: &ast.PreReqList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{Name: "foo"}},
			}}}),
			Entry("with recipes", &ast.Rule{Recipes: []*ast.Recipe{{
				Tok:  token.TAB,
				Text: "foo",
			}}}),
			func(rule *ast.Rule) {
				buf := &bytes.Buffer{}
				w := make.NewWriter(buf)

				_, err := make.WriteRule(w, rule)

				Expect(err).To(MatchError("no targets"))
			},
		)

		DescribeTable("should return errors",
			Entry("target", 1),
			Entry("prereq", 2),
			Entry("newline", 3),
			Entry("recipe", 4),
			Entry("newline", 5),
			func(position int) {
				writer := testing.NewErrAfterWriter(5)
				w := make.NewWriter(writer)

				_, err := make.WriteRule(w, &ast.Rule{
					Targets: &ast.TargetList{List: []ast.FileName{
						&ast.LiteralFileName{Name: &ast.Ident{Name: "foo"}},
					}},
					PreReqs: &ast.PreReqList{List: []ast.FileName{
						&ast.LiteralFileName{Name: &ast.Ident{Name: "bar"}},
					}},
					Recipes: []*ast.Recipe{{
						Tok:  token.TAB,
						Text: "baz",
					}},
				})

				Expect(err).To(MatchError("write err: 5"))
			},
		)
	})

	It("should write a Makefile", func() {
		buf := &bytes.Buffer{}
		w := make.NewWriter(buf)

		_, err := make.WriteFile(w, &ast.File{
			Rules: []*ast.Rule{{
				Targets: &ast.TargetList{List: []ast.FileName{
					&ast.LiteralFileName{Name: &ast.Ident{Name: "target"}},
				}},
			}},
		})

		Expect(err).NotTo(HaveOccurred())
	})

	It("should return errors found when writing a Makefile", func() {
		w := make.NewWriter(testing.ErrWriter("io error"))

		_, err := make.WriteFile(w, &ast.File{
			Rules: []*ast.Rule{{
				Targets: &ast.TargetList{List: []ast.FileName{
					&ast.LiteralFileName{Name: &ast.Ident{Name: "target"}},
				}},
			}},
		})

		Expect(err).To(MatchError("io error"))
	})
})
