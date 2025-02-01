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

func WriteExpr(w io.Writer, e ast.Expr) (n int, err error) {
	return printer.Fprint(w, e)
}

func WritePreReqList(w io.Writer, l []ast.Expr) (n int, err error) {
	return printer.Fprint(w, l)
}

func WriteRecipe(w io.Writer, r *ast.Recipe) (n int, err error) {
	return printer.Fprint(w, r)
}

func WriteRule(w io.Writer, r *ast.Rule) (n int, err error) {
	return printer.Fprint(w, r)
}

func WriteTargetList(w io.Writer, l []ast.Expr) (n int, err error) {
	return printer.Fprint(w, l)
}

func WriteVar(w io.Writer, v *ast.Variable) (n int, err error) {
	return printer.Fprint(w, v)
}
