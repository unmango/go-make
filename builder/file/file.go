package file

import "github.com/unmango/go-make/builder"

func WithRule(e builder.ExprBuilder, rs ...builder.RuleBuilder) builder.FileBuilder {
	return func(f builder.File) {
		f.Rule(e, rs...)
	}
}
