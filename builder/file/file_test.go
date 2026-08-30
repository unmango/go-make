package file_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/file"
	"github.com/unmango/go-make/builder/rule"
	"github.com/unmango/go-make/printer"
	"github.com/unmango/go-make/token"
)

func printed(node any) string {
	buf := &bytes.Buffer{}
	_, err := printer.Fprint(buf, node)
	Expect(err).NotTo(HaveOccurred())

	return buf.String()
}

var _ = Describe("File", func() {
	Describe("New", func() {
		It("should work", func() {
			f := file.New(1,
				file.Rule(rule.TextTarget("target")),
				file.Rule(rule.TextTarget("target2")),
			)

			Expect(f).To(Equal(&ast.File{
				FileStart: 1,
				Contents: []ast.Obj{
					rule.New(1, rule.TextTarget("target")),
					rule.New(9, rule.TextTarget("target2")),
				},
				FileEnd: 18,
			}))
		})

		It("should position each target and pre-requisite", func() {
			f := file.New(1, file.Rule(
				rule.TextTarget("a"),
				rule.TextTarget("b"),
				rule.TextPreReq("c"),
			))

			Expect(printed(f)).To(Equal("a b: c\n"))
		})

		It("should position each rule", func() {
			f := file.New(1,
				file.Rule(rule.TextTarget("a"), rule.TextPreReq("b")),
				file.Rule(rule.TextTarget("c")),
			)

			Expect(printed(f)).To(Equal("a: b\nc:\n"))
		})
	})

	Describe("InsertRule", func() {
		It("should work", func() {
			f := file.New(1,
				file.Rule(rule.TextTarget("target")),
				file.Rule(rule.TextTarget("target3")),
			)

			file.InsertRule(f, 1, rule.TextTarget("target2"))

			Expect(f).To(Equal(&ast.File{
				FileStart: 1,
				Contents: []ast.Obj{
					rule.New(1, rule.TextTarget("target")),
					rule.New(9, rule.TextTarget("target2")),
					rule.New(18, rule.TextTarget("target3")),
				},
				FileEnd: 27,
			}))
		})

		DescribeTable("should insert the rule at i",
			func(i int, expected string) {
				f := file.New(1,
					file.Rule(rule.TextTarget("a")),
					file.Rule(rule.TextTarget("b")),
				)

				file.InsertRule(f, i, rule.TextTarget("z"), rule.TextPreReq("y"))

				Expect(printed(f)).To(Equal(expected))
			},
			Entry("at the beginning", 0, "z: y\na:\nb:\n"),
			Entry("in the middle", 1, "a:\nz: y\nb:\n"),
			Entry("at the end", 2, "a:\nb:\nz: y\n"),
			Entry("past the end", 5, "a:\nb:\nz: y\n"),
			Entry("before the beginning", -3, "z: y\na:\nb:\n"),
		)

		It("should insert into an empty file", func() {
			f := file.New(1)

			file.InsertRule(f, 0, rule.TextTarget("a"))

			Expect(printed(f)).To(Equal("a:\n"))
			Expect(f.FileEnd).To(Equal(token.Pos(4)))
		})

		It("should append when i is past the end", func() {
			f := file.New(1, file.Rule(rule.TextTarget("a")))

			file.InsertRule(f, 5, rule.TextTarget("b"))

			Expect(f.Contents).To(HaveLen(2))
			Expect(printed(f)).To(Equal("a:\nb:\n"))
		})

		It("should keep the file end consistent with New", func() {
			f := file.New(1, file.Rule(rule.TextTarget("a")))

			file.InsertRule(f, 1, rule.TextTarget("b"))

			Expect(f.FileEnd).To(Equal(file.New(1,
				file.Rule(rule.TextTarget("a")),
				file.Rule(rule.TextTarget("b")),
			).FileEnd))
		})

		It("should insert into a file containing a comment group and a variable", func() {
			f := &ast.File{
				FileStart: 1,
				Contents: []ast.Obj{
					&ast.CommentGroup{List: []*ast.Comment{
						{Pound: 1, Text: "a comment"},
					}},
					&ast.Variable{
						Name:  &ast.Text{Value: "FOO", ValuePos: 13},
						Op:    token.SIMPLE_ASSIGN,
						OpPos: 17,
						Value: []ast.Expr{&ast.Text{Value: "bar", ValuePos: 20}},
					},
				},
				FileEnd: 24,
			}

			file.InsertRule(f, 1, rule.TextTarget("target"), rule.TextPreReq("prereq"))

			Expect(printed(f)).To(Equal("# a comment\ntarget: prereq\nFOO := bar\n"))
		})

		It("should preserve the objects preceding the insertion point", func() {
			group := &ast.CommentGroup{List: []*ast.Comment{
				{Pound: 1, Text: "a comment"},
			}}
			f := &ast.File{FileStart: 1, Contents: []ast.Obj{group}, FileEnd: 13}

			file.InsertRule(f, 1, rule.TextTarget("target"))

			Expect(f.Contents[0]).To(BeIdenticalTo(group))
			Expect(printed(f)).To(Equal("# a comment\ntarget:\n"))
		})
	})
})
