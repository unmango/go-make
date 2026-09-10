package writer_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/writer"
)

var _ = Describe("Writer", func() {
	Describe("New", func() {
		It("should re-use an existing *writer.Writer", func() {
			a := writer.New(nil)

			b := writer.New(a)

			Expect(a).To(BeIdenticalTo(b))
		})
	})

	It("should write a line", func() {
		buf := &bytes.Buffer{}
		w := writer.New(buf)

		n, err := w.WriteLine()

		Expect(err).NotTo(HaveOccurred())
		Expect(buf.String()).To(Equal("\n"))
		Expect(n).To(Equal(1))
	})
})
