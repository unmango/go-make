package file

import (
	"slices"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/builder/rule"
	"github.com/unmango/go-make/token"
)

func New(pos token.Pos, builder ...builder.File) *ast.File {
	file := &ast.File{FileStart: pos}
	for _, fn := range builder {
		fn(pos, file)
	}

	return file
}

func Rule(builder ...builder.Rule) builder.File {
	return func(p token.Pos, f *ast.File) {
		f.Contents = append(f.Contents, rule.New(p, builder...))
	}
}

func InsertRule(i int, builder ...builder.Rule) builder.File {
	return func(p token.Pos, f *ast.File) {
		var r ast.Obj = rule.New(p, builder...)
		f.Contents = slices.Insert(f.Contents, i, r)

		for _, n := range f.Contents[i:] {
			switch n := n.(type) {
			case *ast.Rule:
				rule.RePos(1, n)
			}
		}
	}
}
