package rule

import (
	"github.com/unmango/go-make/builder/build"
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
