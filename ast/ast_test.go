package ast_test

import (
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
				List: []ast.FileName{&ast.LiteralFileName{
					Name: &ast.Ident{NamePos: token.Pos(69)},
				}},
			}}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position after the final recipe", func() {
			c := &ast.Rule{Recipes: []*ast.Recipe{{
				TokPos: token.Pos(420),
			}}}

			// TODO: This is wrong, should be position after text
			Expect(c.End()).To(Equal(token.Pos(420)))
		})
	})

	Describe("TargetList", func() {
		It("should return the position of the first target", func() {
			c := &ast.TargetList{
				List: []ast.FileName{&ast.LiteralFileName{
					Name: &ast.Ident{NamePos: token.Pos(69)},
				}},
			}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position of the last target", func() {
			c := &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{NamePos: token.Pos(69)}},
				&ast.LiteralFileName{Name: &ast.Ident{
					NamePos: token.Pos(420),
					Name:    "foo",
				}},
			}}

			Expect(c.End()).To(Equal(token.Pos(423)))
		})
	})

	Describe("PreReqList", func() {
		It("should return the position of the first target", func() {
			c := &ast.PreReqList{
				List: []ast.FileName{&ast.LiteralFileName{
					Name: &ast.Ident{NamePos: token.Pos(69)},
				}},
			}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position after the lat prereq", func() {
			c := &ast.PreReqList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{NamePos: token.Pos(69)}},
				&ast.LiteralFileName{Name: &ast.Ident{
					NamePos: token.Pos(420),
					Name:    "baz",
				}},
			}}

			Expect(c.End()).To(Equal(token.Pos(423)))
		})
	})

	Describe("LiteralFileName", func() {
		It("should return the position of the identifier", func() {
			c := &ast.LiteralFileName{Name: &ast.Ident{
				NamePos: token.Pos(69),
			}}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position after the identifier", func() {
			c := &ast.LiteralFileName{Name: &ast.Ident{
				NamePos: token.Pos(420),
				Name:    "bar",
			}}

			Expect(c.End()).To(Equal(token.Pos(423)))
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

	Describe("Ident", func() {
		It("should return the position of the name", func() {
			c := &ast.Ident{
				NamePos: token.Pos(69),
			}

			Expect(c.Pos()).To(Equal(token.Pos(69)))
		})

		It("should return the position after the name", func() {
			c := &ast.Ident{
				NamePos: token.Pos(420),
				Name:    "foo",
			}

			Expect(c.End()).To(Equal(token.Pos(423)))
		})
	})
})
