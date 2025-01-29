package make

import (
	"fmt"
	"io"
	"reflect"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

type Writer struct {
	io.Writer
}

// NewWriter returns a new [Writer] writing to w, or w
// if w is already a [Writer]
func NewWriter(w io.Writer) *Writer {
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

func WriteFile(w io.Writer, f *ast.File) (n int, err error) {
	if f == nil {
		err = fmt.Errorf("f was nil")
		return
	}

	for _, r := range f.Decls {
		if c, err := WriteDecl(w, r); err != nil {
			return 0, err
		} else {
			n += c
		}
	}

	return
}

func WriteDecl(w io.Writer, decl ast.Decl) (n int, err error) {
	switch decl := decl.(type) {
	case *ast.Rule:
		return WriteRule(NewWriter(w), decl)
	}

	return // TODO
}

func WritePreReqList(w *Writer, l *ast.PreReqList) (n int, err error) {
	if l == nil {
		return
	}

	for i, p := range l.List {
		if c, err := WriteExpr(w, p); err != nil {
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

func WriteRecipe(w *Writer, r *ast.Recipe) (n int, err error) {
	if r == nil {
		err = fmt.Errorf("r was nil")
		return
	}

	return fmt.Fprintf(w, "%s%s\n", r.Tok, r.Text)
}

func WriteRule(w *Writer, r *ast.Rule) (n int, err error) {
	if r == nil {
		err = fmt.Errorf("r was nil")
		return
	}

	if r.Targets == nil || len(r.Targets.List) == 0 {
		return 0, fmt.Errorf("no targets")
	}

	if c, err := WriteTargetList(w, r.Targets); err != nil {
		return 0, err
	} else {
		n += c
	}

	if r.PreReqs != nil && len(r.PreReqs.List) > 0 {
		if c, err := io.WriteString(w, " "); err != nil {
			return 0, err
		} else {
			n += c
		}
	}

	if x, err := WritePreReqList(w, r.PreReqs); err != nil {
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
		if c, err := WriteRecipe(w, r); err != nil {
			return 0, err
		} else {
			n += c
		}
	}

	return
}

func WriteTargetList(w *Writer, l *ast.TargetList) (n int, err error) {
	if l == nil {
		return
	}

	for i, t := range l.List {
		if c, err := WriteExpr(w, t); err != nil {
			return 0, err
		} else {
			n += c
		}

		if i+1 >= len(l.List) {
			continue
		}

		if c, err := w.WriteSpace(); err != nil {
			return 0, err
		} else {
			n += c
		}
	}

	if c, err := w.WriteToken(token.COLON); err != nil {
		return 0, err
	} else {
		n += c
	}

	return
}

func WriteExpr(w *Writer, f ast.Expr) (n int, err error) {
	switch node := f.(type) {
	case *ast.Text:
		return w.WriteString(node.Value)
	default:
		return 0, fmt.Errorf("unsupported filename node: %v", reflect.TypeOf(f))
	}
}
