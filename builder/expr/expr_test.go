package expr_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/expr"
	"github.com/unmango/go-make/builder/text"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Expr", func() {
	Describe("Copy", func() {
		It("shoudl copy text", func() {
			t := text.New(1)

			actual := expr.Copy(2, t)

			Expect(actual).To(Equal(text.Copy(2, t)))
		})

		DescribeTable("should copy an expression to pos",
			func(e ast.Expr, expected ast.Expr) {
				actual := expr.Copy(2, e)

				Expect(actual).NotTo(BeIdenticalTo(e))
				Expect(actual).To(Equal(expected))
			},
			Entry("text",
				&ast.Text{Value: "test", ValuePos: 1},
				&ast.Text{Value: "test", ValuePos: 2},
			),
			Entry("a var ref",
				&ast.VarRef{Dollar: 1, Open: token.LPAREN, Name: "FOO", Close: token.RPAREN},
				&ast.VarRef{Dollar: 2, Open: token.LPAREN, Name: "FOO", Close: token.RPAREN},
			),
			Entry("a single character var ref",
				&ast.VarRef{Dollar: 1, Open: token.ILLEGAL, Name: "b", Close: token.ILLEGAL},
				&ast.VarRef{Dollar: 2, Open: token.ILLEGAL, Name: "b", Close: token.ILLEGAL},
			),
			Entry("a quoted expression",
				&ast.QuotedExpr{
					Quote: token.QUOTE,
					Open:  1,
					Value: &ast.Text{Value: "test", ValuePos: 2},
					Close: 6,
				},
				&ast.QuotedExpr{
					Quote: token.QUOTE,
					Open:  2,
					Value: &ast.Text{Value: "test", ValuePos: 3},
					Close: 7,
				},
			),
			Entry("a recipe",
				&ast.Recipe{Text: ast.Text{Value: "test", ValuePos: 2}, Prefix: token.TAB, PrefixPos: 1},
				&ast.Recipe{Text: ast.Text{Value: "test", ValuePos: 3}, Prefix: token.TAB, PrefixPos: 2},
			),
		)

		It("should not alias the value of a quoted expression", func() {
			e := &ast.QuotedExpr{Quote: token.APOS, Value: &ast.Text{Value: "test"}}

			actual := expr.Copy(2, e).(*ast.QuotedExpr)

			Expect(actual.Value).NotTo(BeIdenticalTo(e.Value))
			Expect(e.Value.Pos()).To(Equal(token.NoPos))
		})
	})

	DescribeTable("End",
		func(e ast.Expr, expected token.Pos) {
			Expect(expr.End(e)).To(Equal(expected))
		},
		Entry("text", &ast.Text{Value: "test", ValuePos: 1}, token.Pos(5)),
		Entry("a var ref",
			&ast.VarRef{Dollar: 1, Open: token.LBRACE, Name: "FOO", Close: token.RBRACE},
			token.Pos(7), // ${FOO}
		),
		Entry("a single character var ref",
			&ast.VarRef{Dollar: 1, Open: token.ILLEGAL, Name: "b", Close: token.ILLEGAL},
			token.Pos(3), // $b
		),
		Entry("a quoted expression",
			&ast.QuotedExpr{
				Quote: token.APOS,
				Open:  1,
				Value: &ast.Text{Value: "a", ValuePos: 2},
				Close: 3,
			},
			token.Pos(4), // 'a'
		),
		Entry("a recipe",
			&ast.Recipe{Text: ast.Text{Value: "test", ValuePos: 2}, Prefix: token.TAB, PrefixPos: 1},
			token.Pos(6),
		),
	)

	Describe("SetPos", func() {
		It("should move the expression in place", func() {
			e := &ast.Text{Value: "test", ValuePos: 1}

			end := expr.SetPos(4, e)

			Expect(e.ValuePos).To(Equal(token.Pos(4)))
			Expect(end).To(Equal(token.Pos(8)))
		})
	})
})
