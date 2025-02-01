package make_test

import (
	"bytes"
	"embed"
	"fmt"
	"io/fs"
	"os"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
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
		}, "1s", "1ms").Should(Equal(token.EOF))
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
		if d.IsDir() {
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
