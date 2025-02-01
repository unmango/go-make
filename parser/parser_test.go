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

	It("should Parse a target", func() {
		buf := bytes.NewBufferString("target:")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
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
			Expect(f.Decls).To(ConsistOf(&ast.Rule{
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
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
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
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
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
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
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
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
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
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
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

	It("should Parse a target with an order-only prereq", func() {
		buf := bytes.NewBufferString("target: | prereq")
		p := parser.New(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
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
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
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
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
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
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
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
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
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
		Expect(f.Decls).NotTo(BeEmpty())
	})

	DescribeTable("should parse a variable definition",
		func(input string, op token.Token, vpos int) {
			buf := bytes.NewBufferString(input)
			s := parser.New(buf, file)

			f, err := s.ParseFile()

			Expect(err).NotTo(HaveOccurred())
			Expect(f.Decls).To(ConsistOf(&ast.Variable{
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
			Expect(f.Decls).To(ConsistOf(&ast.Variable{
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
			Expect(f.Decls).To(ConsistOf(&ast.Variable{
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
})
