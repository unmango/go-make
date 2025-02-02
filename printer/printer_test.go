package printer_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/internal/testing"
	"github.com/unmango/go-make/printer"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Printer", func() {
	Describe("rules", func() {
		DescribeTable("should print rule with",
			Entry("target",
				&ast.Rule{
					Colon: token.Pos(7),
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(1),
					}},
				},
				"target:\n",
			),
			Entry("multiple targets",
				&ast.Rule{Targets: []ast.Expr{
					&ast.Text{Value: "target", ValuePos: token.Pos(1)},
					&ast.Text{Value: "target2", ValuePos: token.Pos(8)},
				}},
				"target target2:\n",
			),
			Entry("variable reference target",
				&ast.Rule{
					Colon: token.Pos(10),
					Targets: []ast.Expr{&ast.VarRef{
						Dollar: token.Pos(1),
						Open:   token.LPAREN,
						Name:   "target",
						Close:  token.RPAREN,
					}},
				},
				"$(target):\n",
			),
			Entry("single character variable reference target",
				&ast.Rule{
					Colon: token.Pos(3),
					Targets: []ast.Expr{&ast.VarRef{
						Dollar: token.Pos(1),
						Open:   token.ILLEGAL,
						Name:   "t",
						Close:  token.ILLEGAL,
					}},
				},
				"$t:\n",
			),
			Entry("target with prereq",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(1),
					}},
					Colon: token.Pos(7),
					PreReqs: []ast.Expr{&ast.Text{
						Value:    "prereq",
						ValuePos: token.Pos(9),
					}},
				},
				"target: prereq\n",
			),
			Entry("target with order-only prereq",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(1),
					}},
					Colon: token.Pos(7),
					Pipe:  token.Pos(9),
					OrderPreReqs: []ast.Expr{&ast.Text{
						Value:    "prereq",
						ValuePos: token.Pos(11),
					}},
				},
				"target: | prereq\n",
			),
			Entry("target with prereq variable reference",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(1),
					}},
					Colon: token.Pos(7),
					PreReqs: []ast.Expr{&ast.VarRef{
						Dollar: token.Pos(9),
						Open:   token.LPAREN,
						Name:   "prereq",
						Close:  token.RPAREN,
					}},
				},
				"target: $(prereq)\n",
			),
			Entry("target, prereq, and recipe",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(1),
					}},
					Colon: token.Pos(7),
					PreReqs: []ast.Expr{&ast.Text{
						Value:    "prereq",
						ValuePos: token.Pos(9),
					}},
					Recipes: []*ast.Recipe{{
						Prefix:    token.TAB,
						PrefixPos: token.Pos(16),
						Text:      ast.Text{Value: "curl https://example.com"},
					}},
				},
				"target: prereq\n\tcurl https://example.com\n",
			),
			Entry("target with recipe",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "target"}},
					Recipes: []*ast.Recipe{{
						Prefix: token.TAB,
						Text:   ast.Text{Value: "curl https://example.com"},
					}},
				},
				"target:\n\tcurl https://example.com\n",
			),
			func(r *ast.Rule, expected string) {
				buf := &bytes.Buffer{}

				n, err := printer.Fprint(buf, r)

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal(expected))
				Expect(n).To(Equal(len(expected)))
			},
		)

		It("should write multiple rules", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, []ast.Obj{
				&ast.Rule{Targets: []ast.Expr{&ast.Text{Value: "target"}}},
				&ast.Rule{Targets: []ast.Expr{&ast.Text{Value: "target2"}}},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target:\ntarget2:\n"))
		})

		It("should ignore nil", func() {
			Expect(printer.Fprint(&bytes.Buffer{}, nil)).To(Equal(0))
		})

		It("should return write errors", func() {
			w := testing.NewErrAfterWriter(1)

			_, err := printer.Fprint(w, &ast.Rule{
				Targets: []ast.Expr{&ast.Text{Value: "foo"}},
				PreReqs: []ast.Expr{&ast.Text{Value: "bar"}},
				Recipes: []*ast.Recipe{{
					Prefix: token.TAB,
					Text:   ast.Text{Value: "baz"},
				}},
			})

			Expect(err).To(MatchError("write err: 1"))
		})
	})

	Describe("files", func() {
		It("should write a rule", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				Contents: []ast.Obj{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "target"}},
				}},
			})

			Expect(err).NotTo(HaveOccurred())
		})

		It("should write a comment", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				Contents: []ast.Obj{&ast.CommentGroup{List: []*ast.Comment{
					{Pound: token.Pos(1), Text: "comment text"},
				}}},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("# comment text\n"))
		})

		It("should write a comment group", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				Contents: []ast.Obj{&ast.CommentGroup{List: []*ast.Comment{
					{Pound: token.Pos(1), Text: "comment text"},
					{Pound: token.Pos(16), Text: "new line"},
				}}},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("# comment text\n# new line\n"))
		})

		It("should return errors found when writing a Makefile", func() {
			w := testing.ErrWriter("io error")

			_, err := printer.Fprint(w, &ast.File{
				Contents: []ast.Obj{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "target"}},
				}},
			})

			Expect(err).To(MatchError("io error"))
		})
	})

	Describe("expressions", func() {
		It("should write text", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.Text{Value: "foo"})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("foo"))
			Expect(n).To(Equal(3))
		})

		It("should write multiple text nodes", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, []ast.Expr{
				&ast.Text{Value: "foo", ValuePos: token.Pos(1)},
				&ast.Text{Value: "bar", ValuePos: token.Pos(5)},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("foo bar"))
			Expect(n).To(Equal(7))
		})
	})

	Describe("variables", func() {
		When("Value is empty", func() {
			It("should write a variable", func() {
				buf := &bytes.Buffer{}

				n, err := printer.Fprint(buf, &ast.Variable{
					Name:  &ast.Text{Value: "TEST"},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(5),
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST:=\n"))
				Expect(n).To(Equal(7))
			})

			It("should write a space-separated variable", func() {
				buf := &bytes.Buffer{}

				n, err := printer.Fprint(buf, &ast.Variable{
					Name:  &ast.Text{Value: "TEST", ValuePos: token.Pos(1)},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(6),
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST :=\n"))
				Expect(n).To(Equal(8))
			})
		})

		When("Value is defined", func() {
			It("should write a variable", func() {
				buf := &bytes.Buffer{}

				n, err := printer.Fprint(buf, &ast.Variable{
					Name:  &ast.Text{Value: "TEST", ValuePos: token.Pos(1)},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(5),
					Value: []ast.Expr{&ast.Text{
						Value:    "value",
						ValuePos: token.Pos(7),
					}},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST:=value\n"))
				Expect(n).To(Equal(12))
			})

			It("should write a space-separated variable", func() {
				buf := &bytes.Buffer{}

				n, err := printer.Fprint(buf, &ast.Variable{
					Name:  &ast.Text{Value: "TEST", ValuePos: token.Pos(1)},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(6),
					Value: []ast.Expr{&ast.Text{
						Value:    "value",
						ValuePos: token.Pos(9),
					}},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal("TEST := value\n"))
				Expect(n).To(Equal(14))
			})

			It("should return write errors", func() {
				w := testing.NewErrAfterWriter(1)

				_, err := printer.Fprint(w, &ast.Variable{
					Name:  &ast.Text{Value: "TEST", ValuePos: token.Pos(1)},
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(6),
					Value: []ast.Expr{&ast.Text{
						Value:    "value",
						ValuePos: token.Pos(9),
					}},
				})

				Expect(err).To(MatchError(ContainSubstring("write err:")))
			})
		})
	})

	When("a token.File is provided", func() {
		It("should work", func() {
			f := token.NewFileSet().AddFile("test", 1, 5)
			_, err := printer.Fprint(&bytes.Buffer{},
				&ast.Text{Value: "foo", ValuePos: token.Pos(1)},
				printer.WithFile(f),
			)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	When("the given node is not supported", func() {
		It("should return an error", func() {
			_, err := printer.Fprint(&bytes.Buffer{}, "blah")

			Expect(err).To(MatchError(`unsupported node: "blah"`))
		})
	})
})
