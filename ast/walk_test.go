package ast_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/unmango/go-make/ast"
)

type visitor struct {
	nodes []ast.Node
}

// Visit implements ast.Visitor.
func (v *visitor) Visit(node ast.Node) (w ast.Visitor) {
	v.nodes = append(v.nodes, node)
	return v
}

var _ = Describe("Walk", func() {
	It("should walk an empty file", func() {
		v := &visitor{}
		file := &ast.File{}

		ast.Walk(v, file)

		Expect(v.nodes).To(HaveExactElements(file))
	})

	It("should walk a file with an empty rule", func() {
		v := &visitor{}
		rule := &ast.Rule{}
		file := &ast.File{Contents: []ast.Obj{rule}}

		ast.Walk(v, file)

		Expect(v.nodes).To(HaveExactElements(file, rule))
	})

	It("should walk an empty rule", func() {
		v := &visitor{}
		rule := &ast.Rule{}

		ast.Walk(v, rule)

		Expect(v.nodes).To(HaveExactElements(rule))
	})

	It("should walk a rule with targets", func() {
		v := &visitor{}
		t1 := &ast.Text{}
		t2 := &ast.Text{}
		rule := &ast.Rule{Targets: []ast.Expr{t1, t2}}

		ast.Walk(v, rule)

		Expect(v.nodes).To(HaveExactElements(rule, t1, t2))
	})

	It("should walk a rule with prereqs", func() {
		v := &visitor{}
		p1 := &ast.Text{}
		p2 := &ast.Text{}
		rule := &ast.Rule{PreReqs: []ast.Expr{p1, p2}}

		ast.Walk(v, rule)

		Expect(v.nodes).To(HaveExactElements(rule, p1, p2))
	})

	It("should walk a rule with targets and prereqs", func() {
		v := &visitor{}
		t1 := &ast.Text{}
		t2 := &ast.Text{}
		p1 := &ast.Text{}
		p2 := &ast.Text{}
		rule := &ast.Rule{
			Targets: []ast.Expr{t1, t2},
			PreReqs: []ast.Expr{p1, p2},
		}

		ast.Walk(v, rule)

		Expect(v.nodes).To(HaveExactElements(rule, t1, t2, p1, p2))
	})

	It("should walk a rule with order-only prereqs", func() {
		v := &visitor{}
		p1 := &ast.Text{}
		p2 := &ast.Text{}
		rule := &ast.Rule{OrderPreReqs: []ast.Expr{p1, p2}}

		ast.Walk(v, rule)

		Expect(v.nodes).To(HaveExactElements(rule, p1, p2))
	})

	It("should walk a rule with prereqs and order-only prereqs", func() {
		v := &visitor{}
		p1 := &ast.Text{}
		p2 := &ast.Text{}
		p3 := &ast.Text{}
		p4 := &ast.Text{}
		rule := &ast.Rule{
			PreReqs:      []ast.Expr{p3, p4},
			OrderPreReqs: []ast.Expr{p1, p2},
		}

		ast.Walk(v, rule)

		Expect(v.nodes).To(HaveExactElements(rule, p3, p4, p1, p2))
	})

	It("should walk a rule with targets and recipes", func() {
		v := &visitor{}
		t1 := ast.Text{}
		t2 := &ast.Text{}
		r1 := &ast.Recipe{Text: t1}
		rule := &ast.Rule{
			Targets: []ast.Expr{t2},
			Recipes: []*ast.Recipe{r1},
		}

		ast.Walk(v, rule)

		Expect(v.nodes).To(HaveExactElements(rule, t2, r1, &t1))
	})

	It("should walk a recipe", func() {
		v := &visitor{}
		t1 := ast.Text{}
		r1 := &ast.Recipe{Text: t1}

		ast.Walk(v, r1)

		Expect(v.nodes).To(HaveExactElements(r1, &t1))
	})

	It("should walk text", func() {
		v := &visitor{}
		t1 := &ast.Text{}

		ast.Walk(v, t1)

		Expect(v.nodes).To(HaveExactElements(t1))
	})

	It("should walk a variable reference", func() {
		v := &visitor{}
		v1 := &ast.VarRef{}

		ast.Walk(v, v1)

		Expect(v.nodes).To(HaveExactElements(v1))
	})

	It("should walk a comment group", func() {
		v := &visitor{}
		c := &ast.Comment{}
		cg := &ast.CommentGroup{
			List: []*ast.Comment{c},
		}

		ast.Walk(v, cg)

		Expect(v.nodes).To(HaveExactElements(cg, c))
	})

	It("should walk a comment group", func() {
		v := &visitor{}
		c := &ast.Comment{}

		ast.Walk(v, c)

		Expect(v.nodes).To(HaveExactElements(c))
	})

	It("should walk a quoted expression", func() {
		v := &visitor{}
		t1 := &ast.Text{}
		q := &ast.QuotedExpr{Value: t1}

		ast.Walk(v, q)

		Expect(v.nodes).To(HaveExactElements(q, t1))
	})

	It("should walk an empty variable", func() {
		v := &visitor{}
		t1 := &ast.Text{}
		v1 := &ast.Variable{Name: t1}

		ast.Walk(v, v1)

		Expect(v.nodes).To(HaveExactElements(v1, t1))
	})

	It("should walk a variable", func() {
		v := &visitor{}
		t1 := &ast.Text{}
		t2 := &ast.Text{}
		t3 := &ast.Text{}
		v1 := &ast.Variable{
			Name:  t1,
			Value: []ast.Expr{t2, t3},
		}

		ast.Walk(v, v1)

		Expect(v.nodes).To(HaveExactElements(v1, t1, t2, t3))
	})

	It("should walk an ifeq directive", func() {
		v := &visitor{}
		t1 := &ast.Text{}
		t2 := &ast.Text{}
		d := &ast.IfeqDir{Arg1: t1, Arg2: t2}

		ast.Walk(v, d)

		Expect(v.nodes).To(HaveExactElements(d, t1, t2))
	})

	It("should walk an ifdef directive", func() {
		v := &visitor{}
		t1 := &ast.Text{}
		d := &ast.IfdefDir{VarName: t1}

		ast.Walk(v, d)

		Expect(v.nodes).To(HaveExactElements(d, t1))
	})

	It("should walk an else block", func() {
		v := &visitor{}
		t1 := &ast.Text{}
		v1 := &ast.Variable{}
		d := &ast.IfdefDir{VarName: t1}
		e := &ast.ElseBlock{Condition: d, Text: []ast.Obj{v1}}

		ast.Walk(v, e)

		Expect(v.nodes).To(HaveExactElements(e, d, t1, v1))
	})

	It("should walk an if block", func() {
		v := &visitor{}
		t1 := &ast.Text{}
		t2 := &ast.Text{}
		v1 := &ast.Variable{}
		v2 := &ast.Variable{}
		d1 := &ast.IfdefDir{VarName: t1}
		d2 := &ast.IfdefDir{VarName: t2}
		e := &ast.ElseBlock{Condition: d1, Text: []ast.Obj{v1}}
		f := &ast.IfBlock{
			Directive: d2,
			Text:      []ast.Obj{v2},
			Else:      []*ast.ElseBlock{e},
		}

		ast.Walk(v, f)

		Expect(v.nodes).To(HaveExactElements(f, d2, t2, v2, e, d1, t1, v1))
	})
})
