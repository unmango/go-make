package expr

import "github.com/unmango/go-make/builder"

func Text(text string) builder.ExprFunc {
	return func(b builder.Expr) {
		b.Text(text)
	}
}

func VarRef(name string) builder.ExprFunc {
	return func(b builder.Expr) {
		b.VarRef(name)
	}
}
