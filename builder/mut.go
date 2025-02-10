package builder

import "github.com/unmango/go-make/ast"

func AtRule(node *ast.Rule, f RuleBuilder) *ast.Rule {
	if node == nil || len(node.Targets) == 0 {
		return nil
	}

	r := &rule{&builder{node.Pos()}, node}
	f(r)
	return r.r
}
