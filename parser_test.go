package make_test

import (
	"bytes"
	"go/token"
	gotoken "go/token"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/ast"
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
		Expect(f.Rules).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(8),
			Targets: &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "target",
					NamePos: token.Pos(0),
				}},
			}},
		}))
	})

	DescribeTable("should error on invalid starting token",
		Entry(nil, ","),
		Entry(nil, ":"),
		Entry(nil, ";"),
		Entry(nil, "|"),
		Entry(nil, "="),
		func(input string) {
			buf := bytes.NewBufferString(input)
			p := make.NewParser(buf, file)

			_, err := p.ParseFile()

			Expect(err).To(MatchError(
				ContainSubstring("expected 'IDENT'"),
			))
		},
	)
})
