package make_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
)

// ruleComment returns the comment ending the target line of rule, failing
// when the rule carries none.
func ruleComment(rule *ast.Rule) *ast.Comment {
	GinkgoHelper()
	Expect(rule.Comment).NotTo(BeNil(), "expected the rule to carry a comment")

	return rule.Comment
}

// variableComment returns the comment ending the assignment line of v,
// failing when the variable carries none.
func variableComment(v *ast.Variable) *ast.Comment {
	GinkgoHelper()
	Expect(v.Comment).NotTo(BeNil(), "expected the variable to carry a comment")

	return v.Comment
}

// A '#' begins a comment wherever make reads make syntax, whether or not it is
// written apart from the text before it. GNU Make 4.4.1 removes the comment
// from the line before reading anything else on it, so "prereq# text" is one
// prerequisite and a comment rather than a prerequisite spelling "prereq#".
var _ = Describe("Comment written against text", func() {
	DescribeTable("should end a prerequisite list",
		func(input string, prereqs, oprereqs []string, text string) {
			rule := onlyRule(parseOne(input))

			Expect(rule.PreReqs).To(HaveLen(len(prereqs)))
			for i, p := range prereqs {
				Expect(exprShape(rule.PreReqs[i])).To(Equal([]string{p}))
			}
			Expect(rule.OrderPreReqs).To(HaveLen(len(oprereqs)))
			for i, o := range oprereqs {
				Expect(exprShape(rule.OrderPreReqs[i])).To(Equal([]string{o}))
			}
			Expect(ruleComment(rule).Text).To(Equal(text))
		},
		Entry("a comment written against a prerequisite",
			"target: prereq# a comment\n",
			[]string{"*ast.Text(prereq)"}, nil, " a comment",
		),
		Entry("a comment written apart from a prerequisite",
			"target: prereq # a comment\n",
			[]string{"*ast.Text(prereq)"}, nil, " a comment",
		),
		Entry("a comment with no prerequisites in front of it",
			"target: # a comment\n",
			nil, nil, " a comment",
		),
		Entry("a comment written against the colon",
			"target:# a comment\n",
			nil, nil, " a comment",
		),
		Entry("a comment written against an order-only prerequisite",
			"target: a | b# a comment\n",
			[]string{"*ast.Text(a)"}, []string{"*ast.Text(b)"}, " a comment",
		),
		Entry("a comment written against a reference",
			"target: $(DEPS)# a comment\n",
			[]string{"*ast.VarRef($(DEPS))"}, nil, " a comment",
		),
		Entry("a comment holding no text",
			"target: prereq #\n",
			[]string{"*ast.Text(prereq)"}, nil, "",
		),
	)

	It("should keep the recipes of a rule whose target line ends in a comment", func() {
		rule := onlyRule(parseOne("target: prereq # a comment\n\t@echo hi\n"))

		Expect(ruleComment(rule).Text).To(Equal(" a comment"))
		Expect(rule.Recipes).To(HaveLen(1))
		Expect(recipeAt(rule, 0).Value).To(Equal("@echo hi"))
	})

	DescribeTable("should end a variable value",
		func(input string, values []string, text string) {
			v := onlyVariable(parseOne(input))

			Expect(v.Value).To(HaveLen(len(values)))
			for i, s := range values {
				Expect(exprShape(v.Value[i])).To(Equal([]string{s}))
			}
			Expect(variableComment(v).Text).To(Equal(text))
		},
		Entry("a comment written against the value",
			"X = a# a comment\n", []string{"*ast.Text(a)"}, " a comment",
		),
		Entry("a comment written apart from the value",
			"X = a # a comment\n", []string{"*ast.Text(a)"}, " a comment",
		),
		Entry("a comment with no value in front of it",
			"X = # a comment\n", nil, " a comment",
		),
		Entry("a comment written against the assignment operator",
			"X =# a comment\n", nil, " a comment",
		),
		Entry("a comment written against a reference",
			"X := $(FOO)# a comment\n", []string{"*ast.VarRef($(FOO))"}, " a comment",
		),
	)

	// A backslash escapes the pound, so make reads the character as text and
	// the line has no comment on it. The source text is kept verbatim, the
	// escape included, because nothing else the parser reads is unescaped.
	DescribeTable("should read an escaped pound as text",
		func(input, shape string) {
			v := onlyVariable(parseOne(input))

			Expect(v.Comment).To(BeNil())
			Expect(v.Value).To(HaveLen(1))
			Expect(exprShape(v.Value[0])).To(Equal([]string{shape}))
		},
		Entry("an escaped pound between two pieces of text", "X := a\\#b\n", `*ast.Text(a\#b)`),
		Entry("an escaped pound ending the value", "X := a\\#\n", `*ast.Text(a\#)`),
		Entry("an escaped pound opening the value", "X := \\#b\n", `*ast.Text(\#b)`),
	)

	It("should read an escaped pound in a prerequisite as text", func() {
		rule := onlyRule(parseOne("target: a\\#b\n"))

		Expect(rule.Comment).To(BeNil())
		Expect(rule.PreReqs).To(HaveLen(1))
		Expect(exprShape(rule.PreReqs[0])).To(Equal([]string{`*ast.Text(a\#b)`}))
	})

	// The backslash of "a\\#" escapes the backslash rather than the pound, so
	// make reads the pound that follows it as the start of a comment.
	It("should begin a comment after an escaped backslash", func() {
		v := onlyVariable(parseOne("X := a\\\\# a comment\n"))

		Expect(v.Value).To(HaveLen(1))
		Expect(exprShape(v.Value[0])).To(Equal([]string{`*ast.Text(a\\)`}))
		Expect(variableComment(v).Text).To(Equal(" a comment"))
	})

	// make hands a recipe line to the shell without removing anything from it,
	// so a pound written there is a character of the command rather than the
	// start of a comment.
	DescribeTable("should keep a pound written in a recipe",
		func(input, body string) {
			rule := onlyRule(parseOne(input))

			Expect(rule.Comment).To(BeNil())
			Expect(rule.Recipes).To(HaveLen(1))
			Expect(recipeAt(rule, 0).Value).To(Equal(body))
		},
		Entry("a pound written against the text before it",
			"target:\n\t@echo x#y\n", "@echo x#y",
		),
		Entry("a pound written apart from the text before it",
			"target:\n\t@echo x # y\n", "@echo x # y",
		),
		// A recipe introduced by a semicolon keeps the blanks between the
		// semicolon and the command, the way every semicolon recipe does.
		Entry("a pound in a recipe introduced by a semicolon",
			"target: a ; @echo hi # not a comment\n", " @echo hi # not a comment",
		),
	)
})
