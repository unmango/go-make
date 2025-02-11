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
	It("should walk nil", func() {
		v := &visitor{}

		ast.Walk(v, nil)

		Expect(v.nodes).To(HaveExactElements(nil))
	})

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

	Describe("Inspect", func() {
		It("should inspect nil", func() {
			var nodes []ast.Node

			ast.Inspect(nil, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(nil))
		})

		It("should inspect an empty file", func() {
			var nodes []ast.Node
			file := &ast.File{}

			ast.Inspect(file, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(file))
		})

		It("should inspect a file with an empty rule", func() {
			var nodes []ast.Node
			rule := &ast.Rule{}
			file := &ast.File{Contents: []ast.Obj{rule}}

			ast.Inspect(file, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(file, rule))
		})

		It("should inspect an empty rule", func() {
			var nodes []ast.Node
			rule := &ast.Rule{}

			ast.Inspect(rule, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(rule))
		})

		It("should inspect a rule with targets", func() {
			var nodes []ast.Node
			t1 := &ast.Text{}
			t2 := &ast.Text{}
			rule := &ast.Rule{Targets: []ast.Expr{t1, t2}}

			ast.Inspect(rule, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(rule, t1, t2))
		})

		It("should inspect a rule with prereqs", func() {
			var nodes []ast.Node
			p1 := &ast.Text{}
			p2 := &ast.Text{}
			rule := &ast.Rule{PreReqs: []ast.Expr{p1, p2}}

			ast.Inspect(rule, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(rule, p1, p2))
		})

		It("should inspect a rule with targets and prereqs", func() {
			var nodes []ast.Node
			t1 := &ast.Text{}
			t2 := &ast.Text{}
			p1 := &ast.Text{}
			p2 := &ast.Text{}
			rule := &ast.Rule{
				Targets: []ast.Expr{t1, t2},
				PreReqs: []ast.Expr{p1, p2},
			}

			ast.Inspect(rule, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(rule, t1, t2, p1, p2))
		})

		It("should inspect a rule with order-only prereqs", func() {
			var nodes []ast.Node
			p1 := &ast.Text{}
			p2 := &ast.Text{}
			rule := &ast.Rule{OrderPreReqs: []ast.Expr{p1, p2}}

			ast.Inspect(rule, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(rule, p1, p2))
		})

		It("should inspect a rule with prereqs and order-only prereqs", func() {
			var nodes []ast.Node
			p1 := &ast.Text{}
			p2 := &ast.Text{}
			p3 := &ast.Text{}
			p4 := &ast.Text{}
			rule := &ast.Rule{
				PreReqs:      []ast.Expr{p3, p4},
				OrderPreReqs: []ast.Expr{p1, p2},
			}

			ast.Inspect(rule, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(rule, p3, p4, p1, p2))
		})

		It("should inspect a rule with targets and recipes", func() {
			var nodes []ast.Node
			t1 := ast.Text{}
			t2 := &ast.Text{}
			r1 := &ast.Recipe{Text: t1}
			rule := &ast.Rule{
				Targets: []ast.Expr{t2},
				Recipes: []*ast.Recipe{r1},
			}

			ast.Inspect(rule, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(rule, t2, r1, &t1))
		})

		It("should inspect a recipe", func() {
			var nodes []ast.Node
			t1 := ast.Text{}
			r1 := &ast.Recipe{Text: t1}

			ast.Inspect(r1, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(r1, &t1))
		})

		It("should inspect text", func() {
			var nodes []ast.Node
			t1 := &ast.Text{}

			ast.Inspect(t1, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(t1))
		})

		It("should inspect a variable reference", func() {
			var nodes []ast.Node
			v1 := &ast.VarRef{}

			ast.Inspect(v1, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(v1))
		})

		It("should inspect a comment group", func() {
			var nodes []ast.Node
			c := &ast.Comment{}
			cg := &ast.CommentGroup{
				List: []*ast.Comment{c},
			}

			ast.Inspect(cg, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(cg, c))
		})

		It("should inspect a comment group", func() {
			var nodes []ast.Node
			c := &ast.Comment{}

			ast.Inspect(c, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(c))
		})

		It("should inspect a quoted expression", func() {
			var nodes []ast.Node
			t1 := &ast.Text{}
			q := &ast.QuotedExpr{Value: t1}

			ast.Inspect(q, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(q, t1))
		})

		It("should inspect an empty variable", func() {
			var nodes []ast.Node
			t1 := &ast.Text{}
			v1 := &ast.Variable{Name: t1}

			ast.Inspect(v1, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(v1, t1))
		})

		It("should inspect a variable", func() {
			var nodes []ast.Node
			t1 := &ast.Text{}
			t2 := &ast.Text{}
			t3 := &ast.Text{}
			v1 := &ast.Variable{
				Name:  t1,
				Value: []ast.Expr{t2, t3},
			}

			ast.Inspect(v1, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(v1, t1, t2, t3))
		})

		It("should inspect an ifeq directive", func() {
			var nodes []ast.Node
			t1 := &ast.Text{}
			t2 := &ast.Text{}
			d := &ast.IfeqDir{Arg1: t1, Arg2: t2}

			ast.Inspect(d, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(d, t1, t2))
		})

		It("should inspect an ifdef directive", func() {
			var nodes []ast.Node
			t1 := &ast.Text{}
			d := &ast.IfdefDir{VarName: t1}

			ast.Inspect(d, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(d, t1))
		})

		It("should inspect an else block", func() {
			var nodes []ast.Node
			t1 := &ast.Text{}
			v1 := &ast.Variable{}
			d := &ast.IfdefDir{VarName: t1}
			e := &ast.ElseBlock{Condition: d, Text: []ast.Obj{v1}}

			ast.Inspect(e, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(e, d, t1, v1))
		})

		It("should inspect an if block", func() {
			var nodes []ast.Node
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

			ast.Inspect(f, func(n ast.Node) bool {
				nodes = append(nodes, n)
				return true
			})

			Expect(nodes).To(HaveExactElements(f, d2, t2, v2, e, d1, t1, v1))
		})
	})

	Describe("Preorder", func() {
		It("should order nil", func() {
			nodes := ast.Preorder(nil)

			Expect(nodes).To(BeEmpty())
		})

		It("should sequence an empty file", func() {
			file := &ast.File{}

			nodes := ast.Preorder(file)

			Expect(nodes).To(HaveExactElements(file))
		})

		It("should sequence a file with an empty rule", func() {
			rule := &ast.Rule{}
			file := &ast.File{Contents: []ast.Obj{rule}}

			nodes := ast.Preorder(file)

			Expect(nodes).To(HaveExactElements(file, rule))
		})

		It("should sequence an empty rule", func() {
			rule := &ast.Rule{}

			nodes := ast.Preorder(rule)

			Expect(nodes).To(HaveExactElements(rule))
		})

		It("should sequence a rule with targets", func() {
			t1 := &ast.Text{}
			t2 := &ast.Text{}
			rule := &ast.Rule{Targets: []ast.Expr{t1, t2}}

			nodes := ast.Preorder(rule)

			Expect(nodes).To(HaveExactElements(rule, t1, t2))
		})

		It("should sequence a rule with prereqs", func() {
			p1 := &ast.Text{}
			p2 := &ast.Text{}
			rule := &ast.Rule{PreReqs: []ast.Expr{p1, p2}}

			nodes := ast.Preorder(rule)

			Expect(nodes).To(HaveExactElements(rule, p1, p2))
		})

		It("should sequence a rule with targets and prereqs", func() {
			t1 := &ast.Text{}
			t2 := &ast.Text{}
			p1 := &ast.Text{}
			p2 := &ast.Text{}
			rule := &ast.Rule{
				Targets: []ast.Expr{t1, t2},
				PreReqs: []ast.Expr{p1, p2},
			}

			nodes := ast.Preorder(rule)

			Expect(nodes).To(HaveExactElements(rule, t1, t2, p1, p2))
		})

		It("should sequence a rule with order-only prereqs", func() {
			p1 := &ast.Text{}
			p2 := &ast.Text{}
			rule := &ast.Rule{OrderPreReqs: []ast.Expr{p1, p2}}

			nodes := ast.Preorder(rule)

			Expect(nodes).To(HaveExactElements(rule, p1, p2))
		})

		It("should sequence a rule with prereqs and order-only prereqs", func() {
			p1 := &ast.Text{}
			p2 := &ast.Text{}
			p3 := &ast.Text{}
			p4 := &ast.Text{}
			rule := &ast.Rule{
				PreReqs:      []ast.Expr{p3, p4},
				OrderPreReqs: []ast.Expr{p1, p2},
			}

			nodes := ast.Preorder(rule)

			Expect(nodes).To(HaveExactElements(rule, p3, p4, p1, p2))
		})

		It("should sequence a rule with targets and recipes", func() {
			t1 := ast.Text{}
			t2 := &ast.Text{}
			r1 := &ast.Recipe{Text: t1}
			rule := &ast.Rule{
				Targets: []ast.Expr{t2},
				Recipes: []*ast.Recipe{r1},
			}

			nodes := ast.Preorder(rule)

			Expect(nodes).To(HaveExactElements(rule, t2, r1, &t1))
		})

		It("should sequence a recipe", func() {
			t1 := ast.Text{}
			recipe := &ast.Recipe{Text: t1}

			nodes := ast.Preorder(recipe)

			Expect(nodes).To(HaveExactElements(recipe, &t1))
		})

		It("should sequence text", func() {
			text := &ast.Text{}

			nodes := ast.Preorder(text)

			Expect(nodes).To(HaveExactElements(text))
		})

		It("should sequence a variable reference", func() {
			varref := &ast.VarRef{}

			nodes := ast.Preorder(varref)

			Expect(nodes).To(HaveExactElements(varref))
		})

		It("should sequence a comment group", func() {
			c := &ast.Comment{}
			cg := &ast.CommentGroup{
				List: []*ast.Comment{c},
			}

			nodes := ast.Preorder(cg)

			Expect(nodes).To(HaveExactElements(cg, c))
		})

		It("should sequence a comment group", func() {
			c := &ast.Comment{}

			nodes := ast.Preorder(c)

			Expect(nodes).To(HaveExactElements(c))
		})

		It("should sequence a quoted expression", func() {
			t1 := &ast.Text{}
			expr := &ast.QuotedExpr{Value: t1}

			nodes := ast.Preorder(expr)

			Expect(nodes).To(HaveExactElements(expr, t1))
		})

		It("should sequence an empty variable", func() {
			t1 := &ast.Text{}
			variable := &ast.Variable{Name: t1}

			nodes := ast.Preorder(variable)

			Expect(nodes).To(HaveExactElements(variable, t1))
		})

		It("should sequence a variable", func() {
			t1 := &ast.Text{}
			t2 := &ast.Text{}
			t3 := &ast.Text{}
			variable := &ast.Variable{
				Name:  t1,
				Value: []ast.Expr{t2, t3},
			}

			nodes := ast.Preorder(variable)

			Expect(nodes).To(HaveExactElements(variable, t1, t2, t3))
		})

		It("should sequence an ifeq directive", func() {
			t1 := &ast.Text{}
			t2 := &ast.Text{}
			directive := &ast.IfeqDir{Arg1: t1, Arg2: t2}

			nodes := ast.Preorder(directive)

			Expect(nodes).To(HaveExactElements(directive, t1, t2))
		})

		It("should sequence an ifdef directive", func() {
			t1 := &ast.Text{}
			directive := &ast.IfdefDir{VarName: t1}

			nodes := ast.Preorder(directive)

			Expect(nodes).To(HaveExactElements(directive, t1))
		})

		It("should sequence an else block", func() {
			t1 := &ast.Text{}
			v1 := &ast.Variable{}
			d := &ast.IfdefDir{VarName: t1}
			block := &ast.ElseBlock{Condition: d, Text: []ast.Obj{v1}}

			nodes := ast.Preorder(block)

			Expect(nodes).To(HaveExactElements(block, d, t1, v1))
		})

		It("should sequence an if block", func() {
			t1 := &ast.Text{}
			t2 := &ast.Text{}
			v1 := &ast.Variable{}
			v2 := &ast.Variable{}
			d1 := &ast.IfdefDir{VarName: t1}
			d2 := &ast.IfdefDir{VarName: t2}
			e := &ast.ElseBlock{Condition: d1, Text: []ast.Obj{v1}}
			block := &ast.IfBlock{
				Directive: d2,
				Text:      []ast.Obj{v2},
				Else:      []*ast.ElseBlock{e},
			}

			nodes := ast.Preorder(block)

			Expect(nodes).To(HaveExactElements(block, d2, t2, v2, e, d1, t1, v1))
		})
	})
})
