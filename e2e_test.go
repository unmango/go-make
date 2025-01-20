package make_test

import (
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/token"
)

var _ = Describe("E2E", func() {
	It("should scan this repo's Makefile", func() {
		f, err := os.Open("Makefile")
		Expect(err).NotTo(HaveOccurred())
		fi, err := f.Stat()
		Expect(err).NotTo(HaveOccurred())
		file := token.NewFileSet().AddFile(f.Name(), 1, int(fi.Size()))
		s := make.NewScanner(f, file)

		// By tweaking the duration and interval we can approximate the number of tokens
		// scanned and pick values that should cover the entire Makefile. This approach
		// should be able to catch infinite loops without using a count or other state
		Eventually(func() token.Token {
			_, tok, _ := s.Scan()
			return tok
		}, "500ms", "1ms").Should(Equal(token.EOF))
	})

	It("should parse this repo's Makefile", Pending, func() {
		f, err := os.Open("Makefile")
		Expect(err).NotTo(HaveOccurred())
		fi, err := f.Stat()
		Expect(err).NotTo(HaveOccurred())
		file := token.NewFileSet().AddFile(f.Name(), 1, int(fi.Size()))
		p := make.NewParser(f, file)

		_, err = p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
	})
})
