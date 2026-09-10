package make_test

import (
	"fmt"

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

// onlyIfBlock returns the conditional block a file holds, failing when the
// file holds anything else.
func onlyIfBlock(f *ast.File) *ast.IfBlock {
	GinkgoHelper()
	Expect(f.Contents).To(HaveLen(1))
	block, ok := f.Contents[0].(*ast.IfBlock)
	Expect(ok).To(BeTrue(), "expected an *ast.IfBlock, got %T", f.Contents[0])

	return block
}

// dirComment returns the comment ending the line of a conditional directive,
// failing when the directive carries none.
func dirComment(d ast.IfDir) *ast.Comment {
	GinkgoHelper()
	var c *ast.Comment
	switch n := d.(type) {
	case *ast.IfeqDir:
		c = n.Comment
	case *ast.IfdefDir:
		c = n.Comment
	default:
		Fail(fmt.Sprintf("expected a conditional directive, got %T", d))
	}
	Expect(c).NotTo(BeNil(), "expected the directive to carry a comment")

	return c
}

// A conditional directive shares its line with the comment ending it, the way
// a target line and an assignment do. The comment is a field of the node
// owning the line rather than an object of the body, because a body holding it
// would print it on a line of its own and move every branch of the block down
// one.
var _ = Describe("Comment ending a conditional line", func() {
	DescribeTable("should end a directive line",
		func(input, text string) {
			block := onlyIfBlock(parseOne(input))

			Expect(dirComment(block.Directive).Text).To(Equal(text))
			// The body holds what was written below the directive and nothing
			// else. A comment left here is the shape this field replaces.
			Expect(block.Text).To(HaveLen(1))
			Expect(block.Text[0]).To(BeAssignableToTypeOf(&ast.Rule{}))
		},
		Entry("an ifeq", "ifeq (a,b) # a comment\ntarget:\nendif\n", " a comment"),
		Entry("an ifneq", "ifneq (a,b) # a comment\ntarget:\nendif\n", " a comment"),
		Entry("an ifdef", "ifdef FOO # a comment\ntarget:\nendif\n", " a comment"),
		Entry("an ifndef", "ifndef FOO # a comment\ntarget:\nendif\n", " a comment"),
		Entry("a comment written against the directive",
			"ifeq (a,b)# a comment\ntarget:\nendif\n", " a comment",
		),
		Entry("a comment written against an ifdef variable name",
			"ifdef FOO# a comment\ntarget:\nendif\n", " a comment",
		),
		Entry("a comment with no text", "ifeq (a,b) #\ntarget:\nendif\n", ""),
		Entry("a comment with no leading space",
			"ifeq (a,b) #a comment\ntarget:\nendif\n", "a comment",
		),
		Entry("a quoted condition", "ifeq \"a\" \"b\" # a comment\ntarget:\nendif\n", " a comment"),
		Entry("an empty condition", "ifeq (,) # a comment\ntarget:\nendif\n", " a comment"),
	)

	It("should keep the variable name of an ifdef written against a comment", func() {
		block := onlyIfBlock(parseOne("ifdef FOO# a comment\ntarget:\nendif\n"))

		dir, ok := block.Directive.(*ast.IfdefDir)
		Expect(ok).To(BeTrue(), "expected an *ast.IfdefDir, got %T", block.Directive)
		Expect(exprShape(dir.VarName)).To(Equal([]string{"*ast.Text(FOO)"}))
	})

	It("should end a bare else line", func() {
		block := onlyIfBlock(parseOne("ifeq (a,b)\ntargetA:\nelse # a comment\ntargetB:\nendif\n"))

		Expect(block.Else).To(HaveLen(1))
		Expect(block.Else[0].Condition).To(BeNil())
		Expect(block.Else[0].Comment).NotTo(BeNil())
		Expect(block.Else[0].Comment.Text).To(Equal(" a comment"))
		Expect(block.Else[0].Text).To(HaveLen(1))
	})

	// An `else ifeq` ends its line with the condition, so the condition holds
	// the comment ending it. The else and the condition are two nodes on one
	// line, and the comment goes to the one it was written after, the same way
	// it goes to the directive of a block rather than to the block.
	DescribeTable("should give an else-if comment to its condition",
		func(input string) {
			block := onlyIfBlock(parseOne(input))

			Expect(block.Else).To(HaveLen(1))
			Expect(block.Else[0].Comment).To(BeNil())
			Expect(dirComment(block.Else[0].Condition).Text).To(Equal(" a comment"))
		},
		Entry("an else ifeq", "ifeq (a,b)\ntargetA:\nelse ifeq (c,d) # a comment\ntargetB:\nendif\n"),
		Entry("an else ifdef", "ifdef A\ntargetA:\nelse ifdef B # a comment\ntargetB:\nendif\n"),
	)

	DescribeTable("should end an endif line",
		func(input, text string) {
			block := onlyIfBlock(parseOne(input))

			Expect(block.EndifComment).NotTo(BeNil())
			Expect(block.EndifComment.Text).To(Equal(text))
		},
		Entry("a comment written apart from the endif",
			"ifeq (a,b)\ntarget:\nendif # a comment\n", " a comment",
		),
		Entry("a comment written against the endif",
			"ifeq (a,b)\ntarget:\nendif# a comment\n", " a comment",
		),
	)

	It("should end every line of a block", func() {
		block := onlyIfBlock(parseOne(
			"ifeq (a,b) # one\ntargetA:\nelse # two\ntargetB:\nendif # three\n",
		))

		Expect(dirComment(block.Directive).Text).To(Equal(" one"))
		Expect(block.Else).To(HaveLen(1))
		Expect(block.Else[0].Comment).NotTo(BeNil())
		Expect(block.Else[0].Comment.Text).To(Equal(" two"))
		Expect(block.EndifComment).NotTo(BeNil())
		Expect(block.EndifComment.Text).To(Equal(" three"))
	})

	It("should end the lines of a nested block", func() {
		outer := onlyIfBlock(parseOne(
			"ifdef A # outer\nifeq (b,c) # inner\ntarget:\nendif # inner end\nendif # outer end\n",
		))

		Expect(dirComment(outer.Directive).Text).To(Equal(" outer"))
		Expect(outer.EndifComment.Text).To(Equal(" outer end"))
		Expect(outer.Text).To(HaveLen(1))
		inner, ok := outer.Text[0].(*ast.IfBlock)
		Expect(ok).To(BeTrue(), "expected an *ast.IfBlock, got %T", outer.Text[0])
		Expect(dirComment(inner.Directive).Text).To(Equal(" inner"))
		Expect(inner.EndifComment.Text).To(Equal(" inner end"))
	})

	// A comment on the directive line and one on the line below it are written
	// in two places, so they are two nodes. The one below opens the body and
	// stays the [ast.CommentGroup] it has always been.
	It("should leave a comment written below the directive in the body", func() {
		block := onlyIfBlock(parseOne(
			"ifeq (a,b) # on the directive\n# on its own line\ntarget:\nendif\n",
		))

		Expect(dirComment(block.Directive).Text).To(Equal(" on the directive"))
		Expect(block.Text).To(HaveLen(2))
		group, ok := block.Text[0].(*ast.CommentGroup)
		Expect(ok).To(BeTrue(), "expected an *ast.CommentGroup, got %T", block.Text[0])
		Expect(group.List).To(HaveLen(1))
		Expect(group.List[0].Text).To(Equal(" on its own line"))
	})

	// A backslash escapes the pound, so the line has no comment on it and the
	// text keeps the escape verbatim.
	It("should read an escaped pound in an ifdef as text", func() {
		block := onlyIfBlock(parseOne("ifdef FOO\\#bar\ntarget:\nendif\n"))

		dir, ok := block.Directive.(*ast.IfdefDir)
		Expect(ok).To(BeTrue(), "expected an *ast.IfdefDir, got %T", block.Directive)
		Expect(dir.Comment).To(BeNil())
		Expect(exprShape(dir.VarName)).To(Equal([]string{`*ast.Text(FOO\#bar)`}))
	})

	It("should read an escaped pound in an ifeq argument as text", func() {
		block := onlyIfBlock(parseOne("ifeq (a\\#b,c) # a comment\ntarget:\nendif\n"))

		dir, ok := block.Directive.(*ast.IfeqDir)
		Expect(ok).To(BeTrue(), "expected an *ast.IfeqDir, got %T", block.Directive)
		Expect(exprShape(dir.Arg1)).To(Equal([]string{`*ast.Text(a\#b)`}))
		Expect(dir.Comment.Text).To(Equal(" a comment"))
	})

	// A conditional written under a target line belongs to the rule, so the
	// comments on its lines have to survive the move into Rule.Recipes.
	It("should end the lines of a block inside a rule body", func() {
		rule := onlyRule(parseOne("target:\nifdef VERBOSE # a comment\n\techo hi\nendif # done\n"))

		Expect(rule.Recipes).To(HaveLen(1))
		block, ok := rule.Recipes[0].(*ast.IfBlock)
		Expect(ok).To(BeTrue(), "expected an *ast.IfBlock, got %T", rule.Recipes[0])
		Expect(dirComment(block.Directive).Text).To(Equal(" a comment"))
		Expect(block.EndifComment.Text).To(Equal(" done"))
	})
})
