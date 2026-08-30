package ast_test

import (
	"bytes"
	"fmt"
	"strings"
	"testing/quick"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/parser"
	"github.com/unmango/go-make/printer"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Ast", func() {
	Describe("File", func() {
		When("the file contains no declarations", func() {
			It("should return the start of the file", func() {
				f := &ast.File{FileStart: token.Pos(69)}

				Expect(f.Pos()).To(Equal(token.Pos(69)))
			})

			It("should return the end of the file", func() {
				f := &ast.File{FileEnd: token.Pos(69)}

				Expect(f.End()).To(Equal(token.Pos(69)))
			})
		})

		When("the file contains declarations", func() {
			It("should return the first delcaration", func() {
				err := quick.Check(func(n int) bool {
					v := &ast.Variable{Name: &ast.Text{ValuePos: token.Pos(n)}}
					f := &ast.File{Contents: []ast.Obj{v}}

					return f.Pos() == v.Pos()
				}, nil)

				Expect(err).NotTo(HaveOccurred())
			})

			It("should return the end of the file", func() {
				err := quick.Check(func(n int) bool {
					v := &ast.Variable{Name: &ast.Text{ValuePos: token.Pos(n)}}
					f := &ast.File{Contents: []ast.Obj{v}}

					return f.End() == v.End()
				}, nil)

				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("BadObj", func() {
		It("should return the start of the bad object", func() {
			o := &ast.BadObj{From: token.Pos(69)}

			Expect(o.Pos()).To(Equal(token.Pos(69)))
			Expect(o.Pos()).To(Equal(o.From))
		})

		It("should return the end of the bad object", func() {
			o := &ast.BadObj{From: token.Pos(69), To: token.Pos(420)}

			Expect(o.End()).To(Equal(token.Pos(420)))
			Expect(o.End()).To(Equal(o.To))
		})

		It("should stringify", func() {
			o := &ast.BadObj{Text: "include foo.mk"}

			Expect(o.String()).To(Equal("include foo.mk"))
		})
	})

	Describe("CommentGroup", func() {
		It("should return the position of the first comment", func() {
			c := &ast.CommentGroup{[]*ast.Comment{{
				Pound: token.Pos(69),
			}}}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position after the last comment", func() {
			c := &ast.CommentGroup{[]*ast.Comment{
				{Pound: token.Pos(69), Text: "foo"},
				{Pound: token.Pos(420), Text: " Some comment text"},
			}}

			Expect(c.End()).To(Equal(token.Pos(439)))
		})
	})

	Describe("Comment", func() {
		It("should return the pound position", func() {
			c := &ast.Comment{Pound: token.Pos(69)}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
			Expect(c.Pos()).To(Equal(c.Pound))
		})

		It("should return the position after the comment text", func() {
			c := &ast.Comment{
				Pound: token.Pos(420),
				Text:  " Some comment text",
			}

			// '#' + len(" Some comment text")
			Expect(c.End()).To(Equal(token.Pos(439)))
		})

		It("should return the position after a comment with no leading space", func() {
			c := &ast.Comment{
				Pound: token.Pos(420),
				Text:  "Some comment text",
			}

			// '#' + len("Some comment text")
			Expect(c.End()).To(Equal(token.Pos(438)))
		})

		It("should return the position after the pound when there is no text", func() {
			c := &ast.Comment{Pound: token.Pos(420)}

			Expect(c.End()).To(Equal(token.Pos(421)))
		})
	})

	Describe("Rule", func() {
		It("should return the position of the first target", func() {
			c := &ast.Rule{Targets: []ast.Expr{
				&ast.Text{ValuePos: token.Pos(69)},
			}}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position after the colon", func() {
			r := &ast.Rule{
				Targets: []ast.Expr{&ast.Text{Value: "test"}},
				Colon:   5,
			}

			Expect(r.End()).To(Equal(token.Pos(6)))
		})

		It("should return the position after the final pre-requisite", func() {
			p := &ast.Text{Value: "test", ValuePos: 3}
			r := &ast.Rule{PreReqs: []ast.Expr{p}}

			Expect(r.End()).To(Equal(p.End()))
		})

		It("should return the position after the final order-only pre-requisite", func() {
			p := &ast.Text{Value: "test", ValuePos: 5}
			r := &ast.Rule{OrderPreReqs: []ast.Expr{p}}

			Expect(r.End()).To(Equal(p.End()))
		})

		It("should return the position after the final recipe", func() {
			r := &ast.Recipe{
				PrefixPos: token.Pos(420),
				Text:      ast.Text{Value: "some text"},
			}
			c := &ast.Rule{Recipes: []*ast.Recipe{r}}

			Expect(c.End()).To(Equal(r.End()))
		})
	})

	Describe("Text", func() {
		It("should return the position of the identifier", func() {
			c := &ast.Text{ValuePos: token.Pos(69)}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position after the identifier", func() {
			c := &ast.Text{
				ValuePos: token.Pos(420),
				Value:    "bar",
			}

			Expect(c.End()).To(Equal(token.Pos(423)))
		})

		It("should stringify", func() {
			c := &ast.Text{Value: "foo"}

			Expect(c.String()).To(Equal("foo"))
		})
	})

	DescribeTableSubtree("QuotedExpr",
		Entry(nil, token.QUOTE),
		Entry(nil, token.APOS),
		func(quote token.Token) {
			It("should return the position of the opening quote", func() {
				c := &ast.QuotedExpr{Quote: quote, Open: token.Pos(69)}

				Expect(c.Pos()).To(Equal(token.Pos(69)))
			})

			It("should return the position after the closing quote", func() {
				c := &ast.QuotedExpr{Quote: quote, Close: token.Pos(423)}

				Expect(c.End()).To(Equal(token.Pos(424)))
			})

			It("should stringify", func() {
				c := &ast.QuotedExpr{
					Quote: quote,
					Value: &ast.Text{Value: "foo"},
				}

				Expect(c.String()).To(Equal(fmt.Sprint(quote, "foo", quote)))
			})

			It("should stringify an empty value", func() {
				c := &ast.QuotedExpr{Quote: quote}

				Expect(c.String()).To(Equal(quote.String() + quote.String()))
			})
		},
	)

	Describe("QuotedExpr", func() {
		It("should stringify an absent quote", func() {
			c := &ast.QuotedExpr{Value: &ast.Text{Value: "foo"}}

			Expect(c.String()).To(Equal("foo"))
		})

		It("should stringify an absent quote and an empty value", func() {
			c := &ast.QuotedExpr{}

			Expect(c.String()).To(BeEmpty())
		})

		DescribeTable("should stringify what the printer writes",
			Entry("quotes", &ast.QuotedExpr{
				Quote: token.QUOTE,
				Open:  token.Pos(1),
				Value: &ast.Text{Value: "foo", ValuePos: token.Pos(2)},
				Close: token.Pos(5),
			}, `"foo"`),
			Entry("apostrophes", &ast.QuotedExpr{
				Quote: token.APOS,
				Open:  token.Pos(1),
				Value: &ast.Text{Value: "foo", ValuePos: token.Pos(2)},
				Close: token.Pos(5),
			}, "'foo'"),
			Entry("an empty value", &ast.QuotedExpr{
				Quote: token.QUOTE,
				Open:  token.Pos(1),
				Close: token.Pos(2),
			}, `""`),
			Entry("an absent quote", &ast.QuotedExpr{
				Value: &ast.Text{Value: "foo", ValuePos: token.Pos(1)},
			}, "foo"),
			Entry("an absent quote and an empty value", &ast.QuotedExpr{}, ""),
			Entry("a variable reference", &ast.QuotedExpr{
				Quote: token.QUOTE,
				Open:  token.Pos(1),
				Value: &ast.VarRef{
					Dollar: token.Pos(2),
					Open:   token.LPAREN,
					Name:   "FOO",
					Close:  token.RPAREN,
				},
				Close: token.Pos(8),
			}, `"$(FOO)"`),
			func(c *ast.QuotedExpr, expected string) {
				buf := &bytes.Buffer{}
				_, err := printer.Fprint(buf, c)

				Expect(err).NotTo(HaveOccurred())
				Expect(c.String()).To(Equal(expected))
				Expect(c.String()).To(Equal(buf.String()))
			},
		)
	})

	Describe("VarRef", func() {
		It("should return the position of the dollar sign", func() {
			err := quick.Check(func(p int) bool {
				c := &ast.VarRef{Dollar: token.Pos(p)}
				return c.Pos() == token.Pos(p)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the closing token", func() {
			c := &ast.VarRef{
				Dollar: token.Pos(420),
				Open:   token.LPAREN,
				Name:   "bar",
				Close:  token.RPAREN,
			}

			// '$' + '(' + len("bar") + ')'
			Expect(c.End()).To(Equal(token.Pos(426)))
		})

		It("should return the position after the character", func() {
			c := &ast.VarRef{
				Dollar: token.Pos(420),
				Open:   token.ILLEGAL,
				Name:   "b",
				Close:  token.ILLEGAL,
			}

			// '$' + len("b")
			Expect(c.End()).To(Equal(token.Pos(422)))
		})

		It("should return the position after a delimited single character", func() {
			c := &ast.VarRef{
				Dollar: token.Pos(420),
				Open:   token.LBRACE,
				Name:   "b",
				Close:  token.RBRACE,
			}

			// '$' + '{' + len("b") + '}'
			Expect(c.End()).To(Equal(token.Pos(424)))
		})

		DescribeTable("should stringify",
			Entry("parens", token.LPAREN, "FOO", token.RPAREN, "$(FOO)"),
			Entry("braces", token.LBRACE, "FOO", token.RBRACE, "${FOO}"),
			Entry("a delimited single character", token.LPAREN, "F", token.RPAREN, "$(F)"),
			Entry("a brace delimited single character", token.LBRACE, "F", token.RBRACE, "${F}"),
			Entry("a single character", token.ILLEGAL, "F", token.ILLEGAL, "$F"),
			Entry("an undelimited name", token.ILLEGAL, "FOO", token.ILLEGAL, "$FOO"),
			func(open token.Token, name string, closed token.Token, expected string) {
				c := &ast.VarRef{
					Dollar: token.Pos(1),
					Open:   open,
					Name:   name,
					Close:  closed,
				}

				Expect(c.String()).To(Equal(expected))
			},
		)

		DescribeTable("should stringify what the printer writes",
			Entry("parens", token.LPAREN, "FOO", token.RPAREN),
			Entry("braces", token.LBRACE, "FOO", token.RBRACE),
			Entry("a delimited single character", token.LPAREN, "F", token.RPAREN),
			Entry("a brace delimited single character", token.LBRACE, "F", token.RBRACE),
			Entry("a single character", token.ILLEGAL, "F", token.ILLEGAL),
			Entry("an undelimited name", token.ILLEGAL, "FOO", token.ILLEGAL),
			func(open token.Token, name string, closed token.Token) {
				c := &ast.VarRef{
					Dollar: token.Pos(1),
					Open:   open,
					Name:   name,
					Close:  closed,
				}

				buf := &bytes.Buffer{}
				_, err := printer.Fprint(buf, c)

				Expect(err).NotTo(HaveOccurred())
				Expect(c.String()).To(Equal(buf.String()))
			},
		)
	})

	Describe("FuncCall", func() {
		It("should return the position of the dollar sign", func() {
			err := quick.Check(func(p int) bool {
				c := &ast.FuncCall{Dollar: token.Pos(p)}
				return c.Pos() == token.Pos(p)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the closing delimiter", func() {
			c := &ast.FuncCall{
				Dollar:   token.Pos(420),
				Open:     token.LPAREN,
				Name:     &ast.Text{ValuePos: token.Pos(422), Value: "shell"},
				Close:    token.RPAREN,
				ClosePos: token.Pos(431),
			}

			// ')' + len(")")
			Expect(c.End()).To(Equal(token.Pos(432)))
		})

		It("should return the position of the name", func() {
			c := &ast.FuncCall{Dollar: token.Pos(420), Open: token.LPAREN}

			// '$' + '('
			Expect(c.NamePos()).To(Equal(token.Pos(422)))
		})

		DescribeTable("should stringify",
			Entry("a call with no arguments", "A := $(shell)\n", "$(shell)"),
			Entry("a call with one argument", "A := $(shell pwd)\n", "$(shell pwd)"),
			Entry("a brace delimited call", "A := ${shell pwd}\n", "${shell pwd}"),
			Entry("a call with several arguments", "A := $(subst a,b,c)\n", "$(subst a,b,c)"),
			Entry("an argument with leading space", "A := $(subst a, b,c)\n", "$(subst a, b,c)"),
			Entry("a nested call", "A := $(dir $(shell pwd))\n", "$(dir $(shell pwd))"),
			func(input, expected string) {
				p := parser.New(strings.NewReader(input), nil)

				f, err := p.ParseFile()

				Expect(err).NotTo(HaveOccurred())
				v, ok := f.Contents[0].(*ast.Variable)
				Expect(ok).To(BeTrue(), "expected a *ast.Variable, got %T", f.Contents[0])
				c, ok := v.Value[0].(*ast.FuncCall)
				Expect(ok).To(BeTrue(), "expected a *ast.FuncCall, got %T", v.Value[0])
				Expect(c.String()).To(Equal(expected))
			},
		)

		It("should stringify what the printer writes", func() {
			// "$(subst a, b,c)" laid out from the first position in the file.
			c := &ast.FuncCall{
				Dollar: token.Pos(1),
				Open:   token.LPAREN,
				Name:   &ast.Text{ValuePos: token.Pos(3), Value: "subst"},
				Args: []*ast.FuncArg{
					{From: token.Pos(9), To: token.Pos(10), Parts: []ast.Expr{
						&ast.Text{ValuePos: token.Pos(9), Value: "a"},
					}},
					{From: token.Pos(11), To: token.Pos(13), Parts: []ast.Expr{
						&ast.Text{ValuePos: token.Pos(12), Value: "b"},
					}},
					{From: token.Pos(14), To: token.Pos(15), Parts: []ast.Expr{
						&ast.Text{ValuePos: token.Pos(14), Value: "c"},
					}},
				},
				Commas:   []token.Pos{token.Pos(10), token.Pos(13)},
				Close:    token.RPAREN,
				ClosePos: token.Pos(15),
			}

			buf := &bytes.Buffer{}
			_, err := printer.Fprint(buf, c)

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("$(subst a, b,c)"))
			Expect(c.String()).To(Equal(buf.String()))
		})
	})

	Describe("FuncArg", func() {
		It("should return the start of the argument", func() {
			err := quick.Check(func(p int) bool {
				a := &ast.FuncArg{From: token.Pos(p)}
				return a.Pos() == token.Pos(p)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the end of the argument", func() {
			err := quick.Check(func(p int) bool {
				a := &ast.FuncArg{To: token.Pos(p)}
				return a.End() == token.Pos(p)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should stringify the whitespace it contains", func() {
			a := &ast.FuncArg{
				From: token.Pos(1),
				To:   token.Pos(6),
				Parts: []ast.Expr{
					&ast.Text{ValuePos: token.Pos(2), Value: "a"},
					&ast.Text{ValuePos: token.Pos(5), Value: "b"},
				},
			}

			Expect(a.String()).To(Equal(" a  b"))
		})
	})

	Describe("Recipe", func() {
		It("should return the position of the tab", func() {
			c := &ast.Recipe{
				PrefixPos: token.Pos(420),
			}

			Expect(c.Pos()).To(Equal(token.Pos(420)))
		})

		It("should return the position after the text", func() {
			c := &ast.Recipe{
				PrefixPos: token.Pos(420),
				Prefix:    token.TAB,
				Text:      ast.Text{Value: "foo"},
			}

			// '\t' + len("foo")
			Expect(c.End()).To(Equal(token.Pos(424)))
		})
	})

	Describe("Variable", func() {
		It("should return the position of the name", func() {
			err := quick.Check(func(n int) bool {
				v := &ast.Variable{Name: &ast.Text{ValuePos: token.Pos(n)}}

				return v.Pos() == token.Pos(n)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the value", func() {
			err := quick.Check(func(n int) bool {
				v := &ast.Variable{Value: []ast.Expr{&ast.Text{
					ValuePos: token.Pos(n),
					Value:    "foo",
				}}}

				return v.End() == token.Pos(n+3)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		When("there is no value", func() {
			DescribeTable("should return the position after the operator",
				Entry(":=", token.SIMPLE_ASSIGN, 2),
				Entry("=", token.RECURSIVE_ASSIGN, 1),
				Entry("::=", token.POSIX_ASSIGN, 3),
				Entry(":::=", token.IMMEDIATE_ASSIGN, 4),
				Entry("?=", token.IFNDEF_ASSIGN, 2),
				Entry("!=", token.SHELL_ASSIGN, 2),
				Entry("+=", token.APPEND_ASSIGN, 2),
				func(tok token.Token, l int) {
					err := quick.Check(func(n int) bool {
						v := &ast.Variable{
							Op:    tok,
							OpPos: token.Pos(n),
						}

						return v.End() == token.Pos(n+l)
					}, nil)

					Expect(err).NotTo(HaveOccurred())
				},
			)
		})
	})

	Describe("IfeqDir", func() {
		It("should return the position of the directive token", func() {
			err := quick.Check(func(n int) bool {
				d := &ast.IfeqDir{
					Tok:    token.IFEQ,
					TokPos: token.Pos(n),
				}

				return d.Pos() == token.Pos(n)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the closing parethesis", func() {
			err := quick.Check(func(n int) bool {
				d := &ast.IfeqDir{Close: token.Pos(n)}

				return d.End() == token.Pos(n+1)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the second arg", func() {
			err := quick.Check(func(n int) bool {
				d := &ast.IfeqDir{Arg2: &ast.Text{ValuePos: token.Pos(n)}}

				return d.End() == token.Pos(n)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the first arg when the second is empty", func() {
			err := quick.Check(func(n int) bool {
				d := &ast.IfeqDir{Arg1: &ast.Text{ValuePos: token.Pos(n)}}

				return d.End() == token.Pos(n)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the directive token when both args are empty", func() {
			d := &ast.IfeqDir{Tok: token.IFEQ, TokPos: token.Pos(1)}

			Expect(d.End()).To(Equal(token.Pos(5)))
		})
	})

	Describe("IfdefDir", func() {
		It("should return the position of the directive token", func() {
			err := quick.Check(func(n int) bool {
				d := &ast.IfdefDir{
					Tok:    token.IFDEF,
					TokPos: token.Pos(n),
				}

				return d.Pos() == token.Pos(n)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the arg", func() {
			err := quick.Check(func(n int) bool {
				d := &ast.IfdefDir{VarName: &ast.Text{ValuePos: token.Pos(n)}}

				return d.End() == token.Pos(n)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("ElseBlock", func() {
		It("should return the position of the directive token", func() {
			err := quick.Check(func(n int) bool {
				d := &ast.ElseBlock{Else: token.Pos(n)}

				return d.Pos() == token.Pos(n)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the directive token", func() {
			err := quick.Check(func(n int) bool {
				d := &ast.ElseBlock{Else: token.Pos(n)}

				return d.End() == token.Pos(n+4)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the condition", func() {
			err := quick.Check(func(n int) bool {
				ifeq := &ast.IfeqDir{Close: token.Pos(n)}
				d := &ast.ElseBlock{Condition: ifeq}

				return d.End() == ifeq.End()
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the text", func() {
			err := quick.Check(func(n int) bool {
				d := &ast.ElseBlock{Text: []ast.Obj{&ast.Variable{
					Op:    token.SIMPLE_ASSIGN,
					OpPos: token.Pos(n),
				}}}

				return d.End() == token.Pos(n+2)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})
	})

	Describe("IfBlock", func() {
		It("should return the position of the directive token", func() {
			err := quick.Check(func(n int) bool {
				d := &ast.IfBlock{
					Directive: &ast.IfeqDir{TokPos: token.Pos(n)},
				}

				return d.Pos() == token.Pos(n)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})

		It("should return the position after the endif directive", func() {
			err := quick.Check(func(n int) bool {
				d := &ast.IfBlock{Endif: token.Pos(n)}

				return d.End() == token.Pos(n+5)
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})
	})
})

var _ = Describe("End", func() {
	// assertSpans parses src, prints the resulting file back out, and asserts
	// that the node selected by pick covers exactly the bytes of snippet.
	// Positions are 1-based, so End is the position immediately after the last
	// character the printer writes for the node.
	assertSpans := func(src, snippet string, pick func(*ast.File) ast.Node) {
		GinkgoHelper()

		f, err := parser.New(bytes.NewBufferString(src), nil).ParseFile()
		Expect(err).NotTo(HaveOccurred())

		buf := &bytes.Buffer{}
		_, err = printer.Fprint(buf, f)
		Expect(err).NotTo(HaveOccurred())
		Expect(buf.String()).To(Equal(src), "source did not round-trip")

		offset := strings.Index(src, snippet)
		Expect(offset).To(BeNumerically(">=", 0), "snippet is not present in the source")

		node := pick(f)
		Expect(node.Pos()).To(Equal(token.Pos(offset + 1)))
		Expect(node.End()).To(Equal(token.Pos(offset + 1 + len(snippet))))
		Expect(src[node.Pos()-1 : node.End()-1]).To(Equal(snippet))
	}

	It("should span the comment including the pound", func() {
		assertSpans("# a comment\n", "# a comment", func(f *ast.File) ast.Node {
			return f.Contents[0].(*ast.CommentGroup).List[0]
		})
	})

	It("should span the comment with no space after the pound", func() {
		assertSpans("#a comment\n", "#a comment", func(f *ast.File) ast.Node {
			return f.Contents[0].(*ast.CommentGroup).List[0]
		})
	})

	It("should span the comment with extra space after the pound", func() {
		assertSpans("#  a comment\n", "#  a comment", func(f *ast.File) ast.Node {
			return f.Contents[0].(*ast.CommentGroup).List[0]
		})
	})

	It("should span the comment with no text", func() {
		assertSpans("#\n", "#", func(f *ast.File) ast.Node {
			return f.Contents[0].(*ast.CommentGroup).List[0]
		})
	})

	It("should span the variable reference including the closing token", func() {
		assertSpans("target: $(FOO)\n", "$(FOO)", func(f *ast.File) ast.Node {
			return f.Contents[0].(*ast.Rule).PreReqs[0]
		})
	})

	It("should span the single character variable reference", func() {
		assertSpans("target: $b\n", "$b", func(f *ast.File) ast.Node {
			return f.Contents[0].(*ast.Rule).PreReqs[0]
		})
	})

	It("should span the quoted expression including the closing quote", func() {
		assertSpans("ifeq \"a\" \"bcd\"\ntarget:\nendif\n", `"bcd"`, func(f *ast.File) ast.Node {
			return f.Contents[0].(*ast.IfBlock).Directive.(*ast.IfeqDir).Arg2
		})
	})

	It("should span the recipe including the prefix", func() {
		assertSpans("target: dep\n\techo hi\n", "\techo hi", func(f *ast.File) ast.Node {
			return f.Contents[0].(*ast.Rule).Recipes[0]
		})
	})
})
