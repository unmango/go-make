package builder

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/build"
	"github.com/unmango/go-make/token"
)

func ApplyRule(node *ast.Rule, f func(build.Rule)) *ast.Rule {
	if node == nil {
		return nil
	}

	var pos token.Pos = 1
	if len(node.Targets) > 0 {
		pos = node.Pos()
	}

	// apply the transformations from f to node
	f(&rule{&builder{pos}, node})

	// re-write node into a new rule to fix positions
	r := &rule{&builder{pos}, &ast.Rule{}}
	copyRule(node, r)
	r.r.Colon = r.pos

	return r.r
}

func copyRule(rule *ast.Rule, b build.Rule) {
	for _, t := range rule.Targets {
		b.Target(func(e build.Expr) {
			switch n := t.(type) {
			case *ast.Text:
				e.Text(n.Value)
			case *ast.VarRef:
				e.VarRef(n.Name)
			}
		})
	}
}
