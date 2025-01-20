package make

import (
	"fmt"
	"io"
	"reflect"

	"github.com/unmango/go-make/ast"
)

type Writer struct {
	io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w}
}

func (w *Writer) WriteLine() (n int, err error) {
	return io.WriteString(w, "\n")
}

func (w *Writer) WritFile(f *ast.File) (n int, err error) {
	if f == nil {
		err = fmt.Errorf("f was nil")
		return
	}

	for _, r := range f.Rules {
		if c, err := w.WriteRule(r); err != nil {
			return 0, err
		} else {
			n += c
		}
	}

	return
}

func (w *Writer) WritePreReqList(l *ast.PreReqList) (n int, err error) {
	if l == nil {
		return
	}

	if c, err := io.WriteString(w, " "); err != nil {
		return 0, err
	} else {
		n += c
	}

	for i, p := range l.List {
		if c, err := w.WriteFileName(p); err != nil {
			return 0, err
		} else {
			n += c
		}

		if i+1 >= len(l.List) {
			continue
		}

		if c, err := io.WriteString(w, " "); err != nil {
			return 0, err
		} else {
			n += c
		}
	}

	return
}

func (w *Writer) WriteRecipe(r *ast.Recipe) (n int, err error) {
	if r == nil {
		err = fmt.Errorf("r was nil")
		return
	}

	return fmt.Fprintf(w, "%s%s", r.Tok, r.Text)
}

func (w *Writer) WriteRule(r *ast.Rule) (n int, err error) {
	if r == nil {
		err = fmt.Errorf("r was nil")
		return
	}

	if r.Targets == nil || len(r.Targets.List) == 0 {
		return 0, fmt.Errorf("no targets")
	}

	if c, err := w.WriteTargetList(r.Targets); err != nil {
		return 0, err
	} else {
		n += c
	}

	if x, err := w.WritePreReqList(r.PreReqs); err != nil {
		return 0, err
	} else {
		n += x
	}

	if x, err := w.WriteLine(); err != nil {
		return 0, err
	} else {
		n += x
	}

	for _, r := range r.Recipes {
		if c, err := w.WriteRecipe(r); err != nil {
			return 0, err
		} else {
			n += c
		}
	}

	if len(r.Recipes) > 0 {
		if x, err := w.WriteLine(); err != nil {
			return 0, err
		} else {
			return n + x, nil
		}
	}

	return
}

func (w *Writer) WriteTargetList(l *ast.TargetList) (n int, err error) {
	if l == nil {
		return
	}

	for i, t := range l.List {
		if c, err := w.WriteFileName(t); err != nil {
			return 0, err
		} else {
			n += c
		}

		if i+1 >= len(l.List) {
			continue
		}

		if c, err := io.WriteString(w, " "); err != nil {
			return 0, err
		} else {
			n += c
		}
	}

	if c, err := io.WriteString(w, ":"); err != nil {
		return 0, err
	} else {
		n += c
	}

	return
}

func (w *Writer) WriteFileName(f ast.FileName) (n int, err error) {
	if f == nil {
		err = fmt.Errorf("f was nil")
		return
	}

	switch node := f.(type) {
	case *ast.LiteralFileName:
		return fmt.Fprint(w, node.Name)
	default:
		return 0, fmt.Errorf("unsupported node type: %s", reflect.TypeOf(f))
	}
}
