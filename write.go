package make

import (
	"fmt"
	"io"
	"strings"
)

type Writer struct {
	w io.Writer
}

func NewWriter(w io.Writer) *Writer {
	return &Writer{w}
}

func (w *Writer) WriteLine() (n int, err error) {
	return io.WriteString(w.w, "\n")
}

func (w *Writer) WriteMakefile(m Makefile) (n int, err error) {
	for _, r := range m.Rules {
		if x, err := w.WriteRule(r); err != nil {
			return 0, err
		} else {
			n += x
		}
	}

	return
}

func (w *Writer) WritePreReq(p string) (n int, err error) {
	return io.WriteString(w.w, " "+p)
}

func (w *Writer) WritePreReqs(ps []string) (n int, err error) {
	for _, p := range ps {
		if x, err := w.WritePreReq(p); err != nil {
			return 0, err
		} else {
			n += x
		}
	}

	return
}

func (w *Writer) WriteRecipe(r string) (n int, err error) {
	return io.WriteString(w.w, "\t"+r)
}

func (w *Writer) WriteRecipes(rs []string) (n int, err error) {
	for _, p := range rs {
		if x, err := w.WriteRecipe(p); err != nil {
			return 0, err
		} else {
			n += x
		}
	}

	return
}

func (w *Writer) WriteRule(r Rule) (n int, err error) {
	if len(r.Target) == 0 {
		return 0, fmt.Errorf("no targets")
	}

	if n, err = w.WriteTargets(r.Target); err != nil {
		return
	}
	if x, err := w.WritePreReqs(r.PreReqs); err != nil {
		return 0, err
	} else {
		n += x
	}
	if x, err := w.WriteLine(); err != nil {
		return 0, err
	} else {
		n += x
	}
	if x, err := w.WriteRecipes(r.Recipe); err != nil {
		return 0, err
	} else {
		n += x
	}
	if len(r.Recipe) > 0 {
		if x, err := w.WriteLine(); err != nil {
			return 0, err
		} else {
			return n + x, nil
		}
	}

	return
}

func (w *Writer) WriteTarget(t string) (n int, err error) {
	return io.WriteString(w.w, t+":")
}

func (w *Writer) WriteTargets(ts []string) (n int, err error) {
	t := strings.Join(ts, " ")
	return io.WriteString(w.w, t+":")
}
