package make_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
)

var _ = Describe("Parser", func() {
	It("should Parse", func() {
		p := make.NewParser(&bytes.Buffer{})

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f).NotTo(BeNil())
	})
})
