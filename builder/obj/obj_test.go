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

// ifBlockWithComments returns a conditional whose every line ends in a
// comment.
func ifBlockWithComments() *ast.IfBlock {
	return &ast.IfBlock{
		Directive: &ast.IfdefDir{
			Tok:     token.IFDEF,
			VarName: &ast.Text{Value: "test"},
			Comment: &ast.Comment{Text: " one"},
		},
		Text: []ast.Obj{
			&ast.Rule{Targets: []ast.Expr{&ast.Text{Value: "targetA"}}},
		},
		Else: []*ast.ElseBlock{{
			Comment: &ast.Comment{Text: " two"},
			Text: []ast.Obj{
				&ast.Rule{Targets: []ast.Expr{&ast.Text{Value: "targetB"}}},
			},
		}},
		EndifComment: &ast.Comment{Text: " three"},
	}
}

// ruleWithConditional returns a rule whose recipe list holds a recipe line and
// a conditional wrapping another.
func ruleWithConditional() *ast.Rule {
	return &ast.Rule{
		Targets: []ast.Expr{&ast.Text{Value: "target"}},
		Recipes: []ast.RecipeObj{
			&ast.Recipe{Prefix: token.TAB, Text: ast.Text{Value: "one"}},
			&ast.IfBlock{
				Directive: &ast.IfdefDir{
					Tok:     token.IFDEF,
					VarName: &ast.Text{Value: "V"},
				},
				Text: []ast.Obj{
					&ast.Recipe{Prefix: token.TAB, Text: ast.Text{Value: "two"}},
				},
			},
		},
	}
}

// defineDir returns a define block holding a body of two lines, the second of
// which is blank.
func defineDir() *ast.DefineDir {
	return &ast.DefineDir{
		VarName: &ast.Text{Value: "FOO"},
		Op:      token.APPEND_ASSIGN,
		Body: []*ast.Text{
			{Value: "one"},
			{Value: ""},
			{Value: "\ttwo"},
		},
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
			Entry("a comment", "# a comment\n", commentGroup(" a comment")),
			Entry("a comment group", "# a comment\n# another line\n",
				commentGroup(" a comment", " another line"),
			),
			Entry("a variable", "FOO := bar\n", variable("FOO", "bar")),
			Entry("a variable with no value", "FOO :=\n", &ast.Variable{
				Name: &ast.Text{Value: "FOO"},
				Op:   token.SIMPLE_ASSIGN,
			}),
			Entry("an if block", "ifdef test\ntargetA:\nelse ifeq (a, b)\ntargetB:\nendif\n",
				ifBlock(),
			),
			Entry("an if block whose lines end in comments",
				"ifdef test # one\ntargetA:\nelse # two\ntargetB:\nendif # three\n",
				ifBlockWithComments(),
			),
			Entry("a bad object", "include foo.mk\n", &ast.BadObj{Text: "include foo.mk"}),
			Entry("a rule holding a conditional", "target:\n\tone\nifdef V\n\ttwo\nendif\n",
				ruleWithConditional(),
			),
			Entry("a define block", "define FOO +=\none\n\n\ttwo\nendef\n", defineDir()),
			Entry("a define block with an empty body", "define FOO\nendef\n",
				&ast.DefineDir{VarName: &ast.Text{Value: "FOO"}, Op: token.ILLEGAL},
			),
			Entry("an undefine directive", "undefine FOO\n",
				&ast.UndefineDir{VarName: &ast.Text{Value: "FOO"}},
			),
		)

		// A recipe is an expression as well as an object, and the printer
		// matches an expression first, so a recipe copied on its own is
		// checked by its layout rather than by what it prints to.
		It("should lay out a recipe", func() {
			r := &ast.Recipe{Prefix: token.TAB, Text: ast.Text{Value: "one"}}

			actual := obj.Copy(1, r).(*ast.Recipe)

			Expect(actual.PrefixPos).To(Equal(token.Pos(1)))
			Expect(actual.ValuePos).To(Equal(token.Pos(2)))
			Expect(r.PrefixPos).To(Equal(token.NoPos))
		})

		It("should not alias the conditional in a recipe list", func() {
			r := ruleWithConditional()

			actual := obj.Copy(1, r).(*ast.Rule)

			Expect(actual.Recipes[0]).NotTo(BeIdenticalTo(r.Recipes[0]))
			Expect(actual.Recipes[1]).NotTo(BeIdenticalTo(r.Recipes[1]))
		})

		It("should not alias the copied object", func() {
			v := variable("FOO", "bar")

			actual := obj.Copy(1, v).(*ast.Variable)

			Expect(actual).NotTo(BeIdenticalTo(v))
			Expect(actual.Name).NotTo(BeIdenticalTo(v.Name))
			Expect(actual.Value[0]).NotTo(BeIdenticalTo(v.Value[0]))
			Expect(v.Name.Pos()).To(Equal(token.NoPos))
		})

		It("should not alias the body of a define block", func() {
			d := defineDir()

			actual := obj.Copy(1, d).(*ast.DefineDir)

			Expect(actual).NotTo(BeIdenticalTo(d))
			Expect(actual.VarName).NotTo(BeIdenticalTo(d.VarName))
			Expect(actual.Body[0]).NotTo(BeIdenticalTo(d.Body[0]))
			Expect(d.Define).To(Equal(token.NoPos))
			Expect(d.Endef).To(Equal(token.NoPos))
		})

		It("should not alias the objects of an if block", func() {
			b := ifBlock()

			actual := obj.Copy(1, b).(*ast.IfBlock)

			Expect(actual.Text[0]).NotTo(BeIdenticalTo(b.Text[0]))
			Expect(actual.Else[0]).NotTo(BeIdenticalTo(b.Else[0]))
			Expect(actual.Else[0].Text[0]).NotTo(BeIdenticalTo(b.Else[0].Text[0]))
			Expect(b.Endif).To(Equal(token.NoPos))
		})

		// A comment is reached through a pointer, so a copy that shares one
		// with the original moves the original's comment when it is laid out.
		It("should not alias the comments of an if block", func() {
			b := ifBlockWithComments()

			actual := obj.Copy(1, b).(*ast.IfBlock)

			Expect(actual.EndifComment).NotTo(BeIdenticalTo(b.EndifComment))
			Expect(actual.Else[0].Comment).NotTo(BeIdenticalTo(b.Else[0].Comment))
			dir := actual.Directive.(*ast.IfdefDir)
			Expect(dir.Comment).NotTo(BeIdenticalTo(b.Directive.(*ast.IfdefDir).Comment))
			Expect(b.EndifComment.Pound).To(Equal(token.NoPos))
		})
	})

	DescribeTable("End",
		func(o ast.Obj, expected token.Pos) {
			Expect(obj.End(obj.Copy(1, o))).To(Equal(expected))
		},
		Entry("a rule", &ast.Rule{
			Targets: []ast.Expr{&ast.Text{Value: "target"}},
		}, token.Pos(9)), // len("target:\n") + 1
		Entry("a comment group", commentGroup(" a comment"), token.Pos(13)),
		Entry("a variable", variable("FOO", "bar"), token.Pos(12)),
		Entry("an if block", ifBlock(), token.Pos(53)),
		Entry("an if block whose lines end in comments", ifBlockWithComments(),
			// len("ifdef test # one\ntargetA:\nelse # two\ntargetB:\nendif # three\n") + 1
			token.Pos(61),
		),
		Entry("a bad object", &ast.BadObj{Text: "include foo.mk"},
			token.Pos(16), // len("include foo.mk\n") + 1
		),
		Entry("a rule holding a conditional", ruleWithConditional(),
			token.Pos(33), // len("target:\n\tone\nifdef V\n\ttwo\nendif\n") + 1
		),
		Entry("a recipe", &ast.Recipe{Prefix: token.TAB, Text: ast.Text{Value: "one"}},
			token.Pos(6), // len("\tone\n") + 1
		),
		Entry("a define block", defineDir(),
			token.Pos(31), // len("define FOO +=\none\n\n\ttwo\nendef\n") + 1
		),
		Entry("an undefine directive", &ast.UndefineDir{VarName: &ast.Text{Value: "FOO"}},
			token.Pos(14), // len("undefine FOO\n") + 1
		),
		Entry("an undefine directive without a name", &ast.UndefineDir{},
			token.Pos(10), // len("undefine\n") + 1
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
