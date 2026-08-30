package make_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
)

func onlyIfeqDir(input string) *ast.IfeqDir {
	GinkgoHelper()
	f := parseOne(input)

	Expect(f.Contents).To(HaveLen(1))
	block, ok := f.Contents[0].(*ast.IfBlock)
	Expect(ok).To(BeTrue(), "expected an *ast.IfBlock, got %T", f.Contents[0])
	dir, ok := block.Directive.(*ast.IfeqDir)
	Expect(ok).To(BeTrue(), "expected an *ast.IfeqDir, got %T", block.Directive)

	return dir
}

// A semicolon ends the prerequisite list of a rule and introduces the recipe
// that follows it, wherever it is written. Everywhere else it is an ordinary
// character of the value it was written in.
var _ = Describe("Semicolon", func() {
	It("should end a prerequisite written against it", func() {
		rule := onlyRule(parseOne("target: prereq;recipe\n"))

		Expect(rule.PreReqs).To(HaveLen(1))
		Expect(exprShape(rule.PreReqs[0])).To(Equal([]string{"*ast.Text(prereq)"}))
		Expect(rule.Recipes).To(HaveLen(1))
		Expect(rule.Recipes[0].Value).To(Equal("recipe"))
	})

	It("should introduce a recipe written against it with no prerequisites", func() {
		rule := onlyRule(parseOne("target: ;recipe\n"))

		Expect(rule.PreReqs).To(BeEmpty())
		Expect(rule.Recipes).To(HaveLen(1))
		Expect(rule.Recipes[0].Value).To(Equal("recipe"))
	})

	It("should keep the rest of the line in the recipe it introduces", func() {
		rule := onlyRule(parseOne("target: a;b c\n"))

		Expect(rule.PreReqs).To(HaveLen(1))
		Expect(exprShape(rule.PreReqs[0])).To(Equal([]string{"*ast.Text(a)"}))
		Expect(rule.Recipes).To(HaveLen(1))
		Expect(rule.Recipes[0].Value).To(Equal("b c"))
	})

	DescribeTable("should read a variable value as the words it was written from",
		func(input string, shapes ...[]string) {
			v := onlyVariable(parseOne(input))

			Expect(v.Value).To(HaveLen(len(shapes)))
			for i, shape := range shapes {
				Expect(exprShape(v.Value[i])).To(Equal(shape))
			}
		},
		Entry("a semicolon against the text before it",
			"X = a; b\n",
			[]string{"*ast.Text(a)", "*ast.Text(;)"},
			[]string{"*ast.Text(b)"},
		),
		Entry("a semicolon between two pieces of text",
			"X = a;b\n",
			[]string{"*ast.Text(a)", "*ast.Text(;)", "*ast.Text(b)"},
		),
		Entry("a semicolon written on its own",
			"X = a ; b\n",
			[]string{"*ast.Text(a)"},
			[]string{"*ast.Text(;)"},
			[]string{"*ast.Text(b)"},
		),
		Entry("a semicolon against the text after it",
			"X = ;b\n",
			[]string{"*ast.Text(;)", "*ast.Text(b)"},
		),
		Entry("a semicolon ending the value",
			"X = a;\n",
			[]string{"*ast.Text(a)", "*ast.Text(;)"},
		),
	)

	It("should read a conditional argument holding a semicolon as one argument", func() {
		dir := onlyIfeqDir("ifeq (a;b,c)\ntarget:\nendif\n")

		Expect(fmt.Sprint(dir.Arg1)).To(Equal("a;b"))
		Expect(fmt.Sprint(dir.Arg2)).To(Equal("c"))
	})
})

// A pipe separates the normal prerequisites of a rule from the order-only
// prerequisites that follow it, whether or not it is written apart from them.
// Everywhere else it is an ordinary character of the value it was written in.
var _ = Describe("Pipe", func() {
	DescribeTable("should separate order-only prerequisites written against it",
		func(input string, prereqs, oprereqs []string) {
			rule := onlyRule(parseOne(input))

			Expect(rule.PreReqs).To(HaveLen(len(prereqs)))
			for i, p := range prereqs {
				Expect(fmt.Sprint(rule.PreReqs[i])).To(Equal(p))
			}
			Expect(rule.OrderPreReqs).To(HaveLen(len(oprereqs)))
			for i, o := range oprereqs {
				Expect(fmt.Sprint(rule.OrderPreReqs[i])).To(Equal(o))
			}
		},
		Entry("a pipe against the prerequisites on both sides",
			"target: prereq|order\n", []string{"prereq"}, []string{"order"},
		),
		Entry("a pipe against the colon",
			"target:|prereq\n", []string{}, []string{"prereq"},
		),
		// make reads only the first pipe as the separator, so a second one is
		// an ordinary character of the order-only prerequisite holding it.
		Entry("a second pipe as part of an order-only prerequisite",
			"target: a|b|c\n", []string{"a"}, []string{"b|c"},
		),
		Entry("a pipe followed by a semicolon recipe",
			"target: a|b;recipe\n", []string{"a"}, []string{"b"},
		),
	)

	It("should introduce a recipe after order-only prerequisites written against a pipe", func() {
		rule := onlyRule(parseOne("target: a|b;recipe\n"))

		Expect(rule.Recipes).To(HaveLen(1))
		Expect(rule.Recipes[0].Value).To(Equal("recipe"))
	})

	DescribeTable("should read a variable value as the words it was written from",
		func(input string, shapes ...[]string) {
			v := onlyVariable(parseOne(input))

			Expect(v.Value).To(HaveLen(len(shapes)))
			for i, shape := range shapes {
				Expect(exprShape(v.Value[i])).To(Equal(shape))
			}
		},
		Entry("a pipe between two pieces of text",
			"X = a|b\n",
			[]string{"*ast.Text(a)", "*ast.Text(|)", "*ast.Text(b)"},
		),
		Entry("a pipe written on its own",
			"X = a | b\n",
			[]string{"*ast.Text(a)"},
			[]string{"*ast.Text(|)"},
			[]string{"*ast.Text(b)"},
		),
		Entry("two pipes written together",
			"X = a||b\n",
			[]string{"*ast.Text(a)", "*ast.Text(|)", "*ast.Text(|)", "*ast.Text(b)"},
		),
		Entry("a pipe ending the value",
			"X = a|\n",
			[]string{"*ast.Text(a)", "*ast.Text(|)"},
		),
		Entry("a pipe against the text after it",
			"X = |b\n",
			[]string{"*ast.Text(|)", "*ast.Text(b)"},
		),
	)

	It("should read a conditional argument holding a pipe as one argument", func() {
		dir := onlyIfeqDir("ifeq (a|b,c)\ntarget:\nendif\n")

		Expect(fmt.Sprint(dir.Arg1)).To(Equal("a|b"))
		Expect(fmt.Sprint(dir.Arg2)).To(Equal("c"))
	})
})

// Only a target line gives a ';' or a '|' a meaning, so the name of an
// expansion and a quoted conditional argument hold either character as the
// text it was written with.
var _ = Describe("Punctuation written inside a delimited construct", func() {
	DescribeTable("should read the name of an expression holding punctuation",
		func(input, name string) {
			v := onlyVariable(parseOne(input))

			Expect(v.Value).To(HaveLen(1))
			ref, ok := v.Value[0].(*ast.VarRef)
			Expect(ok).To(BeTrue(), "expected an *ast.VarRef, got %T", v.Value[0])
			Expect(ref.Name).To(Equal(name))
		},
		Entry("a semicolon in a parenthesized name", "X = $(a;b)\n", "a;b"),
		Entry("a pipe in a parenthesized name", "X = $(a|b)\n", "a|b"),
		Entry("a pipe in a braced name", "X = ${a|b}\n", "a|b"),
		Entry("punctuation ending a name", "X = $(a;)\n", "a;"),
	)

	DescribeTable("should read a quoted conditional argument holding punctuation",
		func(input, arg1, arg2 string) {
			dir := onlyIfeqDir(input)

			Expect(fmt.Sprint(dir.Arg1)).To(Equal(arg1))
			Expect(fmt.Sprint(dir.Arg2)).To(Equal(arg2))
		},
		Entry("apostrophes",
			"ifeq 'a;b' 'a|b'\nX = 1\nendif\n", "'a;b'", "'a|b'",
		),
		Entry("quotes",
			"ifeq \"a;b\" \"a|b\"\nX = 1\nendif\n", `"a;b"`, `"a|b"`,
		),
	)
})

// A recipe is captured as the flat text of the line, so the pieces the scanner
// splits it into are reassembled from their positions and the body reads
// exactly as it was written.
var _ = DescribeTable("should leave a recipe body holding punctuation as the text it was written with",
	func(body string) {
		rule := onlyRule(parseOne("target:\n\t" + body + "\n"))

		Expect(rule.Recipes).To(HaveLen(1))
		Expect(rule.Recipes[0].Value).To(Equal(body))
	},
	Entry(nil, "cd dir; ls"),
	Entry(nil, "test -f x || echo no"),
	Entry(nil, "ls | wc -l"),
	Entry(nil, "a;b;c"),
	Entry(nil, "a|b|c"),
)
