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
	case *ast.FuncCall:
		setCallPos(pos, n)
	case *ast.Recipe:
		// A custom .RECIPEPREFIX is not a token, so the width of the prefix
		// comes from its source text.
		n.PrefixPos = pos
		n.ValuePos = pos + token.Pos(len(n.PrefixText()))
	case *ast.JuxtaposedExpr:
		// The parts are written with nothing between them, so each one begins
		// where the one before it ended. An empty juxtaposition occupies no
		// source, so it ends where it began.
		for _, part := range n.Parts {
			pos = SetPos(pos, part)
		}
		return pos
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
	case *ast.FuncCall:
		return n.ClosePos + length(n.Close)
	case *ast.Recipe:
		return n.ValuePos + token.Pos(len(n.Value))
	case *ast.JuxtaposedExpr:
		if len(n.Parts) == 0 {
			return token.NoPos
		}
		return End(n.Parts[len(n.Parts)-1])
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
	case *ast.FuncCall:
		c := *n
		if n.Name != nil {
			name := *n.Name
			c.Name = &name
		}
		c.Args = make([]*ast.FuncArg, len(n.Args))
		for i, a := range n.Args {
			arg := *a
			arg.Parts = make([]ast.Expr, len(a.Parts))
			for j, part := range a.Parts {
				arg.Parts[j] = clone(part)
			}
			c.Args[i] = &arg
		}
		c.Commas = append([]token.Pos(nil), n.Commas...)
		return &c
	case *ast.Recipe:
		c := *n
		return &c
	case *ast.JuxtaposedExpr:
		c := *n
		c.Parts = make([]ast.Expr, len(n.Parts))
		for i, part := range n.Parts {
			c.Parts[i] = clone(part)
		}
		return &c
	default:
		panic(fmt.Sprintf("builder/expr: Copy: unsupported expression type %T", expr))
	}
}

// setCallPos lays a function call out beginning at pos. The source layout of a
// call is not recoverable from the call alone, so the builder writes the
// canonical one: a single space between the name and the first argument, and
// between the parts of an argument, and nothing around the commas.
func setCallPos(pos token.Pos, call *ast.FuncCall) {
	call.Dollar = pos
	pos += 1 + length(call.Open) // '$' + Open
	if call.Name != nil {
		call.Name.ValuePos = pos
		pos += token.Pos(len(call.Name.Value))
	}
	if n := len(call.Args); n > 0 && len(call.Commas) != n-1 {
		call.Commas = make([]token.Pos, n-1)
	}
	for i, arg := range call.Args {
		if i == 0 {
			pos++ // ' '
		} else {
			call.Commas[i-1] = pos
			pos += length(token.COMMA)
		}

		arg.From = pos
		for j, part := range arg.Parts {
			if j > 0 {
				pos++ // ' '
			}

			pos = SetPos(pos, part)
		}
		arg.To = pos
	}
	call.ClosePos = pos
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
