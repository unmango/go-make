package make

import (
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
	}

	return
}
