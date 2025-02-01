package writer

import (
	"fmt"
	"io"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

// A Writer is an [io.Writer] that facilitates
// writing arbitrary make [ast.Node]s
type Writer struct {
	io.Writer
}

// New returns a new [Writer] writing to w, or w
// if w is already a [Writer]
func New(w io.Writer) *Writer {
	if writer, ok := w.(*Writer); ok {
		return writer
	} else {
		return &Writer{w}
	}
}

func (w *Writer) WriteLine() (n int, err error) {
	return fmt.Fprintln(w)
}

func (w *Writer) WriteToken(tok token.Token) (n int, err error) {
	return fmt.Fprint(w, tok)
}

func (w *Writer) WriteSpace() (n int, err error) {
	return w.WriteString(" ")
}

func (w *Writer) WriteString(s string) (n int, err error) {
	return w.Write([]byte(s))
}

func (w *Writer) WriteExpr(e ast.Expr) (n int, err error) {
	switch node := e.(type) {
	case *ast.Text:
		return w.WriteString(node.Value)
	default:
		return
	}
}
