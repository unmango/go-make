package printer

import (
	"github.com/unmango/go-make/token"
	"github.com/unmango/go/fopt"
)

// PrintPosition prints node and reports the position the printer tracked
// while writing it. It exists so tests can assert on position bookkeeping
// that is not observable from the printed output.
func PrintPosition(node any, opts ...Op) (token.Position, error) {
	p := &printer{f: &token.File{}}
	fopt.ApplyAll(p, opts)

	err := p.printNode(node)
	return p.pos, err
}
