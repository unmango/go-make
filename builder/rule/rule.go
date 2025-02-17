package rule

import (
	"github.com/unmango/go-make/builder/build"
)

func WithTarget(f func(build.Expr)) func(build.Rule) {
	return func(r build.Rule) {
		r.Target(f)
	}
}

func WithTextTarget(text string) func(build.Rule) {
	return WithTarget(func(e build.Expr) {
		e.Text(text)
	})
}

func WithVarRefTarget(name string) func(build.Rule) {
	return WithTarget(func(e build.Expr) {
		e.VarRef(name)
	})
}
