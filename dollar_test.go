package make_test

import (
	"bytes"
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make"
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/printer"
	"github.com/unmango/go-make/writer"
)

// exprShape renders the pieces an expression was written from, each as the
// name of its node and the source text it holds. A juxtaposition renders as
// its parts, so a value make expands as one reads as one entry per piece.
// Byte round-tripping alone cannot tell "prefix$" followed by a parenthesized
// group from "prefix" followed by a reference, and this can.
func exprShape(e ast.Expr) []string {
	if j, ok := e.(*ast.JuxtaposedExpr); ok {
		shape := []string{}
		for _, part := range j.Parts {
			shape = append(shape, exprShape(part)...)
		}

		return shape
	}

	return []string{fmt.Sprintf("%T(%s)", e, e)}
}

func parseOne(input string) *ast.File {
	GinkgoHelper()
	p := make.NewParser(bytes.NewBufferString(input), nil)

	f, err := p.ParseFile()
	Expect(err).NotTo(HaveOccurred())

	buf := &bytes.Buffer{}
	Expect(printer.Fprint(writer.New(buf), f)).To(BeNumerically(">", 0))
	Expect(buf.String()).To(Equal(input), "input did not round-trip")

	return f
}

func onlyRule(f *ast.File) *ast.Rule {
	GinkgoHelper()
	Expect(f.Contents).To(HaveLen(1))
	rule, ok := f.Contents[0].(*ast.Rule)
	Expect(ok).To(BeTrue(), "expected a *ast.Rule, got %T", f.Contents[0])

	return rule
}

// recipeAt returns the i'th recipe line of rule. Rule.Recipes holds an
// ast.RecipeObj so it can carry the conditionals that select recipe lines, so
// a spec asserting on a plain recipe has to narrow it.
func recipeAt(rule *ast.Rule, i int) *ast.Recipe {
	GinkgoHelper()
	Expect(len(rule.Recipes)).To(BeNumerically(">", i))
	r, ok := rule.Recipes[i].(*ast.Recipe)
	Expect(ok).To(BeTrue(), "expected a *ast.Recipe, got %T", rule.Recipes[i])

	return r
}

func onlyVariable(f *ast.File) *ast.Variable {
	GinkgoHelper()
	Expect(f.Contents).To(HaveLen(1))
	v, ok := f.Contents[0].(*ast.Variable)
	Expect(ok).To(BeTrue(), "expected a *ast.Variable, got %T", f.Contents[0])

	return v
}

// A '$' written against the text around it separates a reference from that
// text. make expands the whole run as a single value, so the pieces belong to
// one expression rather than to several, and the value carries a reference
// rather than the parenthesized group a '$' glued to its prefix used to leave
// behind.
var _ = Describe("Dollar written against text", func() {
	It("should read a prerequisite written against a reference as one prerequisite", func() {
		rule := onlyRule(parseOne("target: prefix$(FOO)\n"))

		Expect(rule.PreReqs).To(HaveLen(1))
		Expect(exprShape(rule.PreReqs[0])).To(Equal([]string{
			"*ast.Text(prefix)",
			"*ast.VarRef($(FOO))",
		}))
	})

	It("should read a target written against a reference as one target", func() {
		rule := onlyRule(parseOne("prefix$(FOO): dep$(BAR)\n"))

		Expect(rule.Targets).To(HaveLen(1))
		Expect(rule.PreReqs).To(HaveLen(1))
		Expect(exprShape(rule.Targets[0])).To(Equal([]string{
			"*ast.Text(prefix)",
			"*ast.VarRef($(FOO))",
		}))
	})

	DescribeTable("should read a variable value as the pieces it was written from",
		func(input string, shape ...string) {
			v := onlyVariable(parseOne(input))

			Expect(v.Value).To(HaveLen(1))
			Expect(exprShape(v.Value[0])).To(Equal(shape))
		},
		Entry("text before a reference",
			"X := prefix$(FOO)\n",
			"*ast.Text(prefix)", "*ast.VarRef($(FOO))",
		),
		Entry("text after a reference",
			"X := $(FOO)suffix\n",
			"*ast.VarRef($(FOO))", "*ast.Text(suffix)",
		),
		Entry("text on both sides of a reference",
			"X := a$(FOO)b\n",
			"*ast.Text(a)", "*ast.VarRef($(FOO))", "*ast.Text(b)",
		),
		Entry("text before a braced reference",
			"X := pre${FOO}\n",
			"*ast.Text(pre)", "*ast.VarRef(${FOO})",
		),
		Entry("text before an undelimited reference",
			"X := a$b\n",
			"*ast.Text(a)", "*ast.VarRef($b)",
		),
		Entry("text before a function call",
			"X := pre$(dir a)\n",
			"*ast.Text(pre)", "*ast.FuncCall($(dir a))",
		),
		Entry("text around an escaped dollar",
			"X := pre$$post\n",
			"*ast.Text(pre)", "*ast.Text($$)", "*ast.Text(post)",
		),
		Entry("a dollar sign that opens nothing",
			"X := a$\n",
			"*ast.Text(a)", "*ast.Text($)",
		),
	)

	It("should read a call argument written against a reference as one argument", func() {
		v := onlyVariable(parseOne("X := $(dir pre$(FOO))\n"))

		Expect(v.Value).To(HaveLen(1))
		call, ok := v.Value[0].(*ast.FuncCall)
		Expect(ok).To(BeTrue(), "expected a *ast.FuncCall, got %T", v.Value[0])
		Expect(call.Args).To(HaveLen(1))
		Expect(call.Args[0].Parts).To(HaveLen(2))
		Expect(fmt.Sprint(call.Args[0].Parts[0])).To(Equal("pre"))
		Expect(fmt.Sprint(call.Args[0].Parts[1])).To(Equal("$(FOO)"))
	})

	DescribeTable("should read a conditional argument written against a reference",
		func(input, arg1, arg2 string) {
			f := parseOne(input)

			Expect(f.Contents).To(HaveLen(1))
			block, ok := f.Contents[0].(*ast.IfBlock)
			Expect(ok).To(BeTrue(), "expected an *ast.IfBlock, got %T", f.Contents[0])
			dir, ok := block.Directive.(*ast.IfeqDir)
			Expect(ok).To(BeTrue(), "expected an *ast.IfeqDir, got %T", block.Directive)
			Expect(fmt.Sprint(dir.Arg1)).To(Equal(arg1))
			Expect(fmt.Sprint(dir.Arg2)).To(Equal(arg2))
		},
		Entry("text after a reference",
			"ifeq ($(V)x,y)\ntarget:\nendif\n", "$(V)x", "y",
		),
		Entry("text before a reference",
			"ifeq (x$(V),y)\ntarget:\nendif\n", "x$(V)", "y",
		),
		Entry("text around a reference in both arguments",
			"ifeq (a$(V)b,c$(W)d)\ntarget:\nendif\n", "a$(V)b", "c$(W)d",
		),
	)

	// A recipe is captured as the flat text of the line, so the pieces the
	// scanner splits it into are reassembled from their positions and the
	// body reads exactly as it was written.
	DescribeTable("should leave a recipe body as the text it was written with",
		func(body string) {
			rule := onlyRule(parseOne("target:\n\t" + body + "\n"))

			Expect(rule.Recipes).To(HaveLen(1))
			recipe, ok := rule.Recipes[0].(*ast.Recipe)
			Expect(ok).To(BeTrue(), "expected a *ast.Recipe, got %T", rule.Recipes[0])
			Expect(recipe.Value).To(Equal(body))
		},
		Entry(nil, "echo $$HOME"),
		Entry(nil, "echo $(VAR)"),
		Entry(nil, "echo pre$(VAR)post"),
		Entry(nil, "echo pre$$HOME"),
		Entry(nil, "echo $(dir a)b"),
		Entry(nil, "echo a$"),
	)
})
