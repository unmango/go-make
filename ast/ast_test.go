package ast_test

import (
	"testing/quick"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Ast", func() {
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
			c := &ast.Rule{Targets: &ast.TargetList{
				List: []ast.Expr{&ast.Text{
					ValuePos: token.Pos(69),
				}},
			}}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position after the final recipe", func() {
			c := &ast.Rule{Recipes: []*ast.Recipe{{
				TokPos: token.Pos(420),
				Text:   "some text",
			}}}

			Expect(c.End()).To(Equal(token.Pos(429)))
		})
	})

	Describe("TargetList", func() {
		It("should return the position of the first target", func() {
			c := &ast.TargetList{
				List: []ast.Expr{&ast.Text{
					ValuePos: token.Pos(69),
				}},
			}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position of the last target", func() {
			c := &ast.TargetList{List: []ast.Expr{
				&ast.Text{ValuePos: token.Pos(69)},
				&ast.Text{
					ValuePos: token.Pos(420),
					Value:    "foo",
				},
			}}

			Expect(c.End()).To(Equal(token.Pos(423)))
		})

		It("should append the given target", func() {
			c := &ast.TargetList{}
			elem := &ast.Text{
				ValuePos: token.Pos(69),
			}

			c.Add(elem)

			Expect(c.List).To(ContainElement(elem))
		})
	})

	Describe("PreReqList", func() {
		It("should return the position of the first target", func() {
			c := &ast.PreReqList{
				List: []ast.Expr{&ast.Text{
					ValuePos: token.Pos(69),
				}},
			}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position after the last prereq", func() {
			c := &ast.PreReqList{List: []ast.Expr{
				&ast.Text{ValuePos: token.Pos(69)},
				&ast.Text{
					ValuePos: token.Pos(420),
					Value:    "baz",
				},
			}}

			Expect(c.End()).To(Equal(token.Pos(423)))
		})

		It("should append the given prereq", func() {
			c := &ast.PreReqList{}
			elem := &ast.Text{
				ValuePos: token.Pos(69),
			}

			c.Add(elem)

			Expect(c.List).To(ContainElement(elem))
		})
	})

	Describe("Text", func() {
		It("should return the position of the identifier", func() {
			c := &ast.Text{
				ValuePos: token.Pos(69),
			}

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
			c := &ast.Text{
				Value: "foo",
			}

			Expect(c.String()).To(Equal("foo"))
		})
	})

	Describe("Recipe", func() {
		It("should return the position of the tab", func() {
			c := &ast.Recipe{
				TokPos: token.Pos(420),
			}

			Expect(c.Pos()).To(Equal(token.Pos(420)))
		})

		It("should return the position after the text", func() {
			c := &ast.Recipe{
				TokPos: token.Pos(420),
				Tok:    token.TAB,
				Text:   "foo",
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
