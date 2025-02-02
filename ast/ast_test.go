package ast_test

import (
	"testing/quick"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
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

	Describe("CommentGroup", func() {
		It("should return the position of the first comment", func() {
			c := &ast.CommentGroup{[]*ast.Comment{{
				Pound: token.Pos(69),
			}}}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position of the last comment", func() {
			c := &ast.CommentGroup{[]*ast.Comment{
				{Pound: token.Pos(69), Text: "foo"},
				{Pound: token.Pos(420), Text: "Some comment text"},
			}}

			Expect(c.End()).To(Equal(token.Pos(437)))
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
				Text:  "Some comment text",
			}

			Expect(c.End()).To(Equal(token.Pos(437)))
		})
	})

	Describe("Rule", func() {
		It("should return the position of the first target", func() {
			c := &ast.Rule{Targets: []ast.Expr{
				&ast.Text{ValuePos: token.Pos(69)},
			}}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position after the final recipe", func() {
			c := &ast.Rule{Recipes: []*ast.Recipe{{
				PrefixPos: token.Pos(420),
				Text:      ast.Text{Value: "some text"},
			}}}

			Expect(c.End()).To(Equal(token.Pos(429)))
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

			Expect(c.End()).To(Equal(token.Pos(425)))
		})

		It("should return the position after the character", func() {
			c := &ast.VarRef{
				Dollar: token.Pos(420),
				Open:   token.ILLEGAL,
				Name:   "b",
				Close:  token.ILLEGAL,
			}

			Expect(c.End()).To(Equal(token.Pos(421)))
		})

		It("should stringify with parens", func() {
			c := &ast.VarRef{
				Dollar: token.Pos(1),
				Open:   token.LPAREN,
				Name:   "foo",
				Close:  token.RPAREN,
			}

			Expect(c.String()).To(Equal("$(foo)"))
		})

		It("should stringify with braces", func() {
			c := &ast.VarRef{
				Dollar: token.Pos(1),
				Open:   token.LBRACE,
				Name:   "foo",
				Close:  token.RBRACE,
			}

			Expect(c.String()).To(Equal("${foo}"))
		})

		It("should stringify single characters", func() {
			c := &ast.VarRef{
				Dollar: token.Pos(1),
				Open:   token.ILLEGAL,
				Name:   "f",
				Close:  token.ILLEGAL,
			}

			Expect(c.String()).To(Equal("$f"))
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

			Expect(c.End()).To(Equal(token.Pos(423)))
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
				func(tok token.Token, l int) {
					err := quick.Check(func(n int) bool {
						v := &ast.Variable{
							Name: &ast.Text{ValuePos: token.Pos(n)},
							Op:   tok,
						}

						return v.End() == token.Pos(n+l)
					}, nil)

					Expect(err).NotTo(HaveOccurred())
				},
			)
		})
	})
})
