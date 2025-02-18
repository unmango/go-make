package rule

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/builder/build"
	"github.com/unmango/go-make/token"
)

func Target(f func(build.Expr)) func(build.Rule) {
	return func(r build.Rule) {
		r.Target(f)
	}
}

func TextTarget(text string) func(build.Rule) {
	return Target(func(e build.Expr) {
		e.Text(text)
	})
}

func VarRefTarget(name string) func(build.Rule) {
	return Target(func(e build.Expr) {
		e.VarRef(name)
	})
}

func New(pos token.Pos, builder ...builder.Rule) *ast.Rule {
	rule := &ast.Rule{}
	for _, fn := range builder {
		fn(pos, rule)
	}

	return rule
}

func AddTarget(expr func(token.Pos) ast.Expr) builder.Rule {
	return func(p token.Pos, r *ast.Rule) {
		r.Targets = append(r.Targets, expr(p))
	}
}

func AddTextTarget(builder ...builder.Text) builder.Rule {
	return AddTarget(func(p token.Pos) ast.Expr {
		return &ast.Text{} // text.New(p, builder...)
	})
}
