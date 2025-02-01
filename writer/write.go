package writer

import (
	"io"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/printer"
)

func WriteFile(w io.Writer, f *ast.File) (n int, err error) {
	return printer.Fprint(w, f)
}

func Decl(w io.Writer, d ast.Decl) (n int, err error) {
	return printer.Fprint(w, d)
}

func WriteRule(w io.Writer, r *ast.Rule) (n int, err error) {
	return printer.Fprint(w, r)
}

func WriteVar(w io.Writer, v *ast.Variable) (n int, err error) {
	return printer.Fprint(w, v)
}
