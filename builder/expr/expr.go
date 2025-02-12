package expr

import (
	"github.com/unmango/go-make/builder/build"
)

func Text(text string) func(build.Expr) {
	return func(b build.Expr) {
		b.Text(text)
	}
}

func VarRef(name string) func(build.Expr) {
	return func(b build.Expr) {
		b.VarRef(name)
	}
}
