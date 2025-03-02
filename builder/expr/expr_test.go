package expr_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/builder/expr"
	"github.com/unmango/go-make/builder/text"
)

var _ = Describe("Expr", func() {
	Describe("Copy", func() {
		It("shoudl copy text", func() {
			t := text.New(1)

			actual := expr.Copy(2, t)

			Expect(actual).To(Equal(text.Copy(2, t)))
		})
	})
})
