package ast

import (
	"go/ast"

	"github.com/unmango/go-make/token"
)

type Node = ast.Node

// A File represents text content interpreted as the make syntax.
// Most commonly this is a Makefile, but could also be any file
// understood by make, i.e. include-me.mk
type File struct {
	FileStart, FileEnd token.Pos
	Comments           []*CommentGroup
	Rules              []*Rule
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

// Pos implements Node
func (r *Rule) Pos() token.Pos {
	return r.Targets.Pos()
}

// End implements Node
func (r *Rule) End() token.Pos {
	return r.Recipes[len(r.Recipes)-1].End()
}

// A TargetList represents a list of Targets in a single Rule.
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

// A PreReqList represents all normal and order-only prerequisites in a single Rule.
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

// A Expr represents any Node that can appear where a file name is expected.
type Expr interface {
	Node
	exprNode()
}

// A String represents plain text not interpreted by make.
type String struct {
	Text    string
	TextPos token.Pos
}

func (*String) exprNode() {}

// Pos implements Node
func (l *String) Pos() token.Pos {
	return l.TextPos
}

// End implements Node
func (l *String) End() token.Pos {
	return token.Pos(int(l.TextPos) + len(l.Text))
}

// String returns the literal identifier
func (l *String) String() string {
	return l.Text
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

// An Ident represents an identifier.
type Ident struct {
	Name    string
	NamePos token.Pos
}

// Pos implements Node
func (i *Ident) Pos() token.Pos {
	return i.NamePos
}

// End implements Node
func (i *Ident) End() token.Pos {
	return token.Pos(int(i.NamePos) + len(i.Name))
}

// String returns the literal identifier.
func (i *Ident) String() string {
	return i.Name
}

type AssignStmt struct {
	Lhs    []Expr
	Tok    token.Token
	TokPos token.Pos
	Rhs    []Expr
}
