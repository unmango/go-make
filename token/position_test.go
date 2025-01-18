package token_test

import (
	gotoken "go/token"
	"math"
	"testing/quick"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/token"
)

var _ = Describe("Position", func() {
	Describe("PositionFor", func() {
		var file *token.File

		BeforeEach(func() {
			file = gotoken.NewFileSet().AddFile("test", 1, math.MaxInt-2)
		})

		It("should be equivalent to calling file.PositionFor(p, false)", func() {
			err := quick.Check(func(p int) bool {
				pos := token.Pos(p)

				expected := file.PositionFor(pos, false)
				actual := token.PositionFor(file, pos)

				return actual == expected
			}, nil)

			Expect(err).NotTo(HaveOccurred())
		})
	})
})
