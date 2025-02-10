package rule

import "github.com/unmango/go-make/builder"

func WithTarget(f builder.ExprBuilder) builder.RuleBuilder {
	return func(r builder.Rule) {
		r.Target(f)
	}
}

func WithTextTarget(text string) builder.RuleBuilder {
	return WithTarget(func(e builder.Expr) {
		e.Text(text)
	})
}

func WithVarRefTarget(name string) builder.RuleBuilder {
	return WithTarget(func(e builder.Expr) {
		e.VarRef(name)
	})
}
