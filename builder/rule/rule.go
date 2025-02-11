package rule

import "github.com/unmango/go-make/builder"

func WithTarget(f builder.ExprFunc) builder.RuleFunc {
	return func(r builder.Rule) {
		r.Target(f)
	}
}

func WithTextTarget(text string) builder.RuleFunc {
	return WithTarget(func(e builder.Expr) {
		e.Text(text)
	})
}

func WithVarRefTarget(name string) builder.RuleFunc {
	return WithTarget(func(e builder.Expr) {
		e.VarRef(name)
	})
}
