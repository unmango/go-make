package file

import "github.com/unmango/go-make/builder"

func WithRule(e builder.ExprFunc, rs ...builder.RuleFunc) builder.FileFunc {
	return func(f builder.File) {
		f.Rule(e, rs...)
	}
}
