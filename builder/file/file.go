package file

import (
	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder"
	"github.com/unmango/go-make/builder/obj"
	"github.com/unmango/go-make/builder/rule"
	"github.com/unmango/go-make/token"
)

func New(pos token.Pos, builder ...builder.File) *ast.File {
	file := &ast.File{FileStart: pos}
	for _, fn := range builder {
		fn(pos, file)

		if n := len(file.Contents); n > 0 {
			pos = file.Contents[n-1].End()
		}
	}

	return file
}

func Rule(builder ...builder.Rule) builder.File {
	return func(p token.Pos, f *ast.File) {
		f.Contents = append(f.Contents, rule.New(p, builder...))
	}
}

// func InsertRule(i int, builder ...builder.Rule) builder.File {
// 	return func(p token.Pos, f *ast.File) {
// 		var r ast.Obj = rule.New(p, builder...)
// 		f.Contents = slices.Insert(f.Contents, i, r)

// 		for _, n := range f.Contents[i:] {
// 			switch n := n.(type) {
// 			case *ast.Rule:
// 				rule.RePos(1, n)
// 			}
// 		}
// 	}
// }

func InsertRule(f *ast.File, i int, builder ...builder.Rule) {
	contents := []ast.Obj{}
	for j, c := range f.Contents {
		switch {
		case j == i:
			r := rule.New(69, builder...)
			contents = append(contents, r)
			fallthrough
		case j < i:
			contents = append(contents, c)
		case j > i:
			obj := obj.Copy(69, c)
			contents = append(contents, obj)
		}
	}

	f.Contents = contents
}
