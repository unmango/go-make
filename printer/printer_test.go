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
			Entry("target and semicolon recipe",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(1),
					}},
					Colon: token.Pos(7),
					Recipes: []*ast.Recipe{{
						Prefix:    token.SEMI,
						PrefixPos: token.Pos(9),
						Text:      ast.Text{Value: " recipe", ValuePos: token.Pos(10)},
					}},
				},
				"target: ; recipe\n",
			),
			Entry("target, prereq, and semicolon recipe",
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
						Prefix:    token.SEMI,
						PrefixPos: token.Pos(16),
						Text:      ast.Text{Value: " recipe", ValuePos: token.Pos(17)},
					}},
				},
				"target: prereq ; recipe\n",
			),
			Entry("target, semicolon recipe, and tab recipe",
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(1),
					}},
					Colon: token.Pos(7),
					Recipes: []*ast.Recipe{
						{
							Prefix:    token.SEMI,
							PrefixPos: token.Pos(9),
							Text:      ast.Text{Value: " recipe", ValuePos: token.Pos(10)},
						},
						{
							Prefix:    token.TAB,
							PrefixPos: token.Pos(18),
							Text:      ast.Text{Value: "recipe2", ValuePos: token.Pos(19)},
						},
					},
				},
				"target: ; recipe\n\trecipe2\n",
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

		It("should write newline separated rules", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{Contents: []ast.Obj{
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(1),
					}},
					Colon: token.Pos(7),
				},
				&ast.Rule{Targets: []ast.Expr{&ast.Text{
					Value:    "target2",
					ValuePos: token.Pos(10),
				}}},
			}})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target:\n\ntarget2:\n"))
		})

		It("should write a comment", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				Contents: []ast.Obj{&ast.CommentGroup{List: []*ast.Comment{
					{Pound: token.Pos(1), Text: " comment text"},
				}}},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("# comment text\n"))
		})

		DescribeTable("should write the comment text verbatim after the pound",
			func(text, expected string) {
				buf := &bytes.Buffer{}

				_, err := printer.Fprint(buf, &ast.File{
					Contents: []ast.Obj{&ast.CommentGroup{List: []*ast.Comment{
						{Pound: token.Pos(1), Text: text},
					}}},
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal(expected))
			},
			Entry("no space", "comment text", "#comment text\n"),
			Entry("one space", " comment text", "# comment text\n"),
			Entry("two spaces", "  comment text", "#  comment text\n"),
			Entry("no text", "", "#\n"),
		)

		It("should write a comment group", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				Contents: []ast.Obj{&ast.CommentGroup{List: []*ast.Comment{
					{Pound: token.Pos(1), Text: " comment text"},
					{Pound: token.Pos(16), Text: " new line"},
				}}},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("# comment text\n# new line\n"))
		})

		It("should write newline separated comment groups", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				Contents: []ast.Obj{
					&ast.CommentGroup{List: []*ast.Comment{{
						Pound: token.Pos(1),
						Text:  " comment text",
					}}},
					&ast.CommentGroup{List: []*ast.Comment{{
						Pound: token.Pos(17),
						Text:  " other comment text",
					}}},
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("# comment text\n\n# other comment text\n"))
		})

		It("should write newline separated variables", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				Contents: []ast.Obj{
					&ast.Variable{
						Name:  &ast.Text{Value: "FOO", ValuePos: token.Pos(1)},
						Op:    token.SIMPLE_ASSIGN,
						OpPos: token.Pos(5),
					},
					&ast.Variable{
						Name:  &ast.Text{Value: "BAR", ValuePos: token.Pos(9)},
						Op:    token.SIMPLE_ASSIGN,
						OpPos: token.Pos(13),
					},
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("FOO :=\n\nBAR :=\n"))
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

	Describe("bad objects", func() {
		DescribeTable("should write the source text verbatim",
			Entry("assignment without spaces", "VAR=x"),
			Entry("shell assignment without spaces", "VAR!=x"),
			Entry("include directive", "include foo.mk"),
			Entry("export directive", "export VAR"),
			Entry("define directive", "define greeting"),
			Entry("unattached recipe line", "\tsecond"),
			func(text string) {
				buf := &bytes.Buffer{}

				_, err := printer.Fprint(buf, &ast.BadObj{
					From: token.Pos(1),
					To:   token.Pos(1 + len(text)),
					Text: text,
				})

				Expect(err).NotTo(HaveOccurred())
				Expect(buf.String()).To(Equal(text + "\n"))
			},
		)

		It("should write newline separated bad objects", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				Contents: []ast.Obj{
					&ast.BadObj{
						From: token.Pos(1),
						To:   token.Pos(15),
						Text: "include foo.mk",
					},
					&ast.BadObj{
						From: token.Pos(17),
						To:   token.Pos(31),
						Text: "include bar.mk",
					},
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("include foo.mk\n\ninclude bar.mk\n"))
		})

		It("should write a bad object after a rule", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				Contents: []ast.Obj{
					&ast.Rule{
						Targets: []ast.Expr{&ast.Text{
							Value:    "target",
							ValuePos: token.Pos(1),
						}},
						Colon: token.Pos(7),
					},
					&ast.BadObj{
						From: token.Pos(9),
						To:   token.Pos(23),
						Text: "include foo.mk",
					},
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target:\ninclude foo.mk\n"))
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

		It("should write apostrophe quoted text", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.QuotedExpr{
				Quote: token.APOS,
				Open:  token.Pos(1),
				Value: &ast.Text{Value: "foo", ValuePos: token.Pos(2)},
				Close: token.Pos(5),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("'foo'"))
			Expect(n).To(Equal(5))
		})

		It("should write quoted text", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.QuotedExpr{
				Quote: token.QUOTE,
				Open:  token.Pos(1),
				Value: &ast.Text{Value: "bar", ValuePos: token.Pos(2)},
				Close: token.Pos(5),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal(`"bar"`))
			Expect(n).To(Equal(5))
		})

		It("should write a function call", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.FuncCall{
				Dollar: token.Pos(1),
				Open:   token.LPAREN,
				Name:   &ast.Text{Value: "shell", ValuePos: token.Pos(3)},
				Args: []*ast.FuncArg{{
					From:  token.Pos(9),
					To:    token.Pos(12),
					Parts: []ast.Expr{&ast.Text{Value: "pwd", ValuePos: token.Pos(9)}},
				}},
				Close:    token.RPAREN,
				ClosePos: token.Pos(12),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("$(shell pwd)"))
			Expect(n).To(Equal(12))
		})

		It("should write a function call with no arguments", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.FuncCall{
				Dollar:   token.Pos(1),
				Open:     token.LBRACE,
				Name:     &ast.Text{Value: "shell", ValuePos: token.Pos(3)},
				Close:    token.RBRACE,
				ClosePos: token.Pos(8),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("${shell}"))
			Expect(n).To(Equal(8))
		})

		It("should write the whitespace inside an argument", func() {
			buf := &bytes.Buffer{}

			// "$(subst a, b,c)"
			n, err := printer.Fprint(buf, &ast.FuncCall{
				Dollar: token.Pos(1),
				Open:   token.LPAREN,
				Name:   &ast.Text{Value: "subst", ValuePos: token.Pos(3)},
				Args: []*ast.FuncArg{
					{From: token.Pos(9), To: token.Pos(10), Parts: []ast.Expr{
						&ast.Text{Value: "a", ValuePos: token.Pos(9)},
					}},
					{From: token.Pos(11), To: token.Pos(13), Parts: []ast.Expr{
						&ast.Text{Value: "b", ValuePos: token.Pos(12)},
					}},
					{From: token.Pos(14), To: token.Pos(15), Parts: []ast.Expr{
						&ast.Text{Value: "c", ValuePos: token.Pos(14)},
					}},
				},
				Commas:   []token.Pos{token.Pos(10), token.Pos(13)},
				Close:    token.RPAREN,
				ClosePos: token.Pos(15),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("$(subst a, b,c)"))
			Expect(n).To(Equal(15))
		})

		It("should write a nested function call", func() {
			buf := &bytes.Buffer{}
			inner := &ast.FuncCall{
				Dollar: token.Pos(7),
				Open:   token.LPAREN,
				Name:   &ast.Text{Value: "shell", ValuePos: token.Pos(9)},
				Args: []*ast.FuncArg{{
					From:  token.Pos(15),
					To:    token.Pos(18),
					Parts: []ast.Expr{&ast.Text{Value: "pwd", ValuePos: token.Pos(15)}},
				}},
				Close:    token.RPAREN,
				ClosePos: token.Pos(18),
			}

			n, err := printer.Fprint(buf, &ast.FuncCall{
				Dollar: token.Pos(1),
				Open:   token.LPAREN,
				Name:   &ast.Text{Value: "dir", ValuePos: token.Pos(3)},
				Args: []*ast.FuncArg{{
					From:  token.Pos(7),
					To:    token.Pos(19),
					Parts: []ast.Expr{inner},
				}},
				Close:    token.RPAREN,
				ClosePos: token.Pos(19),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("$(dir $(shell pwd))"))
			Expect(n).To(Equal(19))
		})

		It("should write a function argument", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.FuncArg{
				From: token.Pos(1),
				To:   token.Pos(6),
				Parts: []ast.Expr{
					&ast.Text{Value: "a", ValuePos: token.Pos(1)},
					&ast.Text{Value: "b", ValuePos: token.Pos(5)},
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("a   b"))
			Expect(n).To(Equal(5))
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

	Describe("directives", func() {
		It("should print an ifeq directive", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Arg1: &ast.Text{
					Value:    "foo",
					ValuePos: token.Pos(7),
				},
				Comma: token.Pos(10),
				Arg2: &ast.Text{
					Value:    "bar",
					ValuePos: token.Pos(12),
				},
				Close: token.Pos(15),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("ifeq (foo, bar)"))
			Expect(n).To(Equal(15))
		})

		It("should print an ifeq directive with quotes", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Arg1: &ast.QuotedExpr{
					Quote: token.APOS,
					Open:  token.Pos(6),
					Value: &ast.Text{
						Value:    "foo",
						ValuePos: token.Pos(7),
					},
					Close: token.Pos(10),
				},
				Arg2: &ast.QuotedExpr{
					Quote: token.QUOTE,
					Open:  token.Pos(12),
					Value: &ast.Text{
						Value:    "bar",
						ValuePos: token.Pos(13),
					},
					Close: token.Pos(16),
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("ifeq 'foo' \"bar\""))
			Expect(n).To(Equal(16))
		})

		It("should print an ifeq directive with an empty first argument", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Comma:  token.Pos(7),
				Arg2: &ast.Text{
					Value:    "bar",
					ValuePos: token.Pos(9),
				},
				Close: token.Pos(12),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("ifeq (, bar)"))
			Expect(n).To(Equal(12))
		})

		It("should print an ifeq directive with an empty second argument", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Arg1: &ast.Text{
					Value:    "foo",
					ValuePos: token.Pos(7),
				},
				Comma: token.Pos(10),
				Close: token.Pos(11),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("ifeq (foo,)"))
			Expect(n).To(Equal(11))
		})

		It("should print an ifeq directive with two empty arguments", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Comma:  token.Pos(7),
				Close:  token.Pos(8),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("ifeq (,)"))
			Expect(n).To(Equal(8))
		})

		It("should print an ifeq directive with empty quotes", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Arg1: &ast.QuotedExpr{
					Quote: token.APOS,
					Open:  token.Pos(6),
					Close: token.Pos(7),
				},
				Arg2: &ast.QuotedExpr{
					Quote: token.QUOTE,
					Open:  token.Pos(9),
					Value: &ast.Text{
						Value:    "bar",
						ValuePos: token.Pos(10),
					},
					Close: token.Pos(13),
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("ifeq '' \"bar\""))
			Expect(n).To(Equal(13))
		})

		It("should print an ifdef directive", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.IfdefDir{
				Tok:    token.IFDEF,
				TokPos: token.Pos(1),
				VarName: &ast.Text{
					Value:    "foo",
					ValuePos: token.Pos(7),
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("ifdef foo"))
			Expect(n).To(Equal(9))
		})

		It("should print an if block", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.IfBlock{
				Directive: &ast.IfdefDir{
					Tok:    token.IFDEF,
					TokPos: token.Pos(1),
					VarName: &ast.Text{
						Value:    "foo",
						ValuePos: token.Pos(7),
					},
				},
				Text: []ast.Obj{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "bar",
						ValuePos: token.Pos(11),
					}},
					Colon: token.Pos(14),
				}},
				Endif: token.Pos(16),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("ifdef foo\nbar:\nendif\n"))
			Expect(n).To(Equal(21))
		})

		It("should print an if block with an else", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.IfBlock{
				Directive: &ast.IfdefDir{
					Tok:    token.IFDEF,
					TokPos: token.Pos(1),
					VarName: &ast.Text{
						Value:    "foo",
						ValuePos: token.Pos(7),
					},
				},
				Text: []ast.Obj{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "bar",
						ValuePos: token.Pos(11),
					}},
					Colon: token.Pos(14),
				}},
				Else: []*ast.ElseBlock{{
					Else: token.Pos(16),
					Text: []ast.Obj{&ast.Rule{
						Targets: []ast.Expr{&ast.Text{
							Value:    "baz",
							ValuePos: token.Pos(21),
						}},
						Colon: token.Pos(24),
					}},
				}},
				Endif: token.Pos(26),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("ifdef foo\nbar:\nelse\nbaz:\nendif\n"))
			Expect(n).To(Equal(31))
		})

		It("should print an if block with an else if", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.IfBlock{
				Directive: &ast.IfdefDir{
					Tok:    token.IFDEF,
					TokPos: token.Pos(1),
					VarName: &ast.Text{
						Value:    "foo",
						ValuePos: token.Pos(7),
					},
				},
				Text: []ast.Obj{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "bar",
						ValuePos: token.Pos(11),
					}},
					Colon: token.Pos(14),
				}},
				Else: []*ast.ElseBlock{{
					Else: token.Pos(16),
					Condition: &ast.IfdefDir{
						Tok:    token.IFDEF,
						TokPos: token.Pos(21),
						VarName: &ast.Text{
							Value:    "baz",
							ValuePos: token.Pos(27),
						},
					},
					Text: []ast.Obj{&ast.Rule{
						Targets: []ast.Expr{&ast.Text{
							Value:    "bin",
							ValuePos: token.Pos(21),
						}},
						Colon: token.Pos(24),
					}},
				}},
				Endif: token.Pos(26),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("ifdef foo\nbar:\nelse ifdef baz\nbin:\nendif\n"))
			Expect(n).To(Equal(41))
		})
	})

	Describe("positions", func() {
		It("should not corrupt output when node positions run backwards", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.Rule{
				Targets: []ast.Expr{
					&ast.Text{Value: "one", ValuePos: token.Pos(1)},
					&ast.Text{Value: "two", ValuePos: token.Pos(5)},
					// Positioned before the targets above
					&ast.Text{Value: "three", ValuePos: token.Pos(2)},
				},
				Colon: token.Pos(13),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("one twothree:\n"))
		})

		It("should reset the column after a newline", func() {
			pos, err := printer.PrintPosition(&ast.Rule{
				Targets: []ast.Expr{&ast.Text{
					Value:    "target",
					ValuePos: token.Pos(1),
				}},
				Colon: token.Pos(7),
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(pos.Line).To(Equal(1))
			Expect(pos.Column).To(Equal(1))
			Expect(pos.Offset).To(Equal(8))
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

		It("should return an error for a top-level object", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, unknownObj{})

			Expect(err).To(MatchError(ContainSubstring("unsupported node:")))
			Expect(n).To(Equal(0))
			Expect(buf.String()).To(BeEmpty())
		})

		It("should return an error for a nested object", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.File{Contents: []ast.Obj{
				unknownObj{},
			}})

			Expect(err).To(MatchError(ContainSubstring("unsupported node:")))
			Expect(n).To(Equal(0))
			Expect(buf.String()).To(BeEmpty())
		})

		It("should return an error for a nested directive", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.File{Contents: []ast.Obj{
				unknownDir{},
			}})

			Expect(err).To(MatchError(ContainSubstring("unsupported node:")))
			Expect(n).To(Equal(0))
			Expect(buf.String()).To(BeEmpty())
		})

		It("should return an error for a nested expression", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.Rule{
				Targets: []ast.Expr{unknownExpr{}},
				Colon:   token.Pos(7),
			})

			Expect(err).To(MatchError(ContainSubstring("unsupported node:")))
			Expect(n).To(Equal(0))
			Expect(buf.String()).To(BeEmpty())
		})

		It("should return an error for a nested conditional directive", func() {
			buf := &bytes.Buffer{}

			n, err := printer.Fprint(buf, &ast.IfBlock{
				Directive: unknownIfDir{},
				Endif:     token.Pos(2),
			})

			Expect(err).To(MatchError(ContainSubstring("unsupported node:")))
			Expect(n).To(Equal(0))
			Expect(buf.String()).To(BeEmpty())
		})
	})

	Describe("line endings", func() {
		// The positions below are the offsets the parser would record for the
		// same source, so a two byte line ending leaves a two byte gap
		// between the end of one object and the start of the next.

		It("should default to LF when the file records no line ending", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{Contents: []ast.Obj{
				&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "a", ValuePos: token.Pos(1)}},
					Colon:   token.Pos(2),
				},
			}})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("a:\n"))
		})

		It("should write the line ending the file records", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				LineEnding: "\r\n",
				Contents: []ast.Obj{
					&ast.Rule{
						Targets: []ast.Expr{&ast.Text{Value: "a", ValuePos: token.Pos(1)}},
						Colon:   token.Pos(2),
					},
					&ast.Rule{
						Targets: []ast.Expr{&ast.Text{Value: "b", ValuePos: token.Pos(6)}},
						Colon:   token.Pos(7),
					},
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("a:\r\nb:\r\n"))
		})

		It("should reconstruct a blank line from a two byte gap", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				LineEnding: "\r\n",
				Contents: []ast.Obj{
					&ast.Rule{
						Targets: []ast.Expr{&ast.Text{Value: "a", ValuePos: token.Pos(1)}},
						Colon:   token.Pos(2),
					},
					&ast.Rule{
						Targets: []ast.Expr{&ast.Text{Value: "b", ValuePos: token.Pos(8)}},
						Colon:   token.Pos(9),
					},
				},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("a:\r\n\r\nb:\r\n"))
		})

		It("should write a recipe with the line ending the file records", func() {
			buf := &bytes.Buffer{}

			_, err := printer.Fprint(buf, &ast.File{
				LineEnding: "\r\n",
				Contents: []ast.Obj{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{Value: "a", ValuePos: token.Pos(1)}},
					Colon:   token.Pos(2),
					Recipes: []*ast.Recipe{{
						Prefix:    token.TAB,
						PrefixPos: token.Pos(5),
						Text:      ast.Text{Value: "echo hi", ValuePos: token.Pos(6)},
					}},
				}},
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("a:\r\n\techo hi\r\n"))
		})
	})
})

// ast.Obj, ast.Dir, ast.Expr, and ast.IfDir are sealed by unexported marker
// methods, so no type outside of package ast can implement them directly.
// Embedding the interface promotes the marker method and yields a node the
// printer has no case for. Pos is declared explicitly so the embedded nil
// interface is never called.

type unknownObj struct{ ast.Obj }

func (unknownObj) Pos() token.Pos { return token.Pos(1) }

type unknownDir struct{ ast.Dir }

func (unknownDir) Pos() token.Pos { return token.Pos(1) }

type unknownExpr struct{ ast.Expr }

func (unknownExpr) Pos() token.Pos { return token.Pos(1) }

type unknownIfDir struct{ ast.IfDir }

func (unknownIfDir) Pos() token.Pos { return token.Pos(1) }
