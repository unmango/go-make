package obj

import (
	"fmt"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/builder/expr"
	"github.com/unmango/go-make/token"
)

// Copy returns a deep copy of obj laid out at pos.
func Copy(pos token.Pos, obj ast.Obj) ast.Obj {
	c := clone(obj)
	SetPos(pos, c)
	return c
}

// SetPos lays obj out beginning at pos, assigning every position the printer
// reads when it writes obj. It returns [End] of the moved object.
func SetPos(pos token.Pos, obj ast.Obj) token.Pos {
	switch n := obj.(type) {
	case *ast.Rule:
		return setRulePos(pos, n)
	case *ast.Recipe:
		// A recipe reaches an object list through the body of a conditional
		// written inside a rule.
		expr.SetPos(pos, n)

		return End(n)
	case *ast.CommentGroup:
		for _, c := range n.List {
			c.Pound = pos
			pos = c.End() + 1 // '\n'
		}

		return pos
	case *ast.Variable:
		pos = expr.SetPos(pos, n.Name)
		n.OpPos = pos + 1 // ' '
		pos = n.OpPos + length(n.Op)
		for _, v := range n.Value {
			pos = expr.SetPos(pos+1, v) // ' '
		}

		return pos + 1 // '\n'
	case *ast.BadObj:
		n.From = pos
		n.To = pos + token.Pos(len(n.Text))

		return n.To + 1 // '\n'
	case *ast.IfBlock:
		pos = setDirPos(pos, n.Directive) + 1 // '\n'
		for _, o := range n.Text {
			pos = SetPos(pos, o)
		}
		for _, e := range n.Else {
			e.Else = pos
			pos += length(token.ELSE)
			if e.Condition != nil {
				pos = setDirPos(pos+1, e.Condition) // ' '
			}
			pos++ // '\n'
			for _, o := range e.Text {
				pos = SetPos(pos, o)
			}
		}
		n.Endif = pos

		return End(n)
	default:
		panic(fmt.Sprintf("builder/obj: SetPos: unsupported object type %T", obj))
	}
}

// Pos returns the position of the first character of obj. Unlike
// [ast.Node.Pos] it does not panic when obj has no children.
func Pos(obj ast.Obj) token.Pos {
	switch n := obj.(type) {
	case *ast.Rule:
		if len(n.Targets) == 0 {
			return n.Colon
		}
	case *ast.CommentGroup:
		if len(n.List) == 0 {
			return token.NoPos
		}
	}

	return obj.Pos()
}

// End returns the position of the first character following the newline that
// terminates obj.
func End(obj ast.Obj) token.Pos {
	switch n := obj.(type) {
	case *ast.Rule:
		return ruleEnd(n)
	case *ast.Recipe:
		return expr.End(n) + 1 // '\n'
	case *ast.CommentGroup:
		if len(n.List) == 0 {
			return token.NoPos
		}

		return n.List[len(n.List)-1].End() + 1 // '\n'
	case *ast.Variable:
		if len(n.Value) > 0 {
			return expr.End(n.Value[len(n.Value)-1]) + 1 // '\n'
		}

		return n.OpPos + length(n.Op) + 1 // '\n'
	case *ast.BadObj:
		return n.To + 1 // '\n'
	case *ast.IfBlock:
		return n.Endif + length(token.ENDIF) + 1 // '\n'
	default:
		panic(fmt.Sprintf("builder/obj: End: unsupported object type %T", obj))
	}
}

// setDirPos lays a conditional directive out beginning at pos and returns the
// position immediately after the last character the printer writes for it.
func setDirPos(pos token.Pos, dir ast.IfDir) token.Pos {
	switch n := dir.(type) {
	case *ast.IfdefDir:
		n.TokPos = pos
		return expr.SetPos(pos+length(n.Tok)+1, n.VarName) // ' '
	case *ast.IfeqDir:
		n.TokPos = pos
		pos += length(n.Tok)
		if n.Open.IsValid() {
			n.Open = pos + 1                         // ' '
			n.Comma = expr.SetPos(n.Open+1, n.Arg1)  // '('
			n.Close = expr.SetPos(n.Comma+2, n.Arg2) // ',' and ' '
			return n.Close + length(token.RPAREN)
		}

		pos = expr.SetPos(pos+1, n.Arg1)  // ' '
		return expr.SetPos(pos+1, n.Arg2) // ' '
	default:
		panic(fmt.Sprintf("builder/obj: SetPos: unsupported directive type %T", dir))
	}
}

func clone(obj ast.Obj) ast.Obj {
	switch n := obj.(type) {
	case *ast.Rule:
		return cloneRule(n)
	case *ast.Recipe:
		c := *n

		return &c
	case *ast.CommentGroup:
		c := &ast.CommentGroup{}
		for _, comment := range n.List {
			cc := *comment
			c.List = append(c.List, &cc)
		}

		return c
	case *ast.Variable:
		c := *n
		c.Name = expr.Copy(n.Name.Pos(), n.Name)
		c.Value = nil
		for _, v := range n.Value {
			c.Value = append(c.Value, expr.Copy(v.Pos(), v))
		}

		return &c
	case *ast.BadObj:
		c := *n

		return &c
	case *ast.IfBlock:
		c := &ast.IfBlock{Endif: n.Endif}
		c.Directive = cloneDir(n.Directive)
		for _, o := range n.Text {
			c.Text = append(c.Text, clone(o))
		}
		for _, e := range n.Else {
			ec := &ast.ElseBlock{Else: e.Else}
			if e.Condition != nil {
				ec.Condition = cloneDir(e.Condition)
			}
			for _, o := range e.Text {
				ec.Text = append(ec.Text, clone(o))
			}
			c.Else = append(c.Else, ec)
		}

		return c
	default:
		panic(fmt.Sprintf("builder/obj: Copy: unsupported object type %T", obj))
	}
}

func cloneDir(dir ast.IfDir) ast.IfDir {
	switch n := dir.(type) {
	case *ast.IfdefDir:
		c := *n
		c.VarName = expr.Copy(n.VarName.Pos(), n.VarName)
		return &c
	case *ast.IfeqDir:
		c := *n
		c.Arg1 = expr.Copy(n.Arg1.Pos(), n.Arg1)
		c.Arg2 = expr.Copy(n.Arg2.Pos(), n.Arg2)
		return &c
	default:
		panic(fmt.Sprintf("builder/obj: Copy: unsupported directive type %T", dir))
	}
}

// setRulePos lays r out beginning at pos, assigning every position the printer
// reads when it writes r. It returns [End] of the moved rule.
//
// Targets and prerequisites are separated by a single space, the colon
// immediately follows the last target, and a single space separates the colon
// from the first prerequisite.
func setRulePos(pos token.Pos, r *ast.Rule) token.Pos {
	for _, t := range r.Targets {
		pos = expr.SetPos(pos, t) + 1 // ' '
	}

	if len(r.Targets) > 0 {
		r.Colon = pos - 1 // the colon follows the last target directly
	} else {
		r.Colon = pos
	}

	pos = r.Colon + 2 // ':' and ' '
	for _, p := range r.PreReqs {
		pos = expr.SetPos(pos, p) + 1 // ' '
	}

	if r.Pipe.IsValid() {
		r.Pipe = pos
		pos = r.Pipe + 2 // '|' and ' '
	}
	for _, p := range r.OrderPreReqs {
		pos = expr.SetPos(pos, p) + 1 // ' '
	}

	for i, recipe := range r.Recipes {
		switch n := recipe.(type) {
		case *ast.Recipe:
			if i == 0 && n.Prefix == token.SEMI {
				pos-- // a semicolon recipe follows the rule on the same line
			}

			pos = expr.SetPos(pos, n) + 1 // '\n'
		case *ast.IfBlock:
			pos = SetPos(pos, n)
		default:
			panic(fmt.Sprintf("builder/obj: SetPos: unsupported recipe type %T", recipe))
		}
	}

	return pos
}

// ruleEnd returns the position of the first character following the newline
// that terminates r.
func ruleEnd(r *ast.Rule) token.Pos {
	if n := len(r.Recipes); n > 0 {
		switch last := r.Recipes[n-1].(type) {
		case *ast.Recipe:
			return expr.End(last) + 1 // '\n'
		case *ast.IfBlock:
			return End(last)
		default:
			panic(fmt.Sprintf("builder/obj: End: unsupported recipe type %T", last))
		}
	}
	if n := len(r.OrderPreReqs); n > 0 {
		return expr.End(r.OrderPreReqs[n-1]) + 1
	}
	if n := len(r.PreReqs); n > 0 {
		return expr.End(r.PreReqs[n-1]) + 1
	}
	if r.Pipe.IsValid() {
		return r.Pipe + 2
	}

	return r.Colon + 2
}

func cloneRule(r *ast.Rule) *ast.Rule {
	c := &ast.Rule{Colon: r.Colon, Pipe: r.Pipe}
	for _, t := range r.Targets {
		c.Targets = append(c.Targets, expr.Copy(t.Pos(), t))
	}
	for _, p := range r.PreReqs {
		c.PreReqs = append(c.PreReqs, expr.Copy(p.Pos(), p))
	}
	for _, p := range r.OrderPreReqs {
		c.OrderPreReqs = append(c.OrderPreReqs, expr.Copy(p.Pos(), p))
	}
	for _, recipe := range r.Recipes {
		switch n := recipe.(type) {
		case *ast.Recipe:
			c.Recipes = append(c.Recipes, expr.Copy(n.Pos(), n).(*ast.Recipe))
		case *ast.IfBlock:
			c.Recipes = append(c.Recipes, clone(n).(*ast.IfBlock))
		default:
			panic(fmt.Sprintf("builder/obj: Copy: unsupported recipe type %T", recipe))
		}
	}

	return c
}

func length(tok token.Token) token.Pos {
	return token.Pos(len(tok.String()))
}
