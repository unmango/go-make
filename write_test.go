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

	Describe("WriteRule", func() {
		DescribeTable("Rules",
			Entry("target",
				&ast.Rule{Targets: []ast.Expr{&ast.Text{Value: "target"}}},
				"target:\n",
			),
			Entry("multiple targets",
				&ast.Rule{Targets: []ast.Expr{
					&ast.Text{Value: "target"},
					&ast.Text{Value: "target2"},
				}},
				"target target2:\n",
			),
			Entry("target with prereq",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "target"}},
					PreReqs: []ast.Expr{&ast.Text{Value: "prereq"}},
				},
				"target: prereq\n",
			),
			Entry("target, prereq, and recipe",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "target"}},
					PreReqs: []ast.Expr{&ast.Text{Value: "prereq"}},
					Recipes: []*ast.Recipe{{
						Prefix: token.TAB,
						Text:   "curl https://example.com",
					}},
				},
				"target: prereq\n\tcurl https://example.com\n",
			),
			Entry("target with recipe",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "target"}},
					Recipes: []*ast.Recipe{{
						Prefix: token.TAB,
						Text:   "curl https://example.com",
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

			_, err := make.WriteRule(w, &ast.Rule{Targets: []ast.Expr{
				&ast.Text{Value: "target"},
			}})
			Expect(err).NotTo(HaveOccurred())
			_, err = make.WriteRule(w, &ast.Rule{Targets: []ast.Expr{
				&ast.Text{Value: "target2"},
			}})
			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target:\ntarget2:\n"))
		})

		It("should ignore nil", func() {
			w := make.NewWriter(&bytes.Buffer{})

			Expect(make.WriteRule(w, nil)).To(Equal(0))
		})

		DescribeTable("should error when rule has no targets",
			Entry("empty rule", &ast.Rule{}),
			Entry("with prereqs", &ast.Rule{PreReqs: []ast.Expr{
				&ast.Text{Value: "foo"},
			}}),
			Entry("with recipes", &ast.Rule{Recipes: []*ast.Recipe{{
				Prefix: token.TAB,
				Text:   "foo",
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
					Targets: []ast.Expr{&ast.Text{Value: "foo"}},
					PreReqs: []ast.Expr{&ast.Text{Value: "bar"}},
					Recipes: []*ast.Recipe{{
						Prefix: token.TAB,
						Text:   "baz",
					}},
				})

				Expect(err).To(MatchError("write err: 5"))
			},
		)
	})

	Describe("WriteFile", func() {
		It("should write a Makefile", func() {
			buf := &bytes.Buffer{}
			w := make.NewWriter(buf)

			_, err := make.WriteFile(w, &ast.File{
				Decls: []ast.Decl{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "target"}},
				}},
			})

			Expect(err).NotTo(HaveOccurred())
		})

		It("should ignore nil", func() {
			w := make.NewWriter(&bytes.Buffer{})

			Expect(make.WriteFile(w, nil)).To(Equal(0))
		})

		It("should return errors found when writing a Makefile", func() {
			w := make.NewWriter(testing.ErrWriter("io error"))

			_, err := make.WriteFile(w, &ast.File{
				Decls: []ast.Decl{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "target"}},
				}},
			})

			Expect(err).To(MatchError("io error"))
		})
	})

	Describe("WriteExpr", func() {
		It("should write text", func() {
			buf := &bytes.Buffer{}
			w := make.NewWriter(buf)

			n, err := w.WriteExpr(&ast.Text{
				Value: "foo",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(3))
			Expect(buf.String()).To(Equal("foo"))
		})

		It("should not write unsupported nodes", func() {
			buf := &bytes.Buffer{}
			w := make.NewWriter(buf)

			n, err := w.WriteExpr(nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(0))
			Expect(buf.Len()).To(Equal(0))
		})
	})

	Describe("WriteDecl", func() {
		It("should ignore nil", func() {
			w := make.NewWriter(&bytes.Buffer{})

			Expect(make.WriteDecl(w, nil)).To(Equal(0))
		})

		It("should write a variable", func() {
			w := make.NewWriter(&bytes.Buffer{})

			n, err := make.WriteDecl(w, &ast.Variable{
				Name:  &ast.Text{Value: "TEST"},
				Op:    token.SIMPLE_ASSIGN,
				OpPos: token.Pos(6),
				Value: []ast.Expr{&ast.Text{
					Value:    "value",
					ValuePos: token.Pos(9),
				}},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(13))
		})
	})

	Describe("WriteVar", func() {
		It("should ignore nil variables", func() {
			w := make.NewWriter(&bytes.Buffer{})

			Expect(make.WriteVar(w, nil)).To(Equal(0))
		})

		When("Value is empty", func() {
			It("should write a variable", func() {
				buf := &bytes.Buffer{}
				w := make.NewWriter(buf)

				n, err := make.WriteVar(w, &ast.Variable{
					Name:  &ast.Text{Value: "TEST"},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(5),
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST:="))
				Expect(n).To(Equal(6))
			})

			It("should write a space-separated variable", func() {
				buf := &bytes.Buffer{}
				w := make.NewWriter(buf)

				n, err := make.WriteVar(w, &ast.Variable{
					Name:  &ast.Text{Value: "TEST"},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(6),
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST :="))
				Expect(n).To(Equal(7))
			})
		})

		When("Value is defined", func() {
			It("should write a variable", func() {
				buf := &bytes.Buffer{}
				w := make.NewWriter(buf)

				n, err := make.WriteVar(w, &ast.Variable{
					Name:  &ast.Text{Value: "TEST"},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(5),
					Value: []ast.Expr{
						&ast.Text{Value: "value"},
					},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST:=value"))
				Expect(n).To(Equal(11))
			})

			It("should write a space-separated variable", func() {
				buf := &bytes.Buffer{}
				w := make.NewWriter(buf)

				n, err := make.WriteVar(w, &ast.Variable{
					Name:  &ast.Text{Value: "TEST"},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(6),
					Value: []ast.Expr{&ast.Text{
						Value:    "value",
						ValuePos: token.Pos(9),
					}},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST := value"))
				Expect(n).To(Equal(13))
			})

			DescribeTable("should return write errors",
				Entry("Expr", 1),
				Entry("Whitespace", 2),
				Entry("Token", 3),
				Entry("Whitespace", 4),
				Entry("Value", 5),
				func(after int) {
					w := make.NewWriter(
						testing.NewErrAfterWriter(after),
					)

					_, err := make.WriteVar(w, &ast.Variable{
						Name:  &ast.Text{Value: "TEST"},
						Op:    token.SIMPLE_ASSIGN,
						OpPos: token.Pos(6),
						Value: []ast.Expr{
							&ast.Text{Value: "value", ValuePos: token.Pos(9)},
						},
					})

					Expect(err).To(MatchError(ContainSubstring("write err:")))
				},
			)
		})
	})
})
