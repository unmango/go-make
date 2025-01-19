package make_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/ast"
)

var _ = Describe("Print", func() {
	Describe("Fprint", func() {
		It("should print a literal file name", func() {
			buf := &bytes.Buffer{}
			l := &ast.LiteralFileName{Name: &ast.Ident{
				Name: "target",
			}}

			err := make.Fprint(buf, l)

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target"))
		})

		It("should print a target list", func() {
			buf := &bytes.Buffer{}
			t := &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name: "target",
				}},
			}}

			err := make.Fprint(buf, t)

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("target:"))
		})

		It("should print a prereq list", func() {
			buf := &bytes.Buffer{}
			t := &ast.PreReqList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name: "prereq",
				}},
			}}

			err := make.Fprint(buf, t)

			Expect(err).NotTo(HaveOccurred())
			Expect(buf.String()).To(Equal("prereq"))
		})
	})
})
