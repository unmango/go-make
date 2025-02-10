package expr

import "github.com/unmango/go-make/builder"

func Text(text string) builder.ExprBuilder {
	return func(b builder.Expr) {
		b.Text(text)
	}
}

func VarRef(name string) builder.ExprBuilder {
	return func(b builder.Expr) {
		b.VarRef(name)
	}
}
