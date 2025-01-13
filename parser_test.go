package make_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
)

var _ = Describe("Parser", func() {
	It("should Parse", func() {
		buf := bytes.NewBufferString("target:")
		p := make.NewParser(buf)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f).NotTo(BeNil())
	})

	It("should error when starting at a colon", func() {
		buf := bytes.NewBufferString(":")
		p := make.NewParser(buf)

		_, err := p.ParseFile()

		Expect(err).To(MatchError("expected 'IDENT'"))
	})
})
