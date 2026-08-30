package obj_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/obj"
	"github.com/unmango/go-make/printer"
	"github.com/unmango/go-make/token"
)

func printed(node any) string {
	buf := &bytes.Buffer{}
	_, err := printer.Fprint(buf, node)
	Expect(err).NotTo(HaveOccurred())

	return buf.String()
}

func commentGroup(text ...string) *ast.CommentGroup {
	g := &ast.CommentGroup{}
	for _, t := range text {
		g.List = append(g.List, &ast.Comment{Text: t})
	}

	return g
}

func variable(name, value string) *ast.Variable {
	return &ast.Variable{
		Name:  &ast.Text{Value: name},
		Op:    token.SIMPLE_ASSIGN,
		Value: []ast.Expr{&ast.Text{Value: value}},
	}
}

func ifBlock() *ast.IfBlock {
	return &ast.IfBlock{
		Directive: &ast.IfdefDir{
			Tok:     token.IFDEF,
			VarName: &ast.Text{Value: "test"},
		},
		Text: []ast.Obj{
			&ast.Rule{Targets: []ast.Expr{&ast.Text{Value: "targetA"}}},
		},
		Else: []*ast.ElseBlock{{
			Condition: &ast.IfeqDir{
				Tok:   token.IFEQ,
				Open:  token.Pos(1),
				Arg1:  &ast.Text{Value: "a"},
				Arg2:  &ast.Text{Value: "b"},
				Close: token.Pos(1),
			},
			Text: []ast.Obj{
				&ast.Rule{Targets: []ast.Expr{&ast.Text{Value: "targetB"}}},
			},
		}},
	}
}

var _ = Describe("Obj", func() {
	Describe("Copy", func() {
		DescribeTable("should print the copied object",
			func(expected string, o ast.Obj) {
				Expect(printed(obj.Copy(1, o))).To(Equal(expected))
			},
			Entry("a rule", "target: prereq\n", &ast.Rule{
				Targets: []ast.Expr{&ast.Text{Value: "target"}},
				PreReqs: []ast.Expr{&ast.Text{Value: "prereq"}},
			}),
			Entry("a comment", "# a comment\n", commentGroup("a comment")),
			Entry("a comment group", "# a comment\n# another line\n",
				commentGroup("a comment", "another line"),
			),
			Entry("a variable", "FOO := bar\n", variable("FOO", "bar")),
			Entry("a variable with no value", "FOO :=\n", &ast.Variable{
				Name: &ast.Text{Value: "FOO"},
				Op:   token.SIMPLE_ASSIGN,
			}),
			Entry("an if block", "ifdef test\ntargetA:\nelse ifeq (a, b)\ntargetB:\nendif\n",
				ifBlock(),
			),
			Entry("a bad object", "include foo.mk\n", &ast.BadObj{Text: "include foo.mk"}),
		)

		It("should not alias the copied object", func() {
			v := variable("FOO", "bar")

			actual := obj.Copy(1, v).(*ast.Variable)

			Expect(actual).NotTo(BeIdenticalTo(v))
			Expect(actual.Name).NotTo(BeIdenticalTo(v.Name))
			Expect(actual.Value[0]).NotTo(BeIdenticalTo(v.Value[0]))
			Expect(v.Name.Pos()).To(Equal(token.NoPos))
		})

		It("should not alias the objects of an if block", func() {
			b := ifBlock()

			actual := obj.Copy(1, b).(*ast.IfBlock)

			Expect(actual.Text[0]).NotTo(BeIdenticalTo(b.Text[0]))
			Expect(actual.Else[0]).NotTo(BeIdenticalTo(b.Else[0]))
			Expect(actual.Else[0].Text[0]).NotTo(BeIdenticalTo(b.Else[0].Text[0]))
			Expect(b.Endif).To(Equal(token.NoPos))
		})
	})

	DescribeTable("End",
		func(o ast.Obj, expected token.Pos) {
			Expect(obj.End(obj.Copy(1, o))).To(Equal(expected))
		},
		Entry("a rule", &ast.Rule{
			Targets: []ast.Expr{&ast.Text{Value: "target"}},
		}, token.Pos(9)), // len("target:\n") + 1
		Entry("a comment group", commentGroup("a comment"), token.Pos(13)),
		Entry("a variable", variable("FOO", "bar"), token.Pos(12)),
		Entry("an if block", ifBlock(), token.Pos(53)),
		Entry("a bad object", &ast.BadObj{Text: "include foo.mk"},
			token.Pos(16), // len("include foo.mk\n") + 1
		),
	)

	Describe("Pos", func() {
		It("should not panic for a rule without targets", func() {
			r := &ast.Rule{Colon: 4}

			Expect(obj.Pos(r)).To(Equal(token.Pos(4)))
		})

		It("should not panic for an empty comment group", func() {
			Expect(obj.Pos(&ast.CommentGroup{})).To(Equal(token.NoPos))
		})

		It("should return the position of the first object", func() {
			Expect(obj.Pos(obj.Copy(3, variable("FOO", "bar")))).To(Equal(token.Pos(3)))
		})
	})
})
