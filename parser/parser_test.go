package parser_test

import (
	"bytes"
	gotoken "go/token"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/parser"
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
			List: []*ast.Comment{{Pound: token.Pos(1), Text: "comment text"}},
		}))
	})

	It("should Parse a comment group", func() {
		buf := bytes.NewBufferString("# comment text\n# more text on this line")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.CommentGroup{
			List: []*ast.Comment{
				{Pound: token.Pos(1), Text: "comment text"},
				{Pound: token.Pos(16), Text: "more text on this line"},
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
				{Pound: token.Pos(1), Text: "comment text"},
			}},
			&ast.CommentGroup{List: []*ast.Comment{
				{Pound: token.Pos(17), Text: "new comment group"},
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
			Recipes:      []*ast.Recipe{},
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
				Recipes:      []*ast.Recipe{},
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
			Recipes:      []*ast.Recipe{},
		}))
	})

	It("should Parse a target with a single character variable reference and extra text", func() {
		buf := bytes.NewBufferString("$foo:")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(5),
			Targets: []ast.Expr{
				&ast.VarRef{
					Dollar: token.Pos(1),
					Open:   token.ILLEGAL,
					Name:   "f",
					Close:  token.ILLEGAL,
				},
				&ast.Text{Value: "oo", ValuePos: token.Pos(3)},
			},
			PreReqs:      []ast.Expr{},
			OrderPreReqs: []ast.Expr{},
			Recipes:      []*ast.Recipe{},
		}))
	})

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
			Recipes:      []*ast.Recipe{},
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
			Recipes:      []*ast.Recipe{},
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
			Recipes:      []*ast.Recipe{},
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
			Recipes:      []*ast.Recipe{},
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
			Recipes: []*ast.Recipe{},
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
			Recipes: []*ast.Recipe{{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(9),
				Text: ast.Text{
					Value:    "recipe",
					ValuePos: token.Pos(10),
				},
			}},
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
			Recipes: []*ast.Recipe{
				{
					Prefix:    token.TAB,
					PrefixPos: token.Pos(9),
					Text: ast.Text{
						Value:    "recipe",
						ValuePos: token.Pos(10),
					},
				},
				{
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
			Recipes: []*ast.Recipe{{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(9),
				Text: ast.Text{
					Value:    "recipe part2",
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
			Recipes: []*ast.Recipe{{
				Prefix:    token.TAB,
				PrefixPos: token.Pos(16),
				Text: ast.Text{
					Value:    "recipe",
					ValuePos: token.Pos(17),
				},
			}},
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
		Entry(nil, "VAR ?= test test2", token.IFNDEF_ASSIGN, 8),
		Entry(nil, "VAR = test test2", token.RECURSIVE_ASSIGN, 7),
	)

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
		Entry(nil, "VAR ?=", token.IFNDEF_ASSIGN),
		Entry(nil, "VAR =", token.RECURSIVE_ASSIGN),
	)

	It("should Parse a conditional directive", func() {
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
				Recipes:      []*ast.Recipe{},
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
					Recipes:      []*ast.Recipe{},
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
					Recipes:      []*ast.Recipe{},
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

	It("should Parse a conditional directive", Pending, func() {
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
						Arg: &ast.Text{
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
						OpPos: token.Pos(49),
					}},
				},
				{
					Else: token.Pos(52),
					Text: []ast.Obj{&ast.Variable{
						Name: &ast.Text{
							Value:    "BAR",
							ValuePos: token.Pos(57),
						},
						Op:    token.IMMEDIATE_ASSIGN,
						OpPos: token.Pos(61),
					}},
				},
			},
			Endif: token.Pos(66),
		}))
	})

	It("should error with extra text to the left of the assignment", func() {
		buf := bytes.NewBufferString("VAR invalid :=")
		s := parser.New(buf, file)

		_, err := s.ParseFile()

		Expect(err).To(MatchError("test:1:13: variable may have only one name"))
	})
})
