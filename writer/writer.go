package writer

import (
	"fmt"
	"io"
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
