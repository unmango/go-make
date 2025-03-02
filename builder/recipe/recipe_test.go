package recipe_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/recipe"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Recipe", func() {
	Describe("Copy", func() {
		It("should work", func() {
			r := &ast.Recipe{
				Text:      ast.Text{Value: "test"},
				Prefix:    token.TAB,
				PrefixPos: 1,
			}

			actual := recipe.Copy(2, r)

			Expect(actual.Prefix).To(Equal(r.Prefix))
			Expect(actual.PrefixPos).To(Equal(token.Pos(2)))
			Expect(actual.Text).To(Equal(ast.Text{
				Value:    "test",
				ValuePos: 3,
			}))
		})
	})
})
