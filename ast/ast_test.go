package ast_test

import (
	"go/token"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
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
})
