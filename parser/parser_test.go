package parser_test

import (
	"bytes"
	"fmt"
	gotoken "go/token"
	"io"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/parser"
	"github.com/unmango/go-make/printer"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Parser", func() {
	var file *token.File

	BeforeEach(func() {
		file = gotoken.NewFileSet().AddFile("test", 1, math.MaxInt-2)
	})

	It("should Parse a comment", func() {
		buf := bytes.NewBufferString("# comment text")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.CommentGroup{
			List: []*ast.Comment{{Pound: token.Pos(1), Text: " comment text"}},
		}))
	})

	DescribeTable("should Parse a comment preserving the text after the pound",
		func(input, text string) {
			p := parser.New(bytes.NewBufferString(input), file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.CommentGroup{
				List: []*ast.Comment{{Pound: token.Pos(1), Text: text}},
			}))
		},
		Entry("no space", "#comment text", "comment text"),
		Entry("one space", "# comment text", " comment text"),
		Entry("two spaces", "#  comment text", "  comment text"),
		Entry("no text", "#", ""),
		Entry("no text and a newline", "#\n", ""),
	)

	It("should Parse a comment group", func() {
		buf := bytes.NewBufferString("# comment text\n# more text on this line")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.CommentGroup{
			List: []*ast.Comment{
				{Pound: token.Pos(1), Text: " comment text"},
				{Pound: token.Pos(16), Text: " more text on this line"},
			},
		}))
	})

	It("should Parse multiple comment groups", func() {
		buf := bytes.NewBufferString("# comment text\n\n# new comment group")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(
			&ast.CommentGroup{List: []*ast.Comment{
				{Pound: token.Pos(1), Text: " comment text"},
			}},
			&ast.CommentGroup{List: []*ast.Comment{
				{Pound: token.Pos(17), Text: " new comment group"},
			}},
		))
	})

	It("should Parse a target", func() {
		buf := bytes.NewBufferString("target:")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	DescribeTable("should Parse a target as a variable reference",
		Entry(nil, "${foo}:", "foo", token.LBRACE, token.RBRACE),
		Entry(nil, "$(foo):", "foo", token.LPAREN, token.RPAREN),
		func(input, name string, open, close token.Token) {
			buf := bytes.NewBufferString(input)
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.Rule{
				Colon: token.Pos(7),
				Targets: []ast.Expr{&ast.VarRef{
					Dollar: token.Pos(1),
					Open:   open,
					Name:   name,
					Close:  close,
				}},
				PreReqs:      []ast.Expr{},
				OrderPreReqs: []ast.Expr{},
				Recipes:      []ast.RecipeObj{},
			}))
		},
	)

	It("should Parse a target as a single character variable reference", func() {
		buf := bytes.NewBufferString("$f:")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(3),
			Targets: []ast.Expr{&ast.VarRef{
				Dollar: token.Pos(1),
				Open:   token.ILLEGAL,
				Name:   "f",
				Close:  token.ILLEGAL,
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse a target with a single character variable reference and extra text", func() {
		buf := bytes.NewBufferString("$foo:")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(5),
			Targets: []ast.Expr{&ast.JuxtaposedExpr{Parts: []ast.Expr{
				&ast.VarRef{
					Dollar: token.Pos(1),
					Open:   token.ILLEGAL,
					Name:   "f",
					Close:  token.ILLEGAL,
				},
				&ast.Text{Value: "oo", ValuePos: token.Pos(3)},
			}}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse an escaped dollar sign in a target", func() {
		buf := bytes.NewBufferString("$$V:")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(4),
			Targets: []ast.Expr{&ast.JuxtaposedExpr{Parts: []ast.Expr{
				&ast.Text{Value: "$$", ValuePos: token.Pos(1)},
				&ast.Text{Value: "V", ValuePos: token.Pos(3)},
			}}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse an escaped dollar sign in a prereq", func() {
		buf := bytes.NewBufferString("target: $$V")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{&ast.JuxtaposedExpr{Parts: []ast.Expr{
				&ast.Text{Value: "$$", ValuePos: token.Pos(9)},
				&ast.Text{Value: "V", ValuePos: token.Pos(11)},
			}}},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse an escaped dollar sign in a variable value", func() {
		buf := bytes.NewBufferString("VAR := $$V")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Variable{
			Name: &ast.Text{
				Value:    "VAR",
				ValuePos: token.Pos(1),
			},
			Op:    token.SIMPLE_ASSIGN,
			OpPos: token.Pos(5),
			Value: []ast.Expr{&ast.JuxtaposedExpr{Parts: []ast.Expr{
				&ast.Text{Value: "$$", ValuePos: token.Pos(8)},
				&ast.Text{Value: "V", ValuePos: token.Pos(10)},
			}}},
		}))
	})

	It("should Parse a lone escaped dollar sign", func() {
		buf := bytes.NewBufferString("target: $$")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{&ast.Text{
				Value:    "$$",
				ValuePos: token.Pos(9),
			}},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse an escaped dollar sign followed by a parenthesized name", func() {
		buf := bytes.NewBufferString("target: $$(notdir x)")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{
				&ast.JuxtaposedExpr{Parts: []ast.Expr{
					&ast.Text{Value: "$$", ValuePos: token.Pos(9)},
					&ast.Text{Value: "(", ValuePos: token.Pos(11)},
					&ast.Text{Value: "notdir", ValuePos: token.Pos(12)},
				}},
				&ast.JuxtaposedExpr{Parts: []ast.Expr{
					&ast.Text{Value: "x", ValuePos: token.Pos(19)},
					&ast.Text{Value: ")", ValuePos: token.Pos(20)},
				}},
			},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse a reference joined to the text after it", func() {
		buf := bytes.NewBufferString("target: $(FOO)bar")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{&ast.JuxtaposedExpr{Parts: []ast.Expr{
				&ast.VarRef{
					Dollar: token.Pos(9),
					Open:   token.LPAREN,
					Name:   "FOO",
					Close:  token.RPAREN,
				},
				&ast.Text{Value: "bar", ValuePos: token.Pos(15)},
			}}},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse two references written next to each other", func() {
		buf := bytes.NewBufferString("target: $(FOO)${BAR}")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{&ast.JuxtaposedExpr{Parts: []ast.Expr{
				&ast.VarRef{
					Dollar: token.Pos(9),
					Open:   token.LPAREN,
					Name:   "FOO",
					Close:  token.RPAREN,
				},
				&ast.VarRef{
					Dollar: token.Pos(15),
					Open:   token.LBRACE,
					Name:   "BAR",
					Close:  token.RBRACE,
				},
			}}},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse a call joined to the reference after it", func() {
		buf := bytes.NewBufferString("OUT := $(notdir a)$(FOO)")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Variable{
			Name:  &ast.Text{Value: "OUT", ValuePos: token.Pos(1)},
			Op:    token.SIMPLE_ASSIGN,
			OpPos: token.Pos(5),
			Value: []ast.Expr{&ast.JuxtaposedExpr{Parts: []ast.Expr{
				&ast.FuncCall{
					Dollar: token.Pos(8),
					Open:   token.LPAREN,
					Name:   &ast.Text{Value: "notdir", ValuePos: token.Pos(10)},
					Args: []*ast.FuncArg{{
						From:  token.Pos(17),
						To:    token.Pos(18),
						Parts: []ast.Expr{&ast.Text{Value: "a", ValuePos: token.Pos(17)}},
					}},
					Close:    token.RPAREN,
					ClosePos: token.Pos(18),
				},
				&ast.VarRef{
					Dollar: token.Pos(19),
					Open:   token.LPAREN,
					Name:   "FOO",
					Close:  token.RPAREN,
				},
			}}},
		}))
	})

	It("should Parse a reference joined to a comma", func() {
		buf := bytes.NewBufferString("LIST := $(FOO),$(BAR)")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Variable{
			Name:  &ast.Text{Value: "LIST", ValuePos: token.Pos(1)},
			Op:    token.SIMPLE_ASSIGN,
			OpPos: token.Pos(6),
			Value: []ast.Expr{&ast.JuxtaposedExpr{Parts: []ast.Expr{
				&ast.VarRef{
					Dollar: token.Pos(9),
					Open:   token.LPAREN,
					Name:   "FOO",
					Close:  token.RPAREN,
				},
				&ast.Text{Value: ",", ValuePos: token.Pos(15)},
				&ast.VarRef{
					Dollar: token.Pos(16),
					Open:   token.LPAREN,
					Name:   "BAR",
					Close:  token.RPAREN,
				},
			}}},
		}))
	})

	It("should not join an expression across a blank", func() {
		buf := bytes.NewBufferString("target: $(FOO) bar")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{
				&ast.VarRef{
					Dollar: token.Pos(9),
					Open:   token.LPAREN,
					Name:   "FOO",
					Close:  token.RPAREN,
				},
				&ast.Text{Value: "bar", ValuePos: token.Pos(16)},
			},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should end an ifeq argument at the comma it is joined to", func() {
		buf := bytes.NewBufferString("ifeq ($(FOO)x,y)\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Arg1: &ast.JuxtaposedExpr{Parts: []ast.Expr{
					&ast.VarRef{
						Dollar: token.Pos(7),
						Open:   token.LPAREN,
						Name:   "FOO",
						Close:  token.RPAREN,
					},
					&ast.Text{Value: "x", ValuePos: token.Pos(13)},
				}},
				Comma: token.Pos(14),
				Arg2:  &ast.Text{Value: "y", ValuePos: token.Pos(15)},
				Close: token.Pos(16),
			},
			Endif: token.Pos(18),
		}))
	})

	// A '$' opens an expansion only when an expansion follows it immediately.
	// Written anywhere else the character stands for itself, so the parser
	// records it as text rather than reporting an error it cannot recover the
	// source from.
	DescribeTable("should parse a dollar sign that opens nothing as text",
		Entry("at the end of the input", "target: $", "$"),
		Entry("followed by a closing paren", "target: $)", "$)"),
		Entry("followed by a blank", "target: $ x", "$"),
		Entry("written against the text before it", "target: a$", "a$"),
		func(input, prereq string) {
			buf := bytes.NewBufferString(input)
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(HaveLen(1))
			rule, ok := f.Contents[0].(*ast.Rule)
			Expect(ok).To(BeTrue(), "expected a *ast.Rule, got %T", f.Contents[0])
			Expect(rule.PreReqs).NotTo(BeEmpty())
			Expect(rule.PreReqs[0].(fmt.Stringer).String()).To(Equal(prereq))
		},
	)

	DescribeTable("should error when variable reference has no closing token",
		Entry(nil, "${foo:"),
		Entry(nil, "$(foo:"),
		func(input string) {
			buf := bytes.NewBufferString(input)
			p := parser.New(buf, file)

			_, err := p.ParseFile()

			Expect(err).To(MatchError("test:1:6: expected one of ')', '}', found ':'"))
		},
	)

	It("should Parse a rule with multiple targets", func() {
		buf := bytes.NewBufferString("target target2:")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(15),
			Targets: []ast.Expr{
				&ast.Text{Value: "target", ValuePos: token.Pos(1)},
				&ast.Text{Value: "target2", ValuePos: token.Pos(8)},
			},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse a target with a prereq", func() {
		buf := bytes.NewBufferString("target: prereq")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{&ast.Text{
				Value:    "prereq",
				ValuePos: token.Pos(9),
			}},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse a target with multiple prereqs", func() {
		buf := bytes.NewBufferString("target: prereq prereq2")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{
				&ast.Text{Value: "prereq", ValuePos: token.Pos(9)},
				&ast.Text{Value: "prereq2", ValuePos: token.Pos(16)},
			},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse a target with a prereq variable reference", func() {
		buf := bytes.NewBufferString("target: ${FOO}")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{&ast.VarRef{
				Dollar: token.Pos(9),
				Open:   token.LBRACE,
				Name:   "FOO",
				Close:  token.RBRACE,
			}},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []ast.RecipeObj{},
		}))
	})

	It("should Parse a target with an order-only prereq", func() {
		buf := bytes.NewBufferString("target: | prereq")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			Colon:   token.Pos(7),
			Pipe:    token.Pos(9),
			PreReqs: []ast.Expr{},
			OrderPreReqs: []ast.Expr{
				&ast.Text{Value: "prereq", ValuePos: token.Pos(11)},
			},
			Recipes: []ast.RecipeObj{},
		}))
	})

	It("should Parse a target with a recipe", func() {
		buf := bytes.NewBufferString("target:\n\trecipe")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(9),
				Text: ast.Text{
					Value:    "recipe",
					ValuePos: token.Pos(10),
				},
			}},
		}))
	})

	It("should Parse a target with a semicolon recipe", func() {
		buf := bytes.NewBufferString("target: ; recipe")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.SEMI,
				PrefixPos: token.Pos(9),
				Text: ast.Text{
					Value:    " recipe",
					ValuePos: token.Pos(10),
				},
			}},
		}))
	})

	It("should Parse a semicolon recipe written against the colon", func() {
		buf := bytes.NewBufferString("target:; recipe")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.SEMI,
				PrefixPos: token.Pos(8),
				Text: ast.Text{
					Value:    " recipe",
					ValuePos: token.Pos(9),
				},
			}},
		}))
	})

	// A semicolon is its own token wherever it is written, so ";recipe"
	// introduces a recipe rather than naming a pre-requisite.
	It("should Parse an unseparated semicolon as a recipe", func() {
		buf := bytes.NewBufferString("target: ;recipe")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.SEMI,
				PrefixPos: token.Pos(9),
				Text: ast.Text{
					Value:    "recipe",
					ValuePos: token.Pos(10),
				},
			}},
		}))
	})

	It("should Parse a semicolon recipe after a prereq", func() {
		buf := bytes.NewBufferString("target: prereq ; recipe")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{&ast.Text{
				Value:    "prereq",
				ValuePos: token.Pos(9),
			}},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.SEMI,
				PrefixPos: token.Pos(16),
				Text: ast.Text{
					Value:    " recipe",
					ValuePos: token.Pos(17),
				},
			}},
		}))
	})

	It("should Parse a semicolon recipe after an order-only prereq", func() {
		buf := bytes.NewBufferString("target: | prereq ; recipe")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{},
			Pipe:    token.Pos(9),
			OrderPreReqs: []ast.Expr{&ast.Text{
				Value:    "prereq",
				ValuePos: token.Pos(11),
			}},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.SEMI,
				PrefixPos: token.Pos(18),
				Text: ast.Text{
					Value:    " recipe",
					ValuePos: token.Pos(19),
				},
			}},
		}))
	})

	It("should Parse a semicolon recipe followed by tab recipes", func() {
		buf := bytes.NewBufferString("target: ; recipe\n\trecipe2")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{
				&ast.Recipe{
					Prefix:    token.SEMI,
					PrefixPos: token.Pos(9),
					Text: ast.Text{
						Value:    " recipe",
						ValuePos: token.Pos(10),
					},
				},
				&ast.Recipe{
					Prefix:    token.TAB,
					PrefixPos: token.Pos(18),
					Text: ast.Text{
						Value:    "recipe2",
						ValuePos: token.Pos(19),
					},
				},
			},
		}))
	})

	It("should Parse a recipe introduced by a custom .RECIPEPREFIX", func() {
		buf := bytes.NewBufferString(".RECIPEPREFIX = >\ntarget:\n>echo hi\n")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(
			&ast.Variable{
				Name:  &ast.Text{Value: ".RECIPEPREFIX", ValuePos: token.Pos(1)},
				Op:    token.RECURSIVE_ASSIGN,
				OpPos: token.Pos(15),
				Value: []ast.Expr{&ast.Text{Value: ">", ValuePos: token.Pos(17)}},
			},
			&ast.Rule{
				Colon:        token.Pos(25),
				Targets:      []ast.Expr{&ast.Text{Value: "target", ValuePos: token.Pos(19)}},
				PreReqs:      []ast.Expr{},
				OrderPreReqs: []ast.Expr{},
				Recipes: []ast.RecipeObj{&ast.Recipe{
					Prefix:    token.TEXT,
					PrefixLit: ">",
					PrefixPos: token.Pos(27),
					Text: ast.Text{
						Value:    "echo hi",
						ValuePos: token.Pos(28),
					},
				}},
			},
		))
	})

	It("should Parse a .RECIPEPREFIX that changes partway through the file", func() {
		buf := bytes.NewBufferString(".RECIPEPREFIX = >\na:\n>echo a\n.RECIPEPREFIX = !\nb:\n!echo b\n")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(HaveLen(4))
		Expect(f.Contents[1]).To(Equal(&ast.Rule{
			Colon:        token.Pos(20),
			Targets:      []ast.Expr{&ast.Text{Value: "a", ValuePos: token.Pos(19)}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.TEXT,
				PrefixLit: ">",
				PrefixPos: token.Pos(22),
				Text: ast.Text{
					Value:    "echo a",
					ValuePos: token.Pos(23),
				},
			}},
		}))
		Expect(f.Contents[3]).To(Equal(&ast.Rule{
			Colon:        token.Pos(49),
			Targets:      []ast.Expr{&ast.Text{Value: "b", ValuePos: token.Pos(48)}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.TEXT,
				PrefixLit: "!",
				PrefixPos: token.Pos(51),
				Text: ast.Text{
					Value:    "echo b",
					ValuePos: token.Pos(52),
				},
			}},
		}))
	})

	It("should Parse a tab recipe after .RECIPEPREFIX is set back to empty", func() {
		buf := bytes.NewBufferString(".RECIPEPREFIX = >\na:\n>echo a\n.RECIPEPREFIX =\nb:\n\techo b\n")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(HaveLen(4))
		Expect(f.Contents[3]).To(Equal(&ast.Rule{
			Colon:        token.Pos(47),
			Targets:      []ast.Expr{&ast.Text{Value: "b", ValuePos: token.Pos(46)}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(49),
				Text: ast.Text{
					Value:    "echo b",
					ValuePos: token.Pos(50),
				},
			}},
		}))
	})

	// make uses the first character of the value of .RECIPEPREFIX and ignores
	// the rest of it.
	It("should Parse a recipe introduced by the first character of .RECIPEPREFIX", func() {
		buf := bytes.NewBufferString(".RECIPEPREFIX = >>\ntarget:\n>echo hi\n")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(HaveLen(2))
		Expect(f.Contents[1].(*ast.Rule).Recipes).To(ConsistOf(&ast.Recipe{
			Prefix:    token.TEXT,
			PrefixLit: ">",
			PrefixPos: token.Pos(28),
			Text: ast.Text{
				Value:    "echo hi",
				ValuePos: token.Pos(29),
			},
		}))
	})

	It("should Parse a recipe introduced by an appended .RECIPEPREFIX", func() {
		buf := bytes.NewBufferString(".RECIPEPREFIX += >\ntarget:\n>echo hi\n")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(HaveLen(2))
		Expect(f.Contents[1].(*ast.Rule).Recipes).To(ConsistOf(&ast.Recipe{
			Prefix:    token.TEXT,
			PrefixLit: ">",
			PrefixPos: token.Pos(28),
			Text: ast.Text{
				Value:    "echo hi",
				ValuePos: token.Pos(29),
			},
		}))
	})

	// make evaluates the value of .RECIPEPREFIX, and the parser does not, so
	// the prefix in effect is left alone rather than guessed at.
	It("should keep the recipe prefix when .RECIPEPREFIX is a variable reference", func() {
		buf := bytes.NewBufferString(".RECIPEPREFIX = $(P)\ntarget:\n\techo hi\n")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(HaveLen(2))
		Expect(f.Contents[1].(*ast.Rule).Recipes).To(ConsistOf(&ast.Recipe{
			Prefix:    token.TAB,
			PrefixPos: token.Pos(30),
			Text: ast.Text{
				Value:    "echo hi",
				ValuePos: token.Pos(31),
			},
		}))
	})

	// '.RECIPEPREFIX' always has a value, so '?=' never assigns to it.
	It("should keep the recipe prefix when .RECIPEPREFIX is assigned with '?='", func() {
		buf := bytes.NewBufferString(".RECIPEPREFIX ?= >\ntarget:\n\techo hi\n")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(HaveLen(2))
		Expect(f.Contents[1].(*ast.Rule).Recipes).To(ConsistOf(&ast.Recipe{
			Prefix:    token.TAB,
			PrefixPos: token.Pos(28),
			Text: ast.Text{
				Value:    "echo hi",
				ValuePos: token.Pos(29),
			},
		}))
	})

	// A tab has no meaning once .RECIPEPREFIX names another character. make
	// reports "missing separator" for the line, and the parser records it
	// verbatim the way it records every other line it does not understand.
	It("should Parse a tab line as a bad object when a custom prefix is in effect", func() {
		buf := bytes.NewBufferString(".RECIPEPREFIX = >\ntarget:\n\techo hi\n")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(HaveLen(3))
		Expect(f.Contents[1].(*ast.Rule).Recipes).To(BeEmpty())
		Expect(f.Contents[2]).To(Equal(&ast.BadObj{
			From: token.Pos(27),
			To:   token.Pos(35),
			Text: "\techo hi",
		}))
	})

	It("should Parse a target with multiple recipes", func() {
		buf := bytes.NewBufferString("target:\n\trecipe\n\trecipe2")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{
				&ast.Recipe{
					Prefix:    token.TAB,
					PrefixPos: token.Pos(9),
					Text: ast.Text{
						Value:    "recipe",
						ValuePos: token.Pos(10),
					},
				},
				&ast.Recipe{
					Prefix:    token.TAB,
					PrefixPos: token.Pos(17),
					Text: ast.Text{
						Value:    "recipe2",
						ValuePos: token.Pos(18),
					},
				},
			},
		}))
	})

	It("should Parse recipes separated by a blank line", func() {
		buf := bytes.NewBufferString("target:\n\trecipe\n\n\trecipe2")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{
				&ast.Recipe{
					Prefix:    token.TAB,
					PrefixPos: token.Pos(9),
					Text: ast.Text{
						Value:    "recipe",
						ValuePos: token.Pos(10),
					},
				},
				&ast.Recipe{
					Prefix:    token.TAB,
					PrefixPos: token.Pos(18),
					Text: ast.Text{
						Value:    "recipe2",
						ValuePos: token.Pos(19),
					},
				},
			},
		}))
	})

	It("should Parse recipes separated by multiple blank lines", func() {
		buf := bytes.NewBufferString("target:\n\trecipe\n\n\n\trecipe2")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{
				&ast.Recipe{
					Prefix:    token.TAB,
					PrefixPos: token.Pos(9),
					Text: ast.Text{
						Value:    "recipe",
						ValuePos: token.Pos(10),
					},
				},
				&ast.Recipe{
					Prefix:    token.TAB,
					PrefixPos: token.Pos(19),
					Text: ast.Text{
						Value:    "recipe2",
						ValuePos: token.Pos(20),
					},
				},
			},
		}))
	})

	It("should Parse a recipe preceded by a blank line", func() {
		buf := bytes.NewBufferString("target:\n\n\trecipe")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(10),
				Text: ast.Text{
					Value:    "recipe",
					ValuePos: token.Pos(11),
				},
			}},
		}))
	})

	It("should not attach the recipes of the next rule after a blank line", func() {
		buf := bytes.NewBufferString("target:\n\trecipe\n\nother:\n\trecipe2")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(HaveExactElements(
			&ast.Rule{
				Colon: token.Pos(7),
				Targets: []ast.Expr{&ast.Text{
					Value:    "target",
					ValuePos: token.Pos(1),
				}},
				PreReqs:      []ast.Expr{},
				OrderPreReqs: []ast.Expr{},
				Recipes: []ast.RecipeObj{&ast.Recipe{
					Prefix:    token.TAB,
					PrefixPos: token.Pos(9),
					Text: ast.Text{
						Value:    "recipe",
						ValuePos: token.Pos(10),
					},
				}},
			},
			&ast.Rule{
				Colon: token.Pos(23),
				Targets: []ast.Expr{&ast.Text{
					Value:    "other",
					ValuePos: token.Pos(18),
				}},
				PreReqs:      []ast.Expr{},
				OrderPreReqs: []ast.Expr{},
				Recipes: []ast.RecipeObj{&ast.Recipe{
					Prefix:    token.TAB,
					PrefixPos: token.Pos(25),
					Text: ast.Text{
						Value:    "recipe2",
						ValuePos: token.Pos(26),
					},
				}},
			},
		))
	})

	It("should Parse a target with spaces in the recipe", func() {
		buf := bytes.NewBufferString("target:\n\trecipe part2")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(9),
				Text: ast.Text{
					Value:    "recipe part2",
					ValuePos: token.Pos(10),
				},
			}},
		}))
	})

	It("should preserve double quotes in a recipe", func() {
		buf := bytes.NewBufferString("target:\n\t@echo \"testing has been started...\"")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(9),
				Text: ast.Text{
					Value:    "@echo \"testing has been started...\"",
					ValuePos: token.Pos(10),
				},
			}},
		}))
	})

	It("should preserve comments and operator tokens in a recipe", func() {
		buf := bytes.NewBufferString("target:\n\t@echo $@: # keep this comment")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(9),
				Text: ast.Text{
					Value:    "@echo $@: # keep this comment",
					ValuePos: token.Pos(10),
				},
			}},
		}))
	})

	It("should preserve an unsupported token in a recipe", func() {
		buf := bytes.NewBufferString("target:\n\t@echo a\rb")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(9),
				Text: ast.Text{
					Value:    "@echo a\rb",
					ValuePos: token.Pos(10),
				},
			}},
		}))
	})

	It("should Parse a target with a prereq and a recipe", func() {
		buf := bytes.NewBufferString("target: prereq\n\trecipe")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{&ast.Text{
				Value:    "prereq",
				ValuePos: token.Pos(9),
			}},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(16),
				Text: ast.Text{
					Value:    "recipe",
					ValuePos: token.Pos(17),
				},
			}},
		}))
	})

	It("should Parse a semicolon inside a recipe as recipe text", func() {
		buf := bytes.NewBufferString("target:\n\tcd dir; ls")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes: []ast.RecipeObj{&ast.Recipe{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(9),
				Text: ast.Text{
					Value:    "cd dir; ls",
					ValuePos: token.Pos(10),
				},
			}},
		}))
	})

	It("should Parse a semicolon in a variable value as text", func() {
		buf := bytes.NewBufferString("VAR = a; b")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Variable{
			Name: &ast.Text{
				Value:    "VAR",
				ValuePos: token.Pos(1),
			},
			Op:    token.RECURSIVE_ASSIGN,
			OpPos: token.Pos(5),
			Value: []ast.Expr{
				&ast.JuxtaposedExpr{Parts: []ast.Expr{
					&ast.Text{Value: "a", ValuePos: token.Pos(7)},
					&ast.Text{Value: ";", ValuePos: token.Pos(8)},
				}},
				&ast.Text{Value: "b", ValuePos: token.Pos(10)},
			},
		}))
	})

	It("should support a nil *token.File value", func() {
		buf := bytes.NewBufferString("target:")
		s := parser.New(buf, nil)

		f, err := s.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).NotTo(BeEmpty())
	})

	DescribeTable("should parse a variable definition",
		func(input string, op token.Token, vpos int) {
			buf := bytes.NewBufferString(input)
			s := parser.New(buf, file)

			f, err := s.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.Variable{
				Name: &ast.Text{
					Value:    "VAR",
					ValuePos: token.Pos(1),
				},
				Op:    op,
				OpPos: token.Pos(5),
				Value: []ast.Expr{&ast.Text{
					Value:    "test",
					ValuePos: token.Pos(vpos),
				}},
			}))
		},
		Entry(nil, "VAR := test", token.SIMPLE_ASSIGN, 8),
		Entry(nil, "VAR ::= test", token.POSIX_ASSIGN, 9),
		Entry(nil, "VAR :::= test", token.IMMEDIATE_ASSIGN, 10),
		Entry(nil, "VAR != test", token.SHELL_ASSIGN, 8),
		Entry(nil, "VAR += test", token.APPEND_ASSIGN, 8),
		Entry(nil, "VAR ?= test", token.IFNDEF_ASSIGN, 8),
		Entry(nil, "VAR = test", token.RECURSIVE_ASSIGN, 7),
	)

	DescribeTable("should parse a space-separated variable definition",
		func(input string, op token.Token, vpos int) {
			buf := bytes.NewBufferString(input)
			s := parser.New(buf, file)

			f, err := s.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.Variable{
				Name: &ast.Text{
					Value:    "VAR",
					ValuePos: token.Pos(1),
				},
				Op:    op,
				OpPos: token.Pos(5),
				Value: []ast.Expr{
					&ast.Text{
						Value:    "test",
						ValuePos: token.Pos(vpos),
					},
					&ast.Text{
						Value:    "test2",
						ValuePos: token.Pos(vpos + 5),
					},
				},
			}))
		},
		Entry(nil, "VAR := test test2", token.SIMPLE_ASSIGN, 8),
		Entry(nil, "VAR ::= test test2", token.POSIX_ASSIGN, 9),
		Entry(nil, "VAR :::= test test2", token.IMMEDIATE_ASSIGN, 10),
		Entry(nil, "VAR != test test2", token.SHELL_ASSIGN, 8),
		Entry(nil, "VAR += test test2", token.APPEND_ASSIGN, 8),
		Entry(nil, "VAR ?= test test2", token.IFNDEF_ASSIGN, 8),
		Entry(nil, "VAR = test test2", token.RECURSIVE_ASSIGN, 7),
	)

	DescribeTable("should parse an unspaced variable definition",
		func(input string, op token.Token, opPos, vpos int) {
			buf := bytes.NewBufferString(input)
			s := parser.New(buf, file)

			f, err := s.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.Variable{
				Name: &ast.Text{
					Value:    "VAR",
					ValuePos: token.Pos(1),
				},
				Op:    op,
				OpPos: token.Pos(opPos),
				Value: []ast.Expr{&ast.Text{
					Value:    "x",
					ValuePos: token.Pos(vpos),
				}},
			}))
		},
		Entry(nil, "VAR=x", token.RECURSIVE_ASSIGN, 4, 5),
		Entry(nil, "VAR:=x", token.SIMPLE_ASSIGN, 4, 6),
		Entry(nil, "VAR::=x", token.POSIX_ASSIGN, 4, 7),
		Entry(nil, "VAR:::=x", token.IMMEDIATE_ASSIGN, 4, 8),
		Entry(nil, "VAR?=x", token.IFNDEF_ASSIGN, 4, 6),
		Entry(nil, "VAR!=x", token.SHELL_ASSIGN, 4, 6),
		Entry(nil, "VAR+=x", token.APPEND_ASSIGN, 4, 6),
	)

	DescribeTable("should parse an unspaced variable declaration",
		func(input string, op token.Token) {
			buf := bytes.NewBufferString(input)
			s := parser.New(buf, file)

			f, err := s.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.Variable{
				Name: &ast.Text{
					Value:    "VAR",
					ValuePos: token.Pos(1),
				},
				Op:    op,
				OpPos: token.Pos(4),
				Value: nil,
			}))
		},
		Entry(nil, "VAR=", token.RECURSIVE_ASSIGN),
		Entry(nil, "VAR?=", token.IFNDEF_ASSIGN),
		Entry(nil, "VAR!=", token.SHELL_ASSIGN),
		Entry(nil, "VAR+=", token.APPEND_ASSIGN),
	)

	// make reads the first assignment operator on the line and takes the rest
	// of the line as the value, so every operator after it is text.
	It("should parse an assignment operator in a variable value as text", func() {
		buf := bytes.NewBufferString("VAR = a = b")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Variable{
			Name: &ast.Text{
				Value:    "VAR",
				ValuePos: token.Pos(1),
			},
			Op:    token.RECURSIVE_ASSIGN,
			OpPos: token.Pos(5),
			Value: []ast.Expr{
				&ast.Text{Value: "a", ValuePos: token.Pos(7)},
				&ast.Text{Value: "=", ValuePos: token.Pos(9)},
				&ast.Text{Value: "b", ValuePos: token.Pos(11)},
			},
		}))
	})

	It("should join an assignment operator written against a value", func() {
		buf := bytes.NewBufferString("VAR = a!=b")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Variable{
			Name: &ast.Text{
				Value:    "VAR",
				ValuePos: token.Pos(1),
			},
			Op:    token.RECURSIVE_ASSIGN,
			OpPos: token.Pos(5),
			Value: []ast.Expr{
				&ast.JuxtaposedExpr{Parts: []ast.Expr{
					&ast.Text{Value: "a", ValuePos: token.Pos(7)},
					&ast.Text{Value: "!=", ValuePos: token.Pos(8)},
					&ast.Text{Value: "b", ValuePos: token.Pos(10)},
				}},
			},
		}))
	})

	// A ':' written in a value is text as well. Only the first one on the line
	// can open a rule, and an operator earlier in the line has already said the
	// line is a variable.
	It("should parse a colon in a variable value as text", func() {
		buf := bytes.NewBufferString("a=b: c")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Variable{
			Name: &ast.Text{
				Value:    "a",
				ValuePos: token.Pos(1),
			},
			Op:    token.RECURSIVE_ASSIGN,
			OpPos: token.Pos(2),
			Value: []ast.Expr{
				&ast.JuxtaposedExpr{Parts: []ast.Expr{
					&ast.Text{Value: "b", ValuePos: token.Pos(3)},
					&ast.Text{Value: ":", ValuePos: token.Pos(4)},
				}},
				&ast.Text{Value: "c", ValuePos: token.Pos(6)},
			},
		}))
	})

	It("should parse a name holding an assignment operator", func() {
		buf := bytes.NewBufferString("VAR = $(a!=b)")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Variable{
			Name: &ast.Text{
				Value:    "VAR",
				ValuePos: token.Pos(1),
			},
			Op:    token.RECURSIVE_ASSIGN,
			OpPos: token.Pos(5),
			Value: []ast.Expr{&ast.VarRef{
				Dollar: token.Pos(7),
				Open:   token.LPAREN,
				Name:   "a!=b",
				Close:  token.RPAREN,
			}},
		}))
	})

	// make reads "target: a=b" as a target-specific variable, which this
	// package does not model. The prerequisite it becomes instead is one word
	// rather than three, so the line prints back as it was written.
	It("should parse an unspaced assignment in a prerequisite as one prerequisite", func() {
		buf := bytes.NewBufferString("target: a=b c!=d")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(HaveLen(1))
		rule, ok := f.Contents[0].(*ast.Rule)
		Expect(ok).To(BeTrue(), "expected a *ast.Rule, got %T", f.Contents[0])
		Expect(rule.PreReqs).To(HaveLen(2))
	})

	DescribeTable("should parse a variable declaration",
		func(input string, op token.Token) {
			buf := bytes.NewBufferString(input)
			s := parser.New(buf, file)

			f, err := s.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.Variable{
				Name: &ast.Text{
					Value:    "VAR",
					ValuePos: token.Pos(1),
				},
				Op:    op,
				OpPos: token.Pos(5),
				Value: nil,
			}))
		},
		Entry(nil, "VAR :=", token.SIMPLE_ASSIGN),
		Entry(nil, "VAR ::=", token.POSIX_ASSIGN),
		Entry(nil, "VAR :::=", token.IMMEDIATE_ASSIGN),
		Entry(nil, "VAR !=", token.SHELL_ASSIGN),
		Entry(nil, "VAR +=", token.APPEND_ASSIGN),
		Entry(nil, "VAR ?=", token.IFNDEF_ASSIGN),
		Entry(nil, "VAR =", token.RECURSIVE_ASSIGN),
	)

	It("should Parse an ifeq conditional directive", func() {
		buf := bytes.NewBufferString("ifeq (baz, bin)\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Arg1: &ast.Text{
					Value:    "baz",
					ValuePos: token.Pos(7),
				},
				Comma: token.Pos(10),
				Arg2: &ast.Text{
					Value:    "bin",
					ValuePos: token.Pos(12),
				},
				Close: token.Pos(15),
			},
			Endif: token.Pos(17),
		}))
	})

	It("should Parse an ifneq conditional directive", func() {
		buf := bytes.NewBufferString("ifneq (baz, bin)\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFNEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(7),
				Arg1: &ast.Text{
					Value:    "baz",
					ValuePos: token.Pos(8),
				},
				Comma: token.Pos(11),
				Arg2: &ast.Text{
					Value:    "bin",
					ValuePos: token.Pos(13),
				},
				Close: token.Pos(16),
			},
			Endif: token.Pos(18),
		}))
	})

	It("should Parse an ifeq conditional directive with quotes", func() {
		buf := bytes.NewBufferString("ifeq 'baz' \"bin\"\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Arg1: &ast.QuotedExpr{
					Quote: token.APOS,
					Open:  token.Pos(6),
					Value: &ast.Text{
						Value:    "baz",
						ValuePos: token.Pos(7),
					},
					Close: token.Pos(10),
				},
				Arg2: &ast.QuotedExpr{
					Quote: token.QUOTE,
					Open:  token.Pos(12),
					Value: &ast.Text{
						Value:    "bin",
						ValuePos: token.Pos(13),
					},
					Close: token.Pos(16),
				},
			},
			Endif: token.Pos(18),
		}))
	})

	It("should Parse an ifeq conditional directive with an empty first argument", func() {
		buf := bytes.NewBufferString("ifeq (,bin)\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Comma:  token.Pos(7),
				Arg2: &ast.Text{
					Value:    "bin",
					ValuePos: token.Pos(8),
				},
				Close: token.Pos(11),
			},
			Endif: token.Pos(13),
		}))
	})

	It("should Parse an ifeq conditional directive with an empty second argument", func() {
		buf := bytes.NewBufferString("ifeq (baz,)\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Arg1: &ast.Text{
					Value:    "baz",
					ValuePos: token.Pos(7),
				},
				Comma: token.Pos(10),
				Close: token.Pos(11),
			},
			Endif: token.Pos(13),
		}))
	})

	It("should Parse an ifeq conditional directive with two empty arguments", func() {
		buf := bytes.NewBufferString("ifeq (,)\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Comma:  token.Pos(7),
				Close:  token.Pos(8),
			},
			Endif: token.Pos(10),
		}))
	})

	It("should Parse an ifeq conditional directive with an empty quoted first argument", func() {
		buf := bytes.NewBufferString("ifeq '' \"bin\"\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
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
						Value:    "bin",
						ValuePos: token.Pos(10),
					},
					Close: token.Pos(13),
				},
			},
			Endif: token.Pos(15),
		}))
	})

	It("should Parse an ifeq conditional directive with an empty quoted second argument", func() {
		buf := bytes.NewBufferString("ifeq 'baz' \"\"\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Arg1: &ast.QuotedExpr{
					Quote: token.APOS,
					Open:  token.Pos(6),
					Value: &ast.Text{
						Value:    "baz",
						ValuePos: token.Pos(7),
					},
					Close: token.Pos(10),
				},
				Arg2: &ast.QuotedExpr{
					Quote: token.QUOTE,
					Open:  token.Pos(12),
					Close: token.Pos(13),
				},
			},
			Endif: token.Pos(15),
		}))
	})

	DescribeTable("should error when quotes are mismatched",
		Entry(nil, "ifeq 'baz\" \"bin\"\nendif", "test:1:10: expected ''', found '\"'"),
		Entry(nil, "ifeq \"baz' \"bin\"\nendif", "test:1:10: expected '\"', found '''"),
		Entry(nil, "ifeq 'baz' \"bin'\nendif", "test:1:16: expected '\"', found '''"),
		Entry(nil, "ifeq 'baz' 'bin\"\nendif", "test:1:16: expected ''', found '\"'"),
		func(input, msg string) {
			buf := bytes.NewBufferString(input)
			p := parser.New(buf, file)

			_, err := p.ParseFile()

			Expect(err).To(MatchError(msg))
		},
	)

	DescribeTable("should error once when a quoted argument is not quoted",
		Entry(nil, "ifeq \"baz\" bin\nendif", "test:1:12: expected one of ''', '\"', found bin"),
		Entry(nil, "ifeq 'baz' bin\nendif", "test:1:12: expected one of ''', '\"', found bin"),
		Entry(nil, "ifeq \"\" bin\nendif", "test:1:9: expected one of ''', '\"', found bin"),
		func(input, msg string) {
			buf := bytes.NewBufferString(input)
			p := parser.New(buf, file)

			_, err := p.ParseFile()

			// The argument is left absent rather than recovered into a node
			// with no quote and no position to record, so the single error
			// naming the unquoted token is the only one reported.
			Expect(err).To(MatchError(msg))
		},
	)

	DescribeTable("should error on unexpected tokens",
		Entry(nil, "ifeq )baz\" \"bin\"\nendif"),
		Entry(nil, "ifeq {baz' \"bin\"\nendif"),
		Entry(nil, "ifeq }baz' \"bin'\nendif"),
		Entry(nil, "ifeq |baz' 'bin\"\nendif"),
		func(input string) {
			buf := bytes.NewBufferString(input)
			p := parser.New(buf, file)

			_, err := p.ParseFile()

			Expect(err).To(MatchError(ContainSubstring("expected one of '(', ''', '\"'")))
		},
	)

	It("should Parse an ifdef conditional directive", func() {
		buf := bytes.NewBufferString("ifdef FOO\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfdefDir{
				Tok:    token.IFDEF,
				TokPos: token.Pos(1),
				VarName: &ast.Text{
					Value:    "FOO",
					ValuePos: token.Pos(7),
				},
			},
			Endif: token.Pos(11),
		}))
	})

	It("should Parse an ifndef conditional directive", func() {
		buf := bytes.NewBufferString("ifndef FOO\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfdefDir{
				Tok:    token.IFNDEF,
				TokPos: token.Pos(1),
				VarName: &ast.Text{
					Value:    "FOO",
					ValuePos: token.Pos(8),
				},
			},
			Endif: token.Pos(12),
		}))
	})

	It("should Parse a conditional directive with text", func() {
		buf := bytes.NewBufferString("ifeq (baz, bin)\ntarget:\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Arg1: &ast.Text{
					Value:    "baz",
					ValuePos: token.Pos(7),
				},
				Comma: token.Pos(10),
				Arg2: &ast.Text{
					Value:    "bin",
					ValuePos: token.Pos(12),
				},
				Close: token.Pos(15),
			},
			Text: []ast.Obj{&ast.Rule{
				Targets: []ast.Expr{&ast.Text{
					Value:    "target",
					ValuePos: token.Pos(17),
				}},
				Colon:        token.Pos(23),
				PreReqs:      []ast.Expr{},
				OrderPreReqs: []ast.Expr{},
				Recipes:      []ast.RecipeObj{},
			}},
			Endif: token.Pos(25),
		}))
	})

	It("should Parse a conditional directive with an else block", func() {
		buf := bytes.NewBufferString("ifeq (baz, bin)\nelse\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Arg1: &ast.Text{
					Value:    "baz",
					ValuePos: token.Pos(7),
				},
				Comma: token.Pos(10),
				Arg2: &ast.Text{
					Value:    "bin",
					ValuePos: token.Pos(12),
				},
				Close: token.Pos(15),
			},
			Else: []*ast.ElseBlock{{
				Else: token.Pos(17),
			}},
			Endif: token.Pos(22),
		}))
	})

	It("should Parse a conditional directive with an else block with text", func() {
		buf := bytes.NewBufferString("ifeq (baz, bin)\nelse\ntarget:\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Arg1: &ast.Text{
					Value:    "baz",
					ValuePos: token.Pos(7),
				},
				Comma: token.Pos(10),
				Arg2: &ast.Text{
					Value:    "bin",
					ValuePos: token.Pos(12),
				},
				Close: token.Pos(15),
			},
			Else: []*ast.ElseBlock{{
				Else: token.Pos(17),
				Text: []ast.Obj{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(22),
					}},
					Colon:        token.Pos(28),
					PreReqs:      []ast.Expr{},
					OrderPreReqs: []ast.Expr{},
					Recipes:      []ast.RecipeObj{},
				}},
			}},
			Endif: token.Pos(30),
		}))
	})

	It("should Parse a conditional directive with an else block that has a condition", func() {
		buf := bytes.NewBufferString("ifeq (baz, bin)\nelse ifeq (baz, bin)\ntarget:\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Arg1: &ast.Text{
					Value:    "baz",
					ValuePos: token.Pos(7),
				},
				Comma: token.Pos(10),
				Arg2: &ast.Text{
					Value:    "bin",
					ValuePos: token.Pos(12),
				},
				Close: token.Pos(15),
			},
			Else: []*ast.ElseBlock{{
				Else: token.Pos(17),
				Condition: &ast.IfeqDir{
					Tok:    token.IFEQ,
					TokPos: token.Pos(22),
					Open:   token.Pos(27),
					Arg1: &ast.Text{
						Value:    "baz",
						ValuePos: token.Pos(28),
					},
					Comma: token.Pos(31),
					Arg2: &ast.Text{
						Value:    "bin",
						ValuePos: token.Pos(33),
					},
					Close: token.Pos(36),
				},
				Text: []ast.Obj{&ast.Rule{
					Targets: []ast.Expr{&ast.Text{
						Value:    "target",
						ValuePos: token.Pos(38),
					}},
					Colon:        token.Pos(44),
					PreReqs:      []ast.Expr{},
					OrderPreReqs: []ast.Expr{},
					Recipes:      []ast.RecipeObj{},
				}},
			}},
			Endif: token.Pos(46),
		}))
	})

	It("should error when a plain else block preceds an else block with a condition", func() {
		buf := bytes.NewBufferString(`ifeq (baz, bin)
else
else ifeq (baz, bin)
endif
`)

		p := parser.New(buf, file)

		_, err := p.ParseFile()

		Expect(err).To(MatchError("test:3:1: expected 'endif', found 'else'"))
	})

	It("should Parse a conditional directive with lots of stuff in it", func() {
		buf := bytes.NewBufferString(`ifeq (baz, bin)
FOO := BAR
else ifdef test
BAR ?=
else
BAR :::=
endif
`)

		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfeqDir{
				Tok:    token.IFEQ,
				TokPos: token.Pos(1),
				Open:   token.Pos(6),
				Arg1: &ast.Text{
					Value:    "baz",
					ValuePos: token.Pos(7),
				},
				Comma: token.Pos(10),
				Arg2: &ast.Text{
					Value:    "bin",
					ValuePos: token.Pos(12),
				},
				Close: token.Pos(15),
			},
			Text: []ast.Obj{&ast.Variable{
				Name: &ast.Text{
					Value:    "FOO",
					ValuePos: token.Pos(17),
				},
				Op:    token.SIMPLE_ASSIGN,
				OpPos: token.Pos(21),
				Value: []ast.Expr{&ast.Text{
					Value:    "BAR",
					ValuePos: token.Pos(24),
				}},
			}},
			Else: []*ast.ElseBlock{
				{
					Else: token.Pos(28),
					Condition: &ast.IfdefDir{
						Tok:    token.IFDEF,
						TokPos: token.Pos(33),
						VarName: &ast.Text{
							Value:    "test",
							ValuePos: token.Pos(39),
						},
					},
					Text: []ast.Obj{&ast.Variable{
						Name: &ast.Text{
							Value:    "BAR",
							ValuePos: token.Pos(44),
						},
						Op:    token.IFNDEF_ASSIGN,
						OpPos: token.Pos(48),
					}},
				},
				{
					Else: token.Pos(51),
					Text: []ast.Obj{&ast.Variable{
						Name: &ast.Text{
							Value:    "BAR",
							ValuePos: token.Pos(56),
						},
						Op:    token.IMMEDIATE_ASSIGN,
						OpPos: token.Pos(60),
					}},
				},
			},
			Endif: token.Pos(65),
		}))
	})

	DescribeTable("should error when an assignment has no name",
		func(input string) {
			buf := bytes.NewBufferString(input)
			s := parser.New(buf, file)

			_, err := s.ParseFile()

			Expect(err).To(MatchError("test:1:1: variable name is empty"))
		},
		Entry(nil, "=x"),
		Entry(nil, "= x"),
		Entry(nil, "!=x"),
		Entry(nil, "+= x"),
	)

	It("should error with extra text to the left of the assignment", func() {
		buf := bytes.NewBufferString("VAR invalid :=")
		s := parser.New(buf, file)

		_, err := s.ParseFile()

		Expect(err).To(MatchError("test:1:13: variable may have only one name"))
	})

	// make reads a spaced assignment operator on a target line as a
	// target-specific variable. This package does not model one, and reading
	// the operator as a prerequisite would print the line back with a
	// structure make never gave it, so the operator is an error where a
	// prerequisite was expected.
	DescribeTable("should error on a target-specific variable",
		func(input, msg string) {
			buf := bytes.NewBufferString(input)
			s := parser.New(buf, file)

			_, err := s.ParseFile()

			Expect(err).To(MatchError(ContainSubstring(msg)))
		},
		Entry(nil, "target: VAR = value", "test:1:13: expected one of 'TEXT', '$', found '='"),
		Entry(nil, "%.o: VAR = value", "test:1:10: expected one of 'TEXT', '$', found '='"),
		Entry(nil, "target: VAR += value", "test:1:13: expected one of 'TEXT', '$', found '+='"),
	)

	Describe("definition blocks", func() {
		It("should parse a define block as a single object", func() {
			buf := bytes.NewBufferString("define greeting\nhello\nendef\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.DefineDir{
				Define: token.Pos(1),
				VarName: &ast.Text{
					Value:    "greeting",
					ValuePos: token.Pos(8),
				},
				Op:    token.ILLEGAL,
				Body:  []*ast.Text{{Value: "hello", ValuePos: token.Pos(17)}},
				Endef: token.Pos(23),
			}))
		})

		DescribeTable("should parse the assignment operator of a define",
			func(input string, op token.Token) {
				p := parser.New(bytes.NewBufferString(input), file)

				f, err := p.ParseFile()

				Expect(err).NotTo(HaveOccurred())
				Expect(f.Contents).To(HaveLen(1))
				Expect(f.Contents[0]).To(HaveField("Op", op))
			},
			Entry(nil, "define FOO\nx\nendef\n", token.ILLEGAL),
			Entry(nil, "define FOO =\nx\nendef\n", token.RECURSIVE_ASSIGN),
			Entry(nil, "define FOO :=\nx\nendef\n", token.SIMPLE_ASSIGN),
			Entry(nil, "define FOO ::=\nx\nendef\n", token.POSIX_ASSIGN),
			Entry(nil, "define FOO :::=\nx\nendef\n", token.IMMEDIATE_ASSIGN),
			Entry(nil, "define FOO +=\nx\nendef\n", token.APPEND_ASSIGN),
			Entry(nil, "define FOO ?=\nx\nendef\n", token.IFNDEF_ASSIGN),
			Entry(nil, "define FOO !=\nx\nendef\n", token.SHELL_ASSIGN),
		)

		It("should record the position of the assignment operator", func() {
			buf := bytes.NewBufferString("define FOO +=\nx\nendef\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.DefineDir{
				Define: token.Pos(1),
				VarName: &ast.Text{
					Value:    "FOO",
					ValuePos: token.Pos(8),
				},
				Op:    token.APPEND_ASSIGN,
				OpPos: token.Pos(12),
				Body:  []*ast.Text{{Value: "x", ValuePos: token.Pos(15)}},
				Endef: token.Pos(17),
			}))
		})

		It("should parse a body of several lines", func() {
			buf := bytes.NewBufferString("define FOO\none\ntwo\nendef\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(HaveLen(1))
			Expect(f.Contents[0]).To(HaveField("Body", []*ast.Text{
				{Value: "one", ValuePos: token.Pos(12)},
				{Value: "two", ValuePos: token.Pos(16)},
			}))
		})

		It("should parse a define with an empty body", func() {
			buf := bytes.NewBufferString("define FOO\nendef\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.DefineDir{
				Define: token.Pos(1),
				VarName: &ast.Text{
					Value:    "FOO",
					ValuePos: token.Pos(8),
				},
				Op:    token.ILLEGAL,
				Endef: token.Pos(12),
			}))
		})

		// make reads the body of a define as the text it was written with, so
		// a line that looks like make syntax is a line of the value.
		DescribeTable("should keep a body line as text",
			func(input string, body ...string) {
				p := parser.New(bytes.NewBufferString(input), file)

				f, err := p.ParseFile()

				Expect(err).NotTo(HaveOccurred())
				Expect(f.Contents).To(HaveLen(1))
				d, ok := f.Contents[0].(*ast.DefineDir)
				Expect(ok).To(BeTrue(), "expected a *ast.DefineDir, got %T", f.Contents[0])

				values := make([]string, len(d.Body))
				for i, line := range d.Body {
					values[i] = line.Value
				}
				Expect(values).To(Equal(body))
			},
			Entry("a rule", "define FOO\ntarget: prereq\nendef\n",
				"target: prereq",
			),
			Entry("a recipe", "define FOO\ntarget:\n\techo hi\nendef\n",
				"target:", "\techo hi",
			),
			Entry("a conditional", "define FOO\nifeq (a,b)\nx = 1\nendif\nendef\n",
				"ifeq (a,b)", "x = 1", "endif",
			),
			Entry("a comment", "define FOO\n# comment\nendef\n",
				"# comment",
			),
			Entry("an expansion", "define FOO\n$(shell date)\nendef\n",
				"$(shell date)",
			),
			Entry("a blank line", "define FOO\none\n\ntwo\nendef\n",
				"one", "", "two",
			),
			Entry("the blanks a line begins with", "define FOO\n  indented\nendef\n",
				"  indented",
			),
		)

		// make counts a define written in a body, so the endef matching it
		// closes the inner block rather than the outer one and the value of
		// the outer variable holds the inner block as text.
		It("should parse a nested define as text of the outer body", func() {
			buf := bytes.NewBufferString("define OUTER\ndefine INNER\nx\nendef\ntail\nendef\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(HaveLen(1))
			Expect(f.Contents[0]).To(HaveField("Body", []*ast.Text{
				{Value: "define INNER", ValuePos: token.Pos(14)},
				{Value: "x", ValuePos: token.Pos(27)},
				{Value: "endef", ValuePos: token.Pos(29)},
				{Value: "tail", ValuePos: token.Pos(35)},
			}))
		})

		// make ends a define at a line holding nothing but endef, so an endef
		// written behind a tab is a line of the body. Here it leaves the block
		// unterminated, which make reports as well.
		DescribeTable("should not end a define at",
			func(input string) {
				p := parser.New(bytes.NewBufferString(input), file)

				_, err := p.ParseFile()

				Expect(err).To(MatchError(ContainSubstring("expected 'endef'")))
			},
			Entry("an endef behind a tab", "define FOO\n\tendef\n"),
			Entry("an endef carrying text", "define FOO\nendef x\n"),
		)

		It("should error when a define reaches the end of the file", func() {
			buf := bytes.NewBufferString("define greeting\nhello\n")
			p := parser.New(buf, file)

			_, err := p.ParseFile()

			Expect(err).To(MatchError("test:2:6: expected 'endef', found 'EOF'"))
		})

		DescribeTable("should keep a directive the parser cannot read as text",
			func(input, text string) {
				p := parser.New(bytes.NewBufferString(input), file)

				f, _ := p.ParsePartial()

				Expect(f.Contents).To(ContainElement(&ast.BadObj{
					From: token.Pos(1),
					To:   token.Pos(1 + len(text)),
					Text: text,
				}))
			},
			Entry("a name written with blanks",
				"define FOO bar\nx\nendef\n", "define FOO bar",
			),
			Entry("text after the assignment operator",
				"define FOO := x\nendef\n", "define FOO := x",
			),
			Entry("an undefine naming several variables",
				"undefine FOO bar\n", "undefine FOO bar",
			),
		)
	})

	Describe("undefine directives", func() {
		It("should parse an undefine directive", func() {
			buf := bytes.NewBufferString("undefine VAR\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.UndefineDir{
				Undefine: token.Pos(1),
				VarName: &ast.Text{
					Value:    "VAR",
					ValuePos: token.Pos(10),
				},
			}))
		})

		It("should parse a variable reference as the name", func() {
			buf := bytes.NewBufferString("undefine $(VAR)\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.UndefineDir{
				Undefine: token.Pos(1),
				VarName: &ast.VarRef{
					Dollar: token.Pos(10),
					Open:   token.LPAREN,
					Name:   "VAR",
					Close:  token.RPAREN,
				},
			}))
		})
	})

	// make reads the whole line before the operator as the name of the
	// variable, this parser reads one expression the way it does for an
	// ordinary assignment, and the line it cannot read is kept as the text it
	// was written with.
	DescribeTable("should error when a directive names several variables",
		func(input, message string) {
			p := parser.New(bytes.NewBufferString(input), file)

			_, err := p.ParseFile()

			Expect(err).To(MatchError(message))
		},
		Entry(nil, "define FOO bar\nx\nendef\n", "test:1:15: variable may have only one name"),
		Entry(nil, "undefine FOO bar\n", "test:1:17: variable may have only one name"),
	)

	DescribeTable("should parse an unsupported line as a bad object",
		Entry(nil, "endef"),
		Entry(nil, "override VAR = x"),
		Entry(nil, "export VAR"),
		Entry(nil, "unexport VAR"),
		Entry(nil, "private VAR = x"),
		Entry(nil, "include foo.mk"),
		Entry(nil, "-include foo.mk"),
		Entry(nil, "sinclude foo.mk"),
		Entry(nil, "vpath %.c src"),
		Entry(nil, "load foo.so"),
		Entry(nil, "|"),
		Entry(nil, "a\rb"),
		func(input string) {
			buf := bytes.NewBufferString(input)
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.BadObj{
				From: token.Pos(1),
				To:   token.Pos(1 + len(input)),
				Text: input,
			}))
		},
	)

	It("should parse an unsupported line without recording an error", func() {
		buf := bytes.NewBufferString("include foo.mk\ntarget:")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).NotTo(ContainElement(BeNil()))
		Expect(f.Contents).To(HaveLen(2))
	})

	It("should stop the bad object at the end of the line", func() {
		buf := bytes.NewBufferString("include foo.mk\ninclude bar.mk")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(HaveExactElements(
			&ast.BadObj{
				From: token.Pos(1),
				To:   token.Pos(15),
				Text: "include foo.mk",
			},
			&ast.BadObj{
				From: token.Pos(16),
				To:   token.Pos(30),
				Text: "include bar.mk",
			},
		))
	})

	It("should include the recipe prefix of an unattached recipe line", func() {
		buf := bytes.NewBufferString("FOO := bar\n\tsecond")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ContainElement(&ast.BadObj{
			From: token.Pos(12),
			To:   token.Pos(19),
			Text: "\tsecond",
		}))
	})

	It("should parse a bad object in a conditional directive", func() {
		buf := bytes.NewBufferString("ifdef FOO\ninclude foo.mk\nendif")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.IfBlock{
			Directive: &ast.IfdefDir{
				Tok:    token.IFDEF,
				TokPos: token.Pos(1),
				VarName: &ast.Text{
					Value:    "FOO",
					ValuePos: token.Pos(7),
				},
			},
			Text: []ast.Obj{&ast.BadObj{
				From: token.Pos(11),
				To:   token.Pos(25),
				Text: "include foo.mk",
			}},
			Endif: token.Pos(26),
		}))
	})

	It("should Parse a function call", func() {
		buf := bytes.NewBufferString("A := $(shell pwd)")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Variable{
			Name:  &ast.Text{Value: "A", ValuePos: token.Pos(1)},
			Op:    token.SIMPLE_ASSIGN,
			OpPos: token.Pos(3),
			Value: []ast.Expr{&ast.FuncCall{
				Dollar: token.Pos(6),
				Open:   token.LPAREN,
				Name:   &ast.Text{Value: "shell", ValuePos: token.Pos(8)},
				Args: []*ast.FuncArg{{
					From:  token.Pos(14),
					To:    token.Pos(17),
					Parts: []ast.Expr{&ast.Text{Value: "pwd", ValuePos: token.Pos(14)}},
				}},
				Close:    token.RPAREN,
				ClosePos: token.Pos(17),
			}},
		}))
	})

	It("should Parse a brace delimited function call", func() {
		buf := bytes.NewBufferString("A := ${shell pwd}")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		v, ok := f.Contents[0].(*ast.Variable)
		Expect(ok).To(BeTrue(), "expected a *ast.Variable, got %T", f.Contents[0])
		Expect(v.Value).To(ConsistOf(&ast.FuncCall{
			Dollar: token.Pos(6),
			Open:   token.LBRACE,
			Name:   &ast.Text{Value: "shell", ValuePos: token.Pos(8)},
			Args: []*ast.FuncArg{{
				From:  token.Pos(14),
				To:    token.Pos(17),
				Parts: []ast.Expr{&ast.Text{Value: "pwd", ValuePos: token.Pos(14)}},
			}},
			Close:    token.RBRACE,
			ClosePos: token.Pos(17),
		}))
	})

	DescribeTable("should Parse a call to every built-in function",
		Entry(nil, "subst"), Entry(nil, "patsubst"), Entry(nil, "strip"),
		Entry(nil, "findstring"), Entry(nil, "filter"), Entry(nil, "filter-out"),
		Entry(nil, "sort"), Entry(nil, "word"), Entry(nil, "words"),
		Entry(nil, "wordlist"), Entry(nil, "firstword"), Entry(nil, "lastword"),
		Entry(nil, "dir"), Entry(nil, "notdir"), Entry(nil, "suffix"),
		Entry(nil, "basename"), Entry(nil, "addsuffix"), Entry(nil, "addprefix"),
		Entry(nil, "join"), Entry(nil, "wildcard"), Entry(nil, "realpath"),
		Entry(nil, "abspath"), Entry(nil, "error"), Entry(nil, "warning"),
		Entry(nil, "shell"), Entry(nil, "origin"), Entry(nil, "flavor"),
		Entry(nil, "let"), Entry(nil, "foreach"), Entry(nil, "if"),
		Entry(nil, "or"), Entry(nil, "and"), Entry(nil, "intcmp"),
		Entry(nil, "call"), Entry(nil, "eval"), Entry(nil, "file"),
		Entry(nil, "value"),
		func(name string) {
			buf := bytes.NewBufferString("A := $(" + name + " x)")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			call := funcCall(f, 0)
			Expect(call.Name.Value).To(Equal(name))
			Expect(call.Args).To(HaveLen(1))
			Expect(call.Args[0].String()).To(Equal("x"))
		},
	)

	It("should Parse a nested function call", func() {
		buf := bytes.NewBufferString("A := $(patsubst %.c,%.o,$(wildcard *.c))")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		call := funcCall(f, 0)
		Expect(call.Name.Value).To(Equal("patsubst"))
		Expect(call.Commas).To(HaveExactElements(token.Pos(20), token.Pos(24)))
		Expect(call.Args).To(HaveLen(3))
		Expect(call.Args[2].Parts).To(ConsistOf(&ast.FuncCall{
			Dollar: token.Pos(25),
			Open:   token.LPAREN,
			Name:   &ast.Text{Value: "wildcard", ValuePos: token.Pos(27)},
			Args: []*ast.FuncArg{{
				From:  token.Pos(36),
				To:    token.Pos(39),
				Parts: []ast.Expr{&ast.Text{Value: "*.c", ValuePos: token.Pos(36)}},
			}},
			Close:    token.RPAREN,
			ClosePos: token.Pos(39),
		}))
	})

	DescribeTable("should Parse the arguments of a function call",
		Entry("a call with no arguments", "$(shell)", []string{}),
		Entry("a call with only whitespace", "$(shell )", []string{}),
		Entry("one argument", "$(shell pwd)", []string{"pwd"}),
		Entry("several arguments", "$(subst a,b,text)", []string{"a", "b", "text"}),
		Entry("an empty argument", "$(subst a,,text)", []string{"a", "", "text"}),
		Entry("a leading empty argument", "$(if ,x)", []string{"", "x"}),
		Entry("significant whitespace", "$(subst a, b,text)", []string{"a", " b", "text"}),
		Entry("trailing whitespace", "$(strip a b )", []string{"a b "}),
		Entry("a nested call holding a comma", "$(if $(findstring a,b),x,y)",
			[]string{"$(findstring a,b)", "x", "y"}),
		Entry("a comma nested in parentheses", "$(shell echo (a,b))", []string{"echo (a,b)"}),
		Entry("a comma past the argument count", "$(subst a,b,c,d)", []string{"a", "b", "c,d"}),
		Entry("a comma in a single argument function", "$(shell echo a,b)", []string{"echo a,b"}),
		Entry("any number of arguments", "$(call f,a,b,c)", []string{"f", "a", "b", "c"}),
		func(input string, args []string) {
			buf := bytes.NewBufferString("A := " + input)
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			call := funcCall(f, 0)
			Expect(call.Args).To(HaveLen(len(args)))
			for i, arg := range args {
				Expect(call.Args[i].String()).To(Equal(arg), "argument %d", i)
			}
			Expect(call.Commas).To(HaveLen(max(len(args)-1, 0)))
		},
	)

	DescribeTable("should Parse an expansion that is not a call as a reference",
		Entry("an ordinary name", "$(FOO)", "FOO"),
		Entry("a braced name", "${FOO}", "FOO"),
		Entry("a name matching a built-in followed by text", "$(shellfoo)", "shellfoo"),
		func(input, name string) {
			buf := bytes.NewBufferString("A := " + input)
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			v, ok := f.Contents[0].(*ast.Variable)
			Expect(ok).To(BeTrue(), "expected a *ast.Variable, got %T", f.Contents[0])
			ref, ok := v.Value[0].(*ast.VarRef)
			Expect(ok).To(BeTrue(), "expected a *ast.VarRef, got %T", v.Value[0])
			Expect(ref.Name).To(Equal(name))
		},
	)

	DescribeTable("should error when an expansion has no name",
		Entry(nil, "A := $()", "test:1:8: expected 'TEXT', found ')'"),
		Entry(nil, "A := ${}", "test:1:8: expected 'TEXT', found '}'"),
		Entry("a name that is only a comma", "A := $(,)",
			"test:1:8: expected 'TEXT', found ','"),
		Entry("a name that is only a blank", "A := $( )",
			"test:1:9: expected 'TEXT', found ')'"),
		func(input, msg string) {
			buf := bytes.NewBufferString(input)
			p := parser.New(buf, file)

			_, err := p.ParseFile()

			Expect(err).To(MatchError(msg))
		},
	)

	DescribeTable("should build no reference for an expansion with no name",
		Entry(nil, "A := $()"),
		Entry(nil, "A := ${}"),
		Entry("a name that is only a comma", "A := $(,)"),
		Entry("a name that is only a blank", "A := $( )"),
		func(input string) {
			buf := bytes.NewBufferString(input)
			p := parser.New(buf, file)

			f, _ := p.ParsePartial()

			v, ok := f.Contents[0].(*ast.Variable)
			Expect(ok).To(BeTrue(), "expected a *ast.Variable, got %T", f.Contents[0])
			Expect(v.Value).To(Equal([]ast.Expr{nil}))
			for n := range ast.Preorder(f) {
				Expect(n).NotTo(BeAssignableToTypeOf(&ast.VarRef{}))
			}
		},
	)

	DescribeTable("should recover from a malformed file without a fabricated node",
		Entry(nil, "a: $()"),
		Entry(nil, "A := $()"),
		Entry(nil, "A := ${}"),
		Entry(nil, "A := $(,)"),
		Entry(nil, "A := $( )"),
		Entry(nil, "A := $()x"),
		Entry(nil, "A := $(shell pwd"),
		Entry(nil, "A := ${shell pwd"),
		Entry(nil, "A := $"),
		Entry(nil, "foo = $"),
		Entry(nil, "$"),
		Entry(nil, "ifdef"),
		Entry(nil, "ifdef\n"),
		Entry(nil, "ifdef foo"),
		Entry(nil, "ifeq \"a\" b"),
		Entry(nil, "ifeq a \"b\""),
		Entry(nil, "ifeq"),
		func(input string) {
			buf := bytes.NewBufferString(input)
			p := parser.New(buf, file)

			f, _ := p.ParsePartial()

			Expect(f).NotTo(BeNil())
			for n := range ast.Preorder(f) {
				Expect(n.Pos().IsValid()).To(BeTrue(),
					"%T has an invalid Pos()", n)
			}
			Expect(func() {
				_, _ = printer.Fprint(io.Discard, f)
			}).NotTo(Panic())
		},
	)

	DescribeTable("should Parse a name make does not know as a call",
		Entry("a name separated from its arguments", "$(foo bar)", "foo"),
		Entry("a built-in with no arguments", "$(dir)", "dir"),
		Entry("a logging function", "$(info msg)", "info"),
		func(input, name string) {
			buf := bytes.NewBufferString("A := " + input)
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(funcCall(f, 0).Name.Value).To(Equal(name))
		},
	)

	DescribeTable("should error when a function call has no closing delimiter",
		Entry(nil, "A := $(shell pwd", "test:1:17: expected ')', found 'EOF'"),
		Entry(nil, "A := ${shell pwd", "test:1:17: expected '}', found 'EOF'"),
		Entry("a mismatched closing delimiter", "A := $(shell pwd}",
			"test:1:18: expected ')', found 'EOF'"),
		func(input, msg string) {
			buf := bytes.NewBufferString(input)
			p := parser.New(buf, file)

			_, err := p.ParseFile()

			Expect(err).To(MatchError(msg))
		},
	)

	Describe("a conditional directive inside a rule body", func() {
		It("should attach the conditional to the rule", func() {
			buf := bytes.NewBufferString("target:\nifdef VERBOSE\n\techo building\nendif\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(ConsistOf(&ast.Rule{
				Targets: []ast.Expr{&ast.Text{
					Value:    "target",
					ValuePos: token.Pos(1),
				}},
				Colon:        token.Pos(7),
				PreReqs:      []ast.Expr{},
				OrderPreReqs: []ast.Expr{},
				Recipes: []ast.RecipeObj{&ast.IfBlock{
					Directive: &ast.IfdefDir{
						Tok:    token.IFDEF,
						TokPos: token.Pos(9),
						VarName: &ast.Text{
							Value:    "VERBOSE",
							ValuePos: token.Pos(15),
						},
					},
					Text: []ast.Obj{&ast.Recipe{
						Prefix:    token.TAB,
						PrefixPos: token.Pos(23),
						Text: ast.Text{
							Value:    "echo building",
							ValuePos: token.Pos(24),
						},
					}},
					Endif: token.Pos(38),
				}},
			}))
		})

		It("should keep the recipes written around the conditional in order", func() {
			buf := bytes.NewBufferString("target:\n\tone\nifdef V\n\ttwo\nendif\n\tthree\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(HaveLen(1))
			rule, ok := f.Contents[0].(*ast.Rule)
			Expect(ok).To(BeTrue(), "expected a *ast.Rule, got %T", f.Contents[0])
			Expect(rule.Recipes).To(HaveLen(3))
			Expect(rule.Recipes[0]).To(HaveField("Value", "one"))
			Expect(rule.Recipes[1]).To(BeAssignableToTypeOf(&ast.IfBlock{}))
			Expect(rule.Recipes[2]).To(HaveField("Value", "three"))
		})

		It("should attach the recipes of every branch to the rule", func() {
			buf := bytes.NewBufferString(
				"target:\nifdef V\n\tone\nelse ifdef W\n\ttwo\nelse\n\tthree\nendif\n",
			)
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			block := recipeIfBlock(f, 0, 0)
			Expect(block.Text).To(HaveLen(1))
			Expect(block.Text[0]).To(HaveField("Value", "one"))
			Expect(block.Else).To(HaveLen(2))
			Expect(block.Else[0].Text[0]).To(HaveField("Value", "two"))
			Expect(block.Else[1].Text[0]).To(HaveField("Value", "three"))
		})

		It("should attach a conditional nested in another one", func() {
			buf := bytes.NewBufferString("target:\nifdef V\nifdef W\n\tone\nendif\nendif\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			outer := recipeIfBlock(f, 0, 0)
			Expect(outer.Text).To(HaveLen(1))
			inner, ok := outer.Text[0].(*ast.IfBlock)
			Expect(ok).To(BeTrue(), "expected a *ast.IfBlock, got %T", outer.Text[0])
			Expect(inner.Text[0]).To(HaveField("Value", "one"))
		})

		It("should read a recipe introduced by a custom prefix", func() {
			buf := bytes.NewBufferString(".RECIPEPREFIX = >\ntarget:\nifdef V\n>one\nendif\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			block := recipeIfBlock(f, 1, 0)
			Expect(block.Text).To(ConsistOf(&ast.Recipe{
				Prefix:    token.TEXT,
				PrefixLit: ">",
				PrefixPos: token.Pos(35),
				Text: ast.Text{
					Value:    "one",
					ValuePos: token.Pos(36),
				},
			}))
		})

		// A conditional that holds no recipe line wraps ordinary make syntax,
		// so it stays beside the rule the way it always has.
		It("should leave a conditional holding a rule at the level it was written", func() {
			buf := bytes.NewBufferString("first:\nifdef V\nsecond:\nendif\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(HaveLen(2))
			Expect(f.Contents[0]).To(BeAssignableToTypeOf(&ast.Rule{}))
			Expect(f.Contents[1]).To(BeAssignableToTypeOf(&ast.IfBlock{}))
			Expect(f.Contents[0].(*ast.Rule).Recipes).To(BeEmpty())
		})

		It("should leave a prefixed line written before any target a bad object", func() {
			buf := bytes.NewBufferString("ifdef V\n\torphan\nendif\n")
			p := parser.New(buf, file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Contents).To(HaveLen(1))
			block, ok := f.Contents[0].(*ast.IfBlock)
			Expect(ok).To(BeTrue(), "expected a *ast.IfBlock, got %T", f.Contents[0])
			Expect(block.Text).To(ConsistOf(&ast.BadObj{
				From: token.Pos(9),
				To:   token.Pos(16),
				Text: "\torphan",
			}))
		})
	})

	DescribeTable("should record the line ending of the file",
		func(input, expected string) {
			p := parser.New(bytes.NewBufferString(input), file)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.LineEnding).To(Equal(expected))
		},
		Entry("LF", "a: b\nc: d\n", "\n"),
		Entry("CRLF", "a: b\r\nc: d\r\n", "\r\n"),
		Entry("a single CRLF line", "a: b\r\n", "\r\n"),
		Entry("no line ending at all", "a: b", "\n"),
		Entry("mostly CRLF", "a: b\r\nc: d\r\ne: f\n", "\r\n"),
		Entry("mostly LF", "a: b\nc: d\ne: f\r\n", "\n"),
	)
})

// recipeIfBlock returns the jth entry of the recipe list of the ith object in
// f, failing the spec when the object is not a rule holding a conditional
// there.
func recipeIfBlock(f *ast.File, i, j int) *ast.IfBlock {
	GinkgoHelper()
	r, ok := f.Contents[i].(*ast.Rule)
	Expect(ok).To(BeTrue(), "expected a *ast.Rule, got %T", f.Contents[i])
	Expect(r.Recipes).To(HaveLen(j + 1))
	block, ok := r.Recipes[j].(*ast.IfBlock)
	Expect(ok).To(BeTrue(), "expected a *ast.IfBlock, got %T", r.Recipes[j])

	return block
}

// funcCall returns the call assigned by the ith object in f, failing the spec
// when the object is not a variable holding one.
func funcCall(f *ast.File, i int) *ast.FuncCall {
	GinkgoHelper()
	v, ok := f.Contents[i].(*ast.Variable)
	Expect(ok).To(BeTrue(), "expected a *ast.Variable, got %T", f.Contents[i])
	Expect(v.Value).NotTo(BeEmpty())
	call, ok := v.Value[0].(*ast.FuncCall)
	Expect(ok).To(BeTrue(), "expected a *ast.FuncCall, got %T", v.Value[0])

	return call
}
