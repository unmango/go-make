package expr

import (
	"fmt"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/token"
)

// Copy returns a deep copy of expr positioned at pos.
func Copy(pos token.Pos, expr ast.Expr) ast.Expr {
	c := clone(expr)
	SetPos(pos, c)
	return c
}

// SetPos moves expr, and every node it contains, so that expr begins at pos.
// It returns [End] of the moved expression.
func SetPos(pos token.Pos, expr ast.Expr) token.Pos {
	switch n := expr.(type) {
	case *ast.Text:
		n.ValuePos = pos
	case *ast.QuotedExpr:
		// A nil Value is an empty quoted expression, `""`, so the closing quote
		// follows the opening one.
		q := quoteLen(n.Quote)
		n.Open = pos
		if n.Value == nil {
			n.Close = pos + q
		} else {
			n.Close = SetPos(pos+q, n.Value)
		}
	case *ast.VarRef:
		n.Dollar = pos
	case *ast.Recipe:
		n.PrefixPos = pos
		n.ValuePos = pos + length(n.Prefix)
	default:
		panic(fmt.Sprintf("builder/expr: SetPos: unsupported expression type %T", expr))
	}

	return End(expr)
}

// End returns the position immediately after the last character the printer
// writes for expr.
//
// The positions [ast.Node] reports do not account for every character the
// printer writes, so the layout builders compute their own ends. A [ast.VarRef]
// ends after its closing brace, and a [ast.QuotedExpr] after its closing quote.
func End(expr ast.Expr) token.Pos {
	switch n := expr.(type) {
	case *ast.Text:
		return n.ValuePos + token.Pos(len(n.Value))
	case *ast.QuotedExpr:
		return n.Close + quoteLen(n.Quote)
	case *ast.VarRef:
		end := n.Dollar + 1 + token.Pos(len(n.Name))
		if n.Open != token.ILLEGAL {
			end += length(n.Open)
		}
		if n.Close != token.ILLEGAL {
			end += length(n.Close)
		}
		return end
	case *ast.Recipe:
		return n.ValuePos + token.Pos(len(n.Value))
	default:
		panic(fmt.Sprintf("builder/expr: End: unsupported expression type %T", expr))
	}
}

func clone(expr ast.Expr) ast.Expr {
	switch n := expr.(type) {
	case *ast.Text:
		c := *n
		return &c
	case *ast.QuotedExpr:
		c := *n
		if n.Value != nil {
			c.Value = clone(n.Value)
		}
		return &c
	case *ast.VarRef:
		c := *n
		return &c
	case *ast.Recipe:
		c := *n
		return &c
	default:
		panic(fmt.Sprintf("builder/expr: Copy: unsupported expression type %T", expr))
	}
}

// quoteLen is the width of the quote characters surrounding a [ast.QuotedExpr].
func quoteLen(quote token.Token) token.Pos {
	if quote == token.ILLEGAL {
		return 0
	}

	return length(quote)
}

func length(tok token.Token) token.Pos {
	return token.Pos(len(tok.String()))
}
