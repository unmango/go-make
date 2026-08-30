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
		r := &ast.Rule{}
		f.Contents = append(f.Contents, r)

		// New accounts for the newline terminating the rule.
		return rule.Build(p, r, builder...) - 1
	}
}

// InsertRule builds a rule and inserts it at index i of f.Contents, laying out
// every object that follows it so the file still prints as valid make syntax.
//
// i is clamped to [0, len(f.Contents)], so an index past the end appends the
// rule to the file and a negative index prepends it. Objects preceding the
// insertion point keep their positions, objects following it are copied and
// re-positioned, which collapses any blank lines that separated them.
func InsertRule(f *ast.File, i int, builder ...builder.Rule) {
	i = min(max(i, 0), len(f.Contents))
	contents := make([]ast.Obj, 0, len(f.Contents)+1)

	var pos token.Pos
	switch {
	case i < len(f.Contents):
		pos = obj.Pos(f.Contents[i])
	case len(f.Contents) > 0:
		pos = obj.End(f.Contents[len(f.Contents)-1])
	default:
		pos = f.FileStart
	}

	contents = append(contents, f.Contents[:i]...)

	r := &ast.Rule{}
	pos = rule.Build(pos, r, builder...)
	contents = append(contents, r)

	for _, c := range f.Contents[i:] {
		c = obj.Copy(pos, c)
		contents = append(contents, c)
		pos = obj.End(c)
	}

	f.Contents = contents
	f.FileEnd = pos
}
