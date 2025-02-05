package ast

import (
	"fmt"
	"go/ast"

	"github.com/unmango/go-make/token"
)

type Node = ast.Node

// All legal top-level make constructs implement the Obj interface.
type Obj interface {
	Node
	objNode()
}

// All directive nodes implements the Dir interface.
type Dir interface {
	Obj
	dirNode()
}

// All expression nodes implement the Expr interface.
type Expr interface {
	Node
	exprNode()
}

// All if* conditional directive nodes implement the IfDir interface.
type IfDir interface {
	Node
	ifDirNode()
}

// A File represents text content interpreted as the make syntax.
// Most commonly this is a Makefile, but could also be any file
// understood by make, i.e. include-me.mk
type File struct {
	FileStart, FileEnd token.Pos

	Contents []Obj // all file content
}

// Pos implements Node
func (f *File) Pos() token.Pos {
	if len(f.Contents) > 0 {
		return f.Contents[0].Pos()
	} else {
		return f.FileStart
	}
}

// End implements Node
func (f *File) End() token.Pos {
	if n := len(f.Contents); n > 0 {
		return f.Contents[n-1].End()
	} else {
		return f.FileEnd
	}
}

// A CommentGroup represents a sequence of comments with no other tokens and no empty lines between.
type CommentGroup struct {
	List []*Comment
}

func (*CommentGroup) objNode() {}

// Pos implements Node
func (c *CommentGroup) Pos() token.Pos {
	return c.List[0].Pos()
}

// End implements Node
func (c *CommentGroup) End() token.Pos {
	return c.List[len(c.List)-1].End()
}

// A comment represents a single comment starting with '#'
type Comment struct {
	Pound token.Pos // position of '#' starting the comment
	Text  string    // comment text, excluding '\n'
}

// Pos implements Node
func (c *Comment) Pos() token.Pos {
	return c.Pound
}

// End implements Node
func (c *Comment) End() token.Pos {
	return token.Pos(int(c.Pound) + len(c.Text))
}

// A Rule represents the Recipes and PreRequisites required to build Targets. [Rule Syntax]
//
// [Rule Syntax]: https://www.gnu.org/software/make/manual/html_node/Rule-Syntax.html
type Rule struct {
	Targets      []Expr    // rule targets
	Colon        token.Pos // position of ':' separating targets and prerequisites
	PreReqs      []Expr    // rule pre-requisites
	Pipe         token.Pos // position of '|' separating normal and order-only prerequisites
	OrderPreReqs []Expr    // order-only pre-requisites
	Recipes      []*Recipe // rule recipe lines
}

func (*Rule) objNode() {}

// Pos implements Node
func (r *Rule) Pos() token.Pos {
	return r.Targets[0].Pos()
}

// End implements Node
func (r *Rule) End() token.Pos {
	if n := len(r.Recipes); n > 0 {
		return r.Recipes[n-1].End()
	}
	if n := len(r.OrderPreReqs); n > 0 {
		return r.OrderPreReqs[n-1].End()
	}
	if n := len(r.PreReqs); n > 0 {
		return r.PreReqs[n-1].End()
	}

	return r.Colon + 1
}

// Text represents a string of text that has no special meaning to make.
type Text struct {
	Value    string
	ValuePos token.Pos
}

func (*Text) exprNode() {}

// Pos implements Node
func (l *Text) Pos() token.Pos {
	return l.ValuePos
}

// End implements Node
func (l *Text) End() token.Pos {
	return token.Pos(int(l.ValuePos) + len(l.Value))
}

// String returns the literal identifier
func (l *Text) String() string {
	return l.Value
}

// QuotedExpr represents an expression enclosed in quotes.
type QuotedExpr struct {
	Quote token.Token // ' or "
	Open  token.Pos   // position of the opening quote
	Value Expr        // position of the inner expression
	Close token.Pos   // position of the closing quote
}

func (*QuotedExpr) exprNode() {}

// Pos implements Node
func (l *QuotedExpr) Pos() token.Pos {
	return l.Open
}

// End implements Node
func (l *QuotedExpr) End() token.Pos {
	return l.Close
}

// String returns the quoted expression
func (l *QuotedExpr) String() string {
	quote := l.Quote.String()
	return fmt.Sprint(quote, l.Value, quote)
}

// VarRef represents a variable reference.
type VarRef struct {
	Dollar token.Pos   // position of '$'
	Open   token.Token // opening token, '(', '{', or ILLEGAL if len(Name) == 1
	Name   string      // variable identifier
	Close  token.Token // closing token, ')', '}', or ILLEGAL if len(Name) == 1
}

func (*VarRef) exprNode() {}

// Pos implements Node
func (v *VarRef) Pos() token.Pos {
	return v.Dollar
}

// End implements Node
func (v *VarRef) End() token.Pos {
	if n := len(v.Name); n == 1 {
		return v.Dollar + 1
	} else {
		// '$' + '{' + len(v.Name) + '}'
		return token.Pos(int(v.Dollar) + 2 + len(v.Name))
	}
}

// String implements fmt.Stringer
func (v *VarRef) String() string {
	if len(v.Name) == 1 {
		return "$" + v.Name
	} else {
		return fmt.Sprint("$", v.Open, v.Name, v.Close)
	}
}

// A Recipe represents a line of text to be passed to the shell to build a Target.
type Recipe struct {
	Text                  // recipe text excluding '\n'
	Prefix    token.Token // TAB, SEMI, or .RECIPEPREFIX
	PrefixPos token.Pos   // position of Tok
}

// Pos implements Node
func (r *Recipe) Pos() token.Pos {
	return r.PrefixPos
}

// End implements Node
func (r *Recipe) End() token.Pos {
	return token.Pos(int(r.PrefixPos) + len(r.Value))
}

// An Variable represents a make variable.
type Variable struct {
	Name  Expr        // left-hand side of the assignment
	Op    token.Token // =, :=, ::=, :::=, !=, ?=
	OpPos token.Pos   // position of Op
	Value []Expr      // right-hand side of the assignment
}

func (*Variable) objNode() {}

// Pos implements Node
func (s *Variable) Pos() token.Pos {
	return s.Name.Pos()
}

// End implements Node
func (s *Variable) End() token.Pos {
	if len(s.Value) > 0 {
		return s.Value[len(s.Value)-1].End()
	} else {
		return token.Pos(int(s.OpPos) + len(s.Op.String()))
	}
}

// DefDir represents a define directive
type DefDir struct {
	TokPos  token.Pos   // position of 'define'
	VarName Expr        // variable name
	Op      token.Token // -, :=, ::=, :::=, +=, ?= if it exists; ILLEGAL otherwise
	OpPos   token.Pos   // position of Op, if it exists
	Value   []Expr      // value expressions
	Endef   token.Pos   // position of 'endef'
}

func (*DefDir) objNode() {}
func (*DefDir) dirNode() {}

// Pos implements Node
func (d *DefDir) Pos() token.Pos {
	return d.TokPos
}

// End implements Node
func (d *DefDir) End() token.Pos {
	return d.Endef + 5 // pos + len("endef")
}

// DefDir represents an undefine directive
type UndefDir struct {
	TokPos  token.Pos // position of 'undefine'
	VarName Expr      // variable name
}

func (*UndefDir) objNode() {}
func (*UndefDir) dirNode() {}

// Pos implements Node
func (d *UndefDir) Pos() token.Pos {
	return d.TokPos
}

// End implements Node
func (d *UndefDir) End() token.Pos {
	return d.VarName.End()
}

// IfBlock represents a conditional directive and its parts.
type IfBlock struct {
	Directive IfDir        // conditional directive
	Text      []Obj        // text-if-true
	Else      []*ElseBlock // else directive blocks
	Endif     token.Pos    // position of ENDIF
}

func (*IfBlock) objNode() {}
func (*IfBlock) dirNode() {}

// Pos implements Node
func (b *IfBlock) Pos() token.Pos {
	return b.Directive.Pos()
}

// End implements Node
func (b *IfBlock) End() token.Pos {
	return b.Endif + 5 // pos + len("endif")
}

// ElseBlock represents and `else` clause in a conditional directive.
type ElseBlock struct {
	Else      token.Pos // position of ELSE
	Condition IfDir     // condition, if it exists; nil otherwise
	Text      []Obj     // text-if-true when a condition exists; text-if-false otherwise
}

// Pos implements Node
func (b *ElseBlock) Pos() token.Pos {
	return b.Else
}

// End implements Node
func (b *ElseBlock) End() token.Pos {
	if n := len(b.Text); n > 0 {
		return b.Text[n-1].End()
	} else if b.Condition != nil {
		return b.Condition.End()
	} else {
		return b.Else + 4 // pos + len("else")
	}
}

// IfeqDir represents a conditional directive block using `ifeq` or `ifneq`.
type IfeqDir struct {
	Tok    token.Token // IFEQ or IFNEQ
	TokPos token.Pos   // position of Tok
	Open   token.Pos   // position of '(', if it exists
	Arg1   Expr        // first argument in the condition
	Comma  token.Pos   // position of ',', if it exists
	Arg2   Expr        // second argument in the condition
	Close  token.Pos   // position of ')', if it exists
}

func (*IfeqDir) ifDirNode() {}

// Pos implements Node
func (d *IfeqDir) Pos() token.Pos {
	return d.TokPos
}

// End implements node
func (d *IfeqDir) End() token.Pos {
	if d.Close.IsValid() {
		return d.Close + 1 // pos + len(')')
	} else {
		return d.Arg2.End()
	}
}

// IfeqDir represents a conditional directive block using `ifdef` or `ifndef`.
type IfdefDir struct {
	Tok     token.Token // IFDEF or IFNDEF
	TokPos  token.Pos   // position of Tok
	VarName Expr        // variable-name
}

func (*IfdefDir) ifDirNode() {}

// Pos implements Node
func (d *IfdefDir) Pos() token.Pos {
	return d.TokPos
}

// End implements node
func (d *IfdefDir) End() token.Pos {
	return d.VarName.End()
}
