package file

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/builder/build"
	"github.com/unmango/go-make/builder/rule"
	"github.com/unmango/go-make/token"
)

func Rule(fn func(build.Rule), fs ...func(build.Rule)) func(build.File) {
	return func(f build.File) {
		f.AddRule(fn, fs...)
	}
}

func AddRule(builder ...builder.Rule) builder.File {
	return func(p token.Pos, f *ast.File) {
		f.Contents = append(f.Contents, rule.New(p, builder...))
	}
}
