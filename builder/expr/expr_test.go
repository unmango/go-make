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
			Entry("an empty quoted expression",
				&ast.QuotedExpr{Quote: token.QUOTE, Open: 1, Close: 2},
				&ast.QuotedExpr{Quote: token.QUOTE, Open: 2, Close: 3},
			),
			Entry("a recipe",
				&ast.Recipe{Text: ast.Text{Value: "test", ValuePos: 2}, Prefix: token.TAB, PrefixPos: 1},
				&ast.Recipe{Text: ast.Text{Value: "test", ValuePos: 3}, Prefix: token.TAB, PrefixPos: 2},
			),
			Entry("a custom prefix recipe",
				&ast.Recipe{Text: ast.Text{Value: "test", ValuePos: 2}, Prefix: token.TEXT, PrefixLit: ">", PrefixPos: 1},
				&ast.Recipe{Text: ast.Text{Value: "test", ValuePos: 3}, Prefix: token.TEXT, PrefixLit: ">", PrefixPos: 2},
			),
			Entry("a function call",
				// "$(shell pwd)"
				&ast.FuncCall{
					Dollar: 1,
					Open:   token.LPAREN,
					Name:   &ast.Text{Value: "shell", ValuePos: 3},
					Args: []*ast.FuncArg{{
						From:  9,
						To:    12,
						Parts: []ast.Expr{&ast.Text{Value: "pwd", ValuePos: 9}},
					}},
					Close:    token.RPAREN,
					ClosePos: 12,
				},
				&ast.FuncCall{
					Dollar: 2,
					Open:   token.LPAREN,
					Name:   &ast.Text{Value: "shell", ValuePos: 4},
					Args: []*ast.FuncArg{{
						From:  10,
						To:    13,
						Parts: []ast.Expr{&ast.Text{Value: "pwd", ValuePos: 10}},
					}},
					Close:    token.RPAREN,
					ClosePos: 13,
				},
			),
			Entry("a juxtaposed expression",
				// "$(FOO)bar"
				&ast.JuxtaposedExpr{Parts: []ast.Expr{
					&ast.VarRef{Dollar: 1, Open: token.LPAREN, Name: "FOO", Close: token.RPAREN},
					&ast.Text{Value: "bar", ValuePos: 7},
				}},
				&ast.JuxtaposedExpr{Parts: []ast.Expr{
					&ast.VarRef{Dollar: 2, Open: token.LPAREN, Name: "FOO", Close: token.RPAREN},
					&ast.Text{Value: "bar", ValuePos: 8},
				}},
			),
			Entry("an empty juxtaposed expression",
				&ast.JuxtaposedExpr{Parts: []ast.Expr{}},
				&ast.JuxtaposedExpr{Parts: []ast.Expr{}},
			),
		)

		It("should not alias the parts of a juxtaposed expression", func() {
			part := &ast.Text{Value: "bar"}
			e := &ast.JuxtaposedExpr{Parts: []ast.Expr{part}}

			actual := expr.Copy(2, e).(*ast.JuxtaposedExpr)

			Expect(actual.Parts[0]).NotTo(BeIdenticalTo(part))
			Expect(part.Pos()).To(Equal(token.NoPos))
		})

		It("should not alias the arguments of a function call", func() {
			part := &ast.Text{Value: "pwd"}
			e := &ast.FuncCall{
				Open:  token.LPAREN,
				Name:  &ast.Text{Value: "shell"},
				Args:  []*ast.FuncArg{{Parts: []ast.Expr{part}}},
				Close: token.RPAREN,
			}

			actual := expr.Copy(1, e).(*ast.FuncCall)

			Expect(actual.Name).NotTo(BeIdenticalTo(e.Name))
			Expect(actual.Args[0]).NotTo(BeIdenticalTo(e.Args[0]))
			Expect(actual.Args[0].Parts[0]).NotTo(BeIdenticalTo(part))
			Expect(part.Pos()).To(Equal(token.NoPos))
		})

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
		Entry("an empty quoted expression",
			&ast.QuotedExpr{Quote: token.APOS, Open: 1, Close: 2},
			token.Pos(3), // ''
		),
		Entry("a recipe",
			&ast.Recipe{Text: ast.Text{Value: "test", ValuePos: 2}, Prefix: token.TAB, PrefixPos: 1},
			token.Pos(6),
		),
		Entry("a function call",
			&ast.FuncCall{Close: token.RPAREN, ClosePos: 12},
			token.Pos(13), // $(shell pwd)
		),
		Entry("a juxtaposed expression",
			&ast.JuxtaposedExpr{Parts: []ast.Expr{
				&ast.VarRef{Dollar: 1, Open: token.LPAREN, Name: "FOO", Close: token.RPAREN},
				&ast.Text{Value: "bar", ValuePos: 7},
			}},
			token.Pos(10), // $(FOO)bar
		),
		Entry("an empty juxtaposed expression",
			&ast.JuxtaposedExpr{},
			token.NoPos,
		),
	)

	Describe("SetPos", func() {
		It("should move the expression in place", func() {
			e := &ast.Text{Value: "test", ValuePos: 1}

			end := expr.SetPos(4, e)

			Expect(e.ValuePos).To(Equal(token.Pos(4)))
			Expect(end).To(Equal(token.Pos(8)))
		})

		It("should lay the parts of a juxtaposition out back to back", func() {
			e := &ast.JuxtaposedExpr{Parts: []ast.Expr{
				&ast.VarRef{Open: token.LPAREN, Name: "FOO", Close: token.RPAREN},
				&ast.Text{Value: "bar"},
			}}

			end := expr.SetPos(1, e)

			// "$(FOO)bar"
			Expect(e.Parts[0].Pos()).To(Equal(token.Pos(1)))
			Expect(e.Parts[1].Pos()).To(Equal(token.Pos(7)))
			Expect(end).To(Equal(token.Pos(10)))
		})

		It("should leave an empty juxtaposition where it was placed", func() {
			e := &ast.JuxtaposedExpr{}

			end := expr.SetPos(4, e)

			Expect(end).To(Equal(token.Pos(4)))
		})

		It("should lay a function call out canonically", func() {
			e := &ast.FuncCall{
				Open: token.LPAREN,
				Name: &ast.Text{Value: "subst"},
				Args: []*ast.FuncArg{
					{Parts: []ast.Expr{&ast.Text{Value: "a"}}},
					{Parts: []ast.Expr{&ast.Text{Value: "b"}}},
					{Parts: []ast.Expr{&ast.Text{Value: "c"}}},
				},
				Close: token.RPAREN,
			}

			end := expr.SetPos(1, e)

			// "$(subst a,b,c)"
			Expect(e.Name.ValuePos).To(Equal(token.Pos(3)))
			Expect(e.Commas).To(HaveExactElements(token.Pos(10), token.Pos(12)))
			Expect(e.Args[1].From).To(Equal(token.Pos(11)))
			Expect(e.ClosePos).To(Equal(token.Pos(14)))
			Expect(end).To(Equal(token.Pos(15)))
		})
	})
})
