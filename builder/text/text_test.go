package text_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/builder/text"
	"github.com/unmango/go-make/token"
)

var _ = Describe("Text", func() {
	Describe("New", func() {
		It("Should work", func() {
			t := text.New(1)

			Expect(t.Pos()).To(Equal(token.Pos(1)))
		})
	})
})
