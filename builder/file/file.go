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
		pos = fn(pos, file) + 1 // \n
	}

	file.FileEnd = pos
	return file
}

func Rule(builder ...builder.Rule) builder.File {
	return func(p token.Pos, f *ast.File) token.Pos {
		r := rule.New(p, builder...)
		f.Contents = append(f.Contents, r)
		return r.End()
	}
}

func InsertRule(f *ast.File, i int, builder ...builder.Rule) {
	var pos token.Pos
	contents := []ast.Obj{}

	for j, c := range f.Contents {
		switch {
		case j < i:
			contents = append(contents, c)
			pos = c.End() + 1
		case j == i:
			r := rule.New(pos, builder...)
			contents = append(contents, r)
			pos = r.End() + 1
			fallthrough
		case j > i:
			obj := obj.Copy(pos, c)
			contents = append(contents, obj)
			pos = obj.End() + 1
		}
	}

	f.Contents = contents
	f.FileEnd = pos - 1
}
