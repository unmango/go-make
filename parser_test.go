package make_test

import (
	"bytes"
	gotoken "go/token"
	"math"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
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
			Colon: token.Pos(7),
			Targets: &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "target",
					NamePos: token.Pos(1),
				}},
			}},
			PreReqs: &ast.PreReqList{},
			Recipes: []*ast.Recipe{},
		}))
	})

	It("should Parse a rule with multiple targets", func() {
		buf := bytes.NewBufferString("target target2:")
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Rules).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(15),
			Targets: &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "target",
					NamePos: token.Pos(1),
				}},
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "target2",
					NamePos: token.Pos(8),
				}},
			}},
			PreReqs: &ast.PreReqList{},
			Recipes: []*ast.Recipe{},
		}))
	})

	It("should Parse a target with a prereq", func() {
		buf := bytes.NewBufferString("target: prereq")
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Rules).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "target",
					NamePos: token.Pos(1),
				}},
			}},
			PreReqs: &ast.PreReqList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "prereq",
					NamePos: token.Pos(9),
				}},
			}},
			Recipes: []*ast.Recipe{},
		}))
	})

	It("should Parse a target with multiple prereqs", func() {
		buf := bytes.NewBufferString("target: prereq prereq2")
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Rules).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "target",
					NamePos: token.Pos(1),
				}},
			}},
			PreReqs: &ast.PreReqList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "prereq",
					NamePos: token.Pos(9),
				}},
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "prereq2",
					NamePos: token.Pos(16),
				}},
			}},
			Recipes: []*ast.Recipe{},
		}))
	})

	It("should Parse a target with a recipe", func() {
		buf := bytes.NewBufferString("target:\n\trecipe")
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Rules).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "target",
					NamePos: token.Pos(1),
				}},
			}},
			PreReqs: &ast.PreReqList{},
			Recipes: []*ast.Recipe{{
				Tok:    token.TAB,
				TokPos: token.Pos(9),
				Text:   "recipe",
			}},
		}))
	})

	It("should Parse a target with multiple recipes", func() {
		buf := bytes.NewBufferString("target:\n\trecipe\n\trecipe2")
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Rules).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "target",
					NamePos: token.Pos(1),
				}},
			}},
			PreReqs: &ast.PreReqList{},
			Recipes: []*ast.Recipe{
				{
					Tok:    token.TAB,
					TokPos: token.Pos(9),
					Text:   "recipe",
				},
				{
					Tok:    token.TAB,
					TokPos: token.Pos(17),
					Text:   "recipe2",
				},
			},
		}))
	})

	It("should Parse a target with a prereq and a recipe", func() {
		buf := bytes.NewBufferString("target: prereq\n\trecipe")
		p := make.NewParser(buf, file)

		f, err := p.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Rules).To(ConsistOf(&ast.Rule{
			Colon: token.Pos(7),
			Targets: &ast.TargetList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "target",
					NamePos: token.Pos(1),
				}},
			}},
			PreReqs: &ast.PreReqList{List: []ast.FileName{
				&ast.LiteralFileName{Name: &ast.Ident{
					Name:    "prereq",
					NamePos: token.Pos(9),
				}},
			}},
			Recipes: []*ast.Recipe{{
				Tok:    token.TAB,
				TokPos: token.Pos(16),
				Text:   "recipe",
			}},
		}))
	})

	DescribeTable("should error on invalid starting token",
		Entry(nil, ","),
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

	It("should support a nil *token.File value", func() {
		buf := bytes.NewBufferString("target:")
		s := make.NewParser(buf, nil)

		f, err := s.ParseFile()

		Expect(err).NotTo(HaveOccurred())
		Expect(f.Rules).NotTo(BeEmpty())
	})
})
