package expr_test

import (
	"go/token"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/builder/expr"
	"github.com/unmango/go-make/builder/text"
)

var _ = Describe("Expr", func() {
	Describe("RePos", func() {
		It("shoudl reposition text", func() {
			t := text.New(1)
			expr.RePos(2, t)

			Expect(t.Pos()).To(Equal(token.Pos(2)))
		})
	})
})
