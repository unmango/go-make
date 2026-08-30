package rule_test

import (
	"bytes"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
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

var _ = Describe("Rule", func() {
	Describe("New", func() {
		It("should work", func() {
			r := rule.New(1)

			Expect(r).NotTo(BeNil())
		})

		It("should apply the given builders", func() {
			var expected *ast.Rule

			r := rule.New(1, func(p token.Pos, r *ast.Rule) token.Pos {
				expected = r
				return p
			})

			Expect(r).To(BeIdenticalTo(expected))
		})

		It("should position the colon", func() {
			r := rule.New(1)

			Expect(r.Colon).To(Equal(token.Pos(1)))
		})

		It("should position the colon after the last target", func() {
			r := rule.New(1, rule.TextTarget("test"))

			Expect(r.Colon).To(Equal(token.Pos(5)))
		})

		It("should advance the position for each builder", func() {
			r := rule.New(1,
				rule.TextTarget("a"),
				rule.TextTarget("b"),
				rule.TextPreReq("c"),
			)

			Expect(printed(r)).To(Equal("a b: c\n"))
		})

		DescribeTable("should print the built rule",
			func(expected string, builder ...builder.Rule) {
				Expect(printed(rule.New(1, builder...))).To(Equal(expected))
			},
			Entry("a single target", "a:\n", rule.TextTarget("a")),
			Entry("multiple targets", "a b:\n",
				rule.TextTarget("a"),
				rule.TextTarget("b"),
			),
			Entry("a single pre-requisite", "a: b\n",
				rule.TextTarget("a"),
				rule.TextPreReq("b"),
			),
			Entry("multiple pre-requisites", "a: b c\n",
				rule.TextTarget("a"),
				rule.TextPreReq("b"),
				rule.TextPreReq("c"),
			),
			Entry("a var ref target", "$(FOO): b\n",
				rule.Target(func(p token.Pos) ast.Expr {
					return &ast.VarRef{
						Dollar: p,
						Open:   token.LPAREN,
						Name:   "FOO",
						Close:  token.RPAREN,
					}
				}),
				rule.TextPreReq("b"),
			),
		)
	})

	Describe("Copy", func() {
		It("should copy an empty rule", func() {
			r := rule.New(1)

			actual := rule.Copy(2, r)

			Expect(actual).To(Equal(&ast.Rule{
				Colon: 2,
			}))
		})

		It("should copy a rule with a target", func() {
			r := rule.New(1, rule.TextTarget("test"))

			actual := rule.Copy(2, r)

			Expect(actual).To(Equal(&ast.Rule{
				Targets: []ast.Expr{
					&ast.Text{Value: "test", ValuePos: 2},
				},
				Colon: 6,
			}))
		})

		It("should copy a rule with multiple targets", func() {
			r := rule.New(1,
				rule.TextTarget("test"),
				rule.TextTarget("test2"),
			)

			actual := rule.Copy(2, r)

			Expect(actual).To(Equal(&ast.Rule{
				Targets: []ast.Expr{
					&ast.Text{Value: "test", ValuePos: 2},
					&ast.Text{Value: "test2", ValuePos: 7},
				},
				Colon: 12,
			}))
		})

		It("should copy a rule with a pre-requisite", func() {
			r := rule.New(1, rule.TextPreReq("test"))

			actual := rule.Copy(2, r)

			Expect(actual).To(Equal(&ast.Rule{
				Colon: 2,
				PreReqs: []ast.Expr{
					&ast.Text{Value: "test", ValuePos: 4},
				},
			}))
		})

		DescribeTable("should print the copied rule",
			func(expected string, r *ast.Rule) {
				Expect(printed(rule.Copy(1, r))).To(Equal(expected))
			},
			Entry("targets and pre-requisites", "a b: c\n", &ast.Rule{
				Targets: []ast.Expr{
					&ast.Text{Value: "a"},
					&ast.Text{Value: "b"},
				},
				PreReqs: []ast.Expr{
					&ast.Text{Value: "c"},
				},
			}),
			Entry("order-only pre-requisites", "target: | prereq\n", &ast.Rule{
				Targets:      []ast.Expr{&ast.Text{Value: "target"}},
				Pipe:         token.Pos(1),
				OrderPreReqs: []ast.Expr{&ast.Text{Value: "prereq"}},
			}),
			Entry("multiple order-only pre-requisites", "target: | a b\n", &ast.Rule{
				Targets: []ast.Expr{&ast.Text{Value: "target"}},
				Pipe:    token.Pos(1),
				OrderPreReqs: []ast.Expr{
					&ast.Text{Value: "a"},
					&ast.Text{Value: "b"},
				},
			}),
			Entry("mixed pre-requisites", "target: prereq | a b\n", &ast.Rule{
				Targets: []ast.Expr{&ast.Text{Value: "target"}},
				PreReqs: []ast.Expr{&ast.Text{Value: "prereq"}},
				Pipe:    token.Pos(1),
				OrderPreReqs: []ast.Expr{
					&ast.Text{Value: "a"},
					&ast.Text{Value: "b"},
				},
			}),
			Entry("recipes", "target:\n\tcat thing\n\ttouch thing\n", &ast.Rule{
				Targets: []ast.Expr{&ast.Text{Value: "target"}},
				Recipes: []*ast.Recipe{
					{Text: ast.Text{Value: "cat thing"}, Prefix: token.TAB},
					{Text: ast.Text{Value: "touch thing"}, Prefix: token.TAB},
				},
			}),
			Entry("custom prefix recipes", "target:\n>cat thing\n>touch thing\n", &ast.Rule{
				Targets: []ast.Expr{&ast.Text{Value: "target"}},
				Recipes: []*ast.Recipe{
					{Text: ast.Text{Value: "cat thing"}, Prefix: token.TEXT, PrefixLit: ">"},
					{Text: ast.Text{Value: "touch thing"}, Prefix: token.TEXT, PrefixLit: ">"},
				},
			}),
			Entry("a var ref target", "$(FOO) bar: baz\n", &ast.Rule{
				Targets: []ast.Expr{
					&ast.VarRef{Open: token.LPAREN, Name: "FOO", Close: token.RPAREN},
					&ast.Text{Value: "bar"},
				},
				PreReqs: []ast.Expr{&ast.Text{Value: "baz"}},
			}),
			Entry("a quoted target", "'foo': bar\n", &ast.Rule{
				Targets: []ast.Expr{
					&ast.QuotedExpr{Quote: token.APOS, Value: &ast.Text{Value: "foo"}},
				},
				PreReqs: []ast.Expr{&ast.Text{Value: "bar"}},
			}),
		)

		It("should not alias the copied rule", func() {
			r := rule.New(1, rule.TextTarget("test"))

			actual := rule.Copy(2, r)

			Expect(actual.Targets[0]).NotTo(BeIdenticalTo(r.Targets[0]))
			Expect(r.Targets[0].Pos()).To(Equal(token.Pos(1)))
		})

		It("should copy a rule with multiple pre-requisites", func() {
			r := rule.New(1,
				rule.TextPreReq("test"),
				rule.TextPreReq("test2"),
			)

			actual := rule.Copy(2, r)

			Expect(actual).To(Equal(&ast.Rule{
				Colon: 2,
				PreReqs: []ast.Expr{
					&ast.Text{Value: "test", ValuePos: 4},
					&ast.Text{Value: "test2", ValuePos: 9},
				},
			}))
		})
	})
})
