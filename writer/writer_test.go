package writer_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/writer"
)

var _ = Describe("Writer", func() {
	It("should write a line", func() {
		buf := &bytes.Buffer{}
		w := writer.New(buf)

		n, err := w.WriteLine()

		Expect(err).NotTo(HaveOccurred())
		Expect(buf.String()).To(Equal("\n"))
		Expect(n).To(Equal(1))
	})

	Describe("WriteExpr", func() {
		It("should write text", func() {
			buf := &bytes.Buffer{}
			w := writer.New(buf)

			n, err := w.WriteExpr(&ast.Text{
				Value: "foo",
			})

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(3))
			Expect(buf.String()).To(Equal("foo"))
		})

		It("should not write unsupported nodes", func() {
			buf := &bytes.Buffer{}
			w := writer.New(buf)

			n, err := w.WriteExpr(nil)

			Expect(err).NotTo(HaveOccurred())
			Expect(n).To(Equal(0))
			Expect(buf.Len()).To(Equal(0))
		})
	})
})
