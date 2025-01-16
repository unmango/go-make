package make_test

import (
	"bytes"
	"go/token"
	gotoken "go/token"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
)

var _ = Describe("Parser", func() {
	var file *token.File

	BeforeEach(func() {
		file = gotoken.NewFileSet().AddFile("test", 1, math.MaxInt-2)
	})

	It("should Parse a target", func() {
		buf := bytes.NewBufferString("target:")
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f).NotTo(BeNil())
	})

	It("should error when starting at a colon", func() {
		buf := bytes.NewBufferString(":")
		p := make.NewParser(buf, file)

		_, err := p.ParseFile()

		Expect(err).To(MatchError("expected 'IDENT'"))
	})
})
