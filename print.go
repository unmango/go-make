package make

import (
	"fmt"
	"io"

	"github.com/unmango/go-make/ast"
)

func Fprint(w io.Writer, node ast.Node) (err error) {
	switch node := node.(type) {
	case *ast.LiteralFileName:
		_, err = io.WriteString(w, node.String())
	case *ast.TargetList:
		for _, t := range node.List {
			if err = Fprint(w, t); err != nil {
				return
			}
		}
		_, err = io.WriteString(w, ":")
	case *ast.PreReqList:
		for _, t := range node.List {
			if err = Fprint(w, t); err != nil {
				return
			}
		}
	case *ast.Recipe:
		if _, err = fmt.Fprint(w, node.Tok); err != nil {
			return err
		}
		if _, err = io.WriteString(w, node.Text); err != nil {
			return err
		}
		if _, err = fmt.Fprintln(w); err != nil {
			return err
		}
	case *ast.Rule:
		if err = Fprint(w, node.Targets); err != nil {
			return err
		}
		if _, err = fmt.Fprint(w, " "); err != nil {
			return err
		}
		if err = Fprint(w, node.PreReqs); err != nil {
			return err
		}
		if _, err = fmt.Fprintln(w); err != nil {
			return err
		}
		for _, r := range node.Recipes {
			if err = Fprint(w, r); err != nil {
				return err
			}
		}
	}

	return
}
