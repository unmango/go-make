package ast

import (
	"go/ast"

	"github.com/unmango/go-make/token"
)

type Node = ast.Node

// All expression nodes implement the Expr interface.
type Expr interface {
	Node
	exprNode()
}

// All declaration nodes implement the Decl interface.
type Decl interface {
	Node
	declNode()
}

// A File represents text content interpreted as the make syntax.
// Most commonly this is a Makefile, but could also be any file
// understood by make, i.e. include-me.mk
type File struct {
	FileStart, FileEnd token.Pos
	Comments           []*CommentGroup
	Decls              []Decl // declarations; or nil
}

// Pos implements Node.
func (f *File) Pos() token.Pos {
	if len(f.Decls) > 0 {
		return f.Decls[0].Pos()
	} else {
		return f.FileStart
	}
}

// End implements Node.
func (f *File) End() token.Pos {
	if n := len(f.Decls); n > 0 {
		return f.Decls[n-1].End()
	} else {
		return f.FileEnd
	}
}

// A CommentGroup represents a sequence of comments with no other tokens and no empty lines between.
type CommentGroup struct {
	List []*Comment
}

// Pos implements Node
func (c *CommentGroup) Pos() token.Pos {
	return c.List[0].Pos()
}

// End implements Node
func (c *CommentGroup) End() token.Pos {
	return c.List[len(c.List)-1].End()
}

// TODO: Handle multi-line comments with '\' escaped newlines

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

// declNode implements Decl
func (*Rule) declNode() {}

// Pos implements Node
func (r *Rule) Pos() token.Pos {
	return r.Targets[0].Pos()
}

// End implements Node
func (r *Rule) End() token.Pos {
	return r.Recipes[len(r.Recipes)-1].End()
}

// Text represents a string of text that has no special meaning to make.
type Text struct {
	Value    string
	ValuePos token.Pos
}

// exprNode implements Expr
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

// declNode implements Decl
func (*Variable) declNode() {}

// Pos implements Node
func (s *Variable) Pos() token.Pos {
	return s.Name.Pos()
}

// End implements Node
func (s *Variable) End() token.Pos {
	if len(s.Value) > 0 {
		return s.Value[len(s.Value)-1].End()
	} else {
		return token.Pos(int(s.Pos()) + len(s.Op.String()))
	}
}
