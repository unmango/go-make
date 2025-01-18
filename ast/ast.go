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
	List []FileName
}

// Add appends target to t.List
func (t *TargetList) Add(target FileName) {
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
	List []FileName
}

// Pos implements Node
func (p *PreReqList) Pos() token.Pos {
	return p.List[0].Pos()
}

// End implements Node
func (p *PreReqList) End() token.Pos {
	return p.List[len(p.List)-1].End()
}

// A FileName represents any Node that can appear where a file name is expected.
type FileName interface {
	Node
	fileNameNode()
}

// A LiteralFileName represents a name identifier with no additional syntax.
type LiteralFileName struct {
	Name *Ident
}

func (*LiteralFileName) fileNameNode() {}

func (l *LiteralFileName) Pos() token.Pos {
	return l.Name.Pos()
}

func (l *LiteralFileName) End() token.Pos {
	return l.Name.End()
}

// A Recipe represents a line of text to be passed to the shell to build a Target.
type Recipe struct {
	Tok    token.Token // TAB or SEMI
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
