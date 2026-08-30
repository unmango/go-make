package ast

import (
	"fmt"
	"go/ast"
	"strings"

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

// A BadObj represents a line containing syntax the parser does not
// understand. It records the line verbatim so unsupported syntax survives a
// parse and print round trip.
type BadObj struct {
	From, To token.Pos // position range of the bad object
	Text     string    // source text of the bad object, excluding '\n'
}

func (*BadObj) objNode() {}

// Pos implements Node
func (o *BadObj) Pos() token.Pos {
	return o.From
}

// End implements Node
func (o *BadObj) End() token.Pos {
	return o.To
}

// String returns the source text of the bad object
func (o *BadObj) String() string {
	return o.Text
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

// A comment represents a single comment starting with '#'.
//
// Text is the source between the pound and the end of the line, verbatim. It
// excludes the '#' and the terminating '\n' and includes any whitespace that
// follows the pound, so "#foo" and "# foo" are distinct nodes and the printer
// writes back exactly what was parsed.
type Comment struct {
	Pound token.Pos // position of '#' starting the comment
	Text  string    // comment text following '#', excluding '\n'
}

// Pos implements Node
func (c *Comment) Pos() token.Pos {
	return c.Pound
}

// End implements Node.
//
// The printer renders a comment as '#' followed by Text, so End accounts for
// the pound. Text never includes the '#', but it does include the whitespace
// that follows it, so no other character has to be accounted for.
func (c *Comment) End() token.Pos {
	return c.Pound + 1 + token.Pos(len(c.Text)) // '#' + Text
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
//
// Quote holds ' or ", or [token.ILLEGAL] when the expression carries no quote.
// An expression without a quote is not valid make, the parser records one when
// it recovers from a conditional whose argument was not quoted, and both End
// and String treat the quote as an absent token of zero width.
type QuotedExpr struct {
	Quote token.Token // ', ", or ILLEGAL when unquoted
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
	return l.Close + tokenLen(l.Quote)
}

// String returns the quoted expression.
//
// Quotes are written when Quote holds one. [token.ILLEGAL] marks an absent
// quote and writes nothing, so an unquoted expression renders as its value. A
// nil Value contributes nothing, leaving a pair of quotes.
func (l *QuotedExpr) String() string {
	var buf strings.Builder
	if l.Quote != token.ILLEGAL {
		buf.WriteString(l.Quote.String())
	}
	if l.Value != nil {
		fmt.Fprint(&buf, l.Value)
	}
	if l.Quote != token.ILLEGAL {
		buf.WriteString(l.Quote.String())
	}

	return buf.String()
}

// VarRef represents a variable reference.
type VarRef struct {
	Dollar token.Pos   // position of '$'
	Open   token.Token // opening token, '(', '{', or ILLEGAL when undelimited
	Name   string      // variable identifier
	Close  token.Token // closing token, ')', '}', or ILLEGAL when undelimited
}

func (*VarRef) exprNode() {}

// Pos implements Node
func (v *VarRef) Pos() token.Pos {
	return v.Dollar
}

// End implements Node
func (v *VarRef) End() token.Pos {
	// '$' + Open + Name + Close
	return v.Dollar + 1 + tokenLen(v.Open) + token.Pos(len(v.Name)) + tokenLen(v.Close)
}

// String implements fmt.Stringer.
//
// Delimiters are written when Open and Close hold them. [token.ILLEGAL] marks
// an absent delimiter, so a reference records whether it is delimited
// independently of the length of Name.
func (v *VarRef) String() string {
	var buf strings.Builder
	buf.WriteString(token.DOLLAR.String())
	if v.Open != token.ILLEGAL {
		buf.WriteString(v.Open.String())
	}
	buf.WriteString(v.Name)
	if v.Close != token.ILLEGAL {
		buf.WriteString(v.Close.String())
	}

	return buf.String()
}

// A FuncCall represents an expansion that calls a function, such as
// $(shell date) or $(patsubst %.c,%.o,$(SOURCES)). It is modeled on
// [go/ast.CallExpr]: Name is the callee and Args are the arguments, with the
// positions of the delimiters and of every separating ',' recorded so the
// printer can reproduce the source.
//
// Name is an [ast.Text] rather than a [token.Token] because make also accepts
// call syntax for a name it does not know, as in $(info text). Use
// [token.IsBuiltinFunction] on Name.Value to tell the two apart.
//
// Commas holds one position per separator, so it has len(Args)-1 entries when
// Args is not empty. A call written without arguments, $(shell), has no Args
// and no Commas.
type FuncCall struct {
	Dollar   token.Pos   // position of '$'
	Open     token.Token // opening delimiter, '(' or '{'
	Name     *Text       // function name
	Args     []*FuncArg  // call arguments
	Commas   []token.Pos // positions of the ',' separating Args
	Close    token.Token // closing delimiter, ')' or '}'
	ClosePos token.Pos   // position of Close
}

func (*FuncCall) exprNode() {}

// Pos implements Node
func (c *FuncCall) Pos() token.Pos {
	return c.Dollar
}

// End implements Node
func (c *FuncCall) End() token.Pos {
	return c.ClosePos + tokenLen(c.Close)
}

// NamePos returns the position of the function name. A name always follows the
// opening delimiter immediately, so it is derived the same way [VarRef] derives
// the position of its own name.
func (c *FuncCall) NamePos() token.Pos {
	return c.Dollar + 1 + tokenLen(c.Open) // '$' + Open
}

// String returns the source text of the call.
//
// Arguments are whitespace significant, so the text is rebuilt from the
// positions of the nodes it contains rather than from a fixed layout.
func (c *FuncCall) String() string {
	b := &textBuilder{pos: c.Dollar}
	b.write(c.Dollar, token.DOLLAR.String())
	b.write(c.Dollar+1, c.Open.String())
	if c.Name != nil {
		b.write(c.Name.ValuePos, c.Name.Value)
	}
	for i, a := range c.Args {
		if i > 0 && i <= len(c.Commas) {
			b.write(c.Commas[i-1], token.COMMA.String())
		}
		a.writeTo(b)
	}
	b.write(c.ClosePos, c.Close.String())

	return b.String()
}

// A FuncArg represents a single argument of a [FuncCall].
//
// An argument is the run of source between two of the call's top-level commas,
// From through To. make does not strip the whitespace inside an argument, so
// the range covers the whitespace that surrounds Parts and an argument that
// holds no parts at all is still a distinct, empty argument.
type FuncArg struct {
	From, To token.Pos // position range of the argument
	Parts    []Expr    // expressions the argument is composed of
}

// Pos implements Node
func (a *FuncArg) Pos() token.Pos {
	return a.From
}

// End implements Node
func (a *FuncArg) End() token.Pos {
	return a.To
}

// String returns the source text of the argument, whitespace included.
func (a *FuncArg) String() string {
	b := &textBuilder{pos: a.From}
	a.writeTo(b)

	return b.String()
}

func (a *FuncArg) writeTo(b *textBuilder) {
	for _, p := range a.Parts {
		if s, ok := p.(fmt.Stringer); ok {
			b.write(p.Pos(), s.String())
		}
	}

	// The argument runs to To, so the whitespace that trails its last part is
	// part of it just as the whitespace between parts is.
	b.write(a.To, "")
}

// textBuilder rebuilds source text from nodes and their positions, padding the
// gaps between them with spaces.
type textBuilder struct {
	buf strings.Builder
	pos token.Pos
}

func (b *textBuilder) write(pos token.Pos, text string) {
	// A gap is negative only when positions are out of order, which a
	// hand-built node can do. Padding is skipped rather than panicking.
	for range max(int(pos-b.pos), 0) {
		b.buf.WriteByte(' ')
	}

	b.buf.WriteString(text)
	b.pos = pos + token.Pos(len(text))
}

func (b *textBuilder) String() string {
	return b.buf.String()
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
	return r.PrefixPos + tokenLen(r.Prefix) + token.Pos(len(r.Value)) // Prefix + Value
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
	switch {
	case d.Close.IsValid():
		return d.Close + 1 // pos + len(')')
	case d.Arg2 != nil:
		return d.Arg2.End()
	case d.Arg1 != nil:
		// An empty second argument, as in `ifeq "a" ""`, leaves the directive
		// ending after the first one.
		return d.Arg1.End()
	default:
		return d.TokPos + tokenLen(d.Tok)
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

// tokenLen is the width of tok as the printer writes it. [token.ILLEGAL] marks
// an absent token rather than a literal, so it has no width.
func tokenLen(tok token.Token) token.Pos {
	if tok == token.ILLEGAL {
		return 0
	}

	return token.Pos(len(tok.String()))
}
