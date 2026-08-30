package make_test

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/printer"
	"github.com/unmango/go-make/scanner"
	"github.com/unmango/go-make/token"
	"github.com/unmango/go-make/writer"
)

//go:embed testdata
var testdata embed.FS

var _ = Describe("E2E", func() {
	It("should scan this repo's Makefile", func() {
		f, err := os.Open("Makefile")
		Expect(err).NotTo(HaveOccurred())
		fi, err := f.Stat()
		Expect(err).NotTo(HaveOccurred())
		file := token.NewFileSet().AddFile(f.Name(), 1, int(fi.Size()))
		s := scanner.New(f, file)

		// By tweaking the duration and interval we can approximate the number of tokens
		// scanned and pick values that should cover the entire Makefile. This approach
		// should be able to catch infinite loops without using a count or other state
		Eventually(func() token.Token {
			_, tok, _ := s.Scan()
			return tok
		}, "5s", "1ms").Should(Equal(token.EOF))
	})

	It("should parse this repo's Makefile", func() {
		src, err := os.ReadFile("Makefile")
		Expect(err).NotTo(HaveOccurred())
		file := token.NewFileSet().AddFile("Makefile", 1, len(src))
		p := make.NewParser(bytes.NewReader(src), file)

		_, err = p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
	})

	It("should round-trip this repo's Makefile", func() {
		src, err := os.ReadFile("Makefile")
		Expect(err).NotTo(HaveOccurred())
		file := token.NewFileSet().AddFile("Makefile", 1, len(src))
		p := make.NewParser(bytes.NewReader(src), file)

		f, err := p.ParseFile()
		Expect(err).NotTo(HaveOccurred())

		buf := &bytes.Buffer{}
		Expect(printer.Fprint(buf, f)).To(BeNumerically(">", 0))
		Expect(buf.String()).To(Equal(string(src)))
	})

	It("should attach recipes separated by a blank line to their rule", func() {
		data, err := testdata.ReadFile("testdata/roundtrip/recipe-blank-line.mk")
		Expect(err).NotTo(HaveOccurred())
		p := make.NewParser(bytes.NewBuffer(data), nil)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Contents).To(HaveLen(1))
		rule, ok := f.Contents[0].(*ast.Rule)
		Expect(ok).To(BeTrue(), "expected a *ast.Rule, got %T", f.Contents[0])
		Expect(rule.Recipes).To(HaveLen(2))
		Expect(rule.Recipes[0].Value).To(Equal("first"))
		Expect(rule.Recipes[1].Value).To(Equal("second"))
	})

	DescribeTable("should round-trip", RoundTripEntries(testdata, "testdata/roundtrip"),
		func(input string) {
			p := make.NewParser(bytes.NewBufferString(input), nil)

			f, err := p.ParseFile()

			Expect(err).NotTo(HaveOccurred())

			buf := &bytes.Buffer{}
			w := writer.New(buf)

			Expect(printer.Fprint(w, f)).To(BeNumerically(">", 0))
			Expect(buf.String()).To(Equal(input))
		},
	)
})

func RoundTripEntries(fsys fs.FS, root string) (entries []TableEntry) {
	err := fs.WalkDir(fsys, root, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || filepath.Ext(path) != ".mk" {
			return nil
		}

		if data, err := fs.ReadFile(fsys, path); err != nil {
			return fmt.Errorf("reading %s: %w", path, err)
		} else {
			entries = append(entries, Entry(path, string(data)))
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	return
}
