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
	Colon   token.Pos // position of ':' delimiting targets and prerequisites
	Pipe    token.Pos // position of '|' delimiting normal and order-only prerequisites
	Semi    token.Pos // position of ';' delimiting prerequisites and recipes
	Targets *TargetList
	PreReqs *PreReqList
	Recipes []*Recipe
}

// declNode implements Decl
func (*Rule) declNode() {}

// Pos implements Node
func (r *Rule) Pos() token.Pos {
	return r.Targets.Pos()
}

// End implements Node
func (r *Rule) End() token.Pos {
	return r.Recipes[len(r.Recipes)-1].End()
}

// A TargetList represents a list of Targets in a Rule.
type TargetList struct {
	List []Expr
}

// Add appends target to t.List
func (t *TargetList) Add(target Expr) {
	t.List = append(t.List, target)
}

// Pos implements Node
func (t *TargetList) Pos() token.Pos {
	return t.List[0].Pos()
}

// End implements Node
func (t *TargetList) End() token.Pos {
	return t.List[len(t.List)-1].End()
}

// A PreReqList represents all normal and order-only prerequisites in a Rule.
type PreReqList struct {
	Pipe token.Pos
	List []Expr
}

// Add appends prereq to p.List
func (p *PreReqList) Add(prereq Expr) {
	p.List = append(p.List, prereq)
}

// Pos implements Node
func (p *PreReqList) Pos() token.Pos {
	return p.List[0].Pos()
}

// End implements Node
func (p *PreReqList) End() token.Pos {
	return p.List[len(p.List)-1].End()
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
	Tok    token.Token // TAB, SEMI, or .RECIPEPREFIX
	TokPos token.Pos   // position of Tok
	Text   string      // recipe text excluding '\n'
}

// Pos implements Node
func (r *Recipe) Pos() token.Pos {
	return r.TokPos
}

// End implements Node
func (r *Recipe) End() token.Pos {
	return token.Pos(int(r.TokPos) + len(r.Text))
}

// An Variable represents a variable assignment.
type Variable struct {
	Name  Expr        // left-hand side of the assignment
	Op    token.Token // =, :=, ::=, :::=, !=, ?=
	OpPos token.Pos   // position of Tok
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
