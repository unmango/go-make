package make_test

import (
	"bytes"
	gotoken "go/token"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Parser", func() {
	var file *token.File

	BeforeEach(func() {
		file = gotoken.NewFileSet().AddFile("test", 1, math.MaxInt-2)
	})

	It("should Parse a target", func() {
		buf := bytes.NewBufferString("target:")
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{},
			Recipes: []*ast.Recipe{},
		}))
	})

	It("should Parse a rule with multiple targets", func() {
		buf := bytes.NewBufferString("target target2:")
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(15),
			Targets: []ast.Expr{
				&ast.Text{Value: "target", ValuePos: token.Pos(1)},
				&ast.Text{Value: "target2", ValuePos: token.Pos(8)},
			},
			PreReqs: []ast.Expr{},
			Recipes: []*ast.Recipe{},
		}))
	})

	It("should Parse a target with a prereq", func() {
		buf := bytes.NewBufferString("target: prereq")
		p := make.NewParser(buf, file)

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
			Recipes: []*ast.Recipe{},
		}))
	})

	It("should Parse a target with multiple prereqs", func() {
		buf := bytes.NewBufferString("target: prereq prereq2")
		p := make.NewParser(buf, file)

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
			Recipes: []*ast.Recipe{},
		}))
	})

	It("should Parse a target with a recipe", func() {
		buf := bytes.NewBufferString("target:\n\trecipe")
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{},
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
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{},
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
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Decls).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: []ast.Expr{&ast.Text{
				Value:    "target",
				ValuePos: token.Pos(1),
			}},
			PreReqs: []ast.Expr{},
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
		p := make.NewParser(buf, file)

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
		s := make.NewParser(buf, nil)

		f, err := s.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Decls).NotTo(BeEmpty())
	})

	DescribeTable("should parse a variable definition",
		func(input string, op token.Token, vpos int) {
			buf := bytes.NewBufferString(input)
			s := make.NewParser(buf, file)

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
			s := make.NewParser(buf, file)

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
			s := make.NewParser(buf, file)

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
