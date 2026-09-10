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

// All nodes that may appear in the recipe list of a [Rule] implement the
// RecipeObj interface.
//
// A recipe list holds the lines make passes to the shell and the conditional
// directives that select between them. make reads a conditional at the point
// the line appears, so a conditional written under a target line wraps the
// recipes it contains rather than the rule that owns them.
type RecipeObj interface {
	Node
	recipeObjNode()
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

	// LineEnding is the sequence that terminates every line of the file, "\n"
	// or "\r\n". The empty string means "\n", so a file built without one
	// prints with LF line endings.
	//
	// The ending is recorded once for the whole file rather than once per
	// line because a blank line is not a node. The printer recreates blank
	// lines from the byte gap between the nodes that surround them, so a
	// per-line record would not cover them. A file that mixes endings is
	// therefore normalized to the one that terminates the majority of its
	// lines.
	LineEnding string

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
	for i := len(f.Contents) - 1; i >= 0; i-- {
		if f.Contents[i] != nil {
			return f.Contents[i].End()
		}
	}
	return f.FileEnd
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
// Recipes holds the lines of the rule body in the order they were written. An
// element is a [Recipe] or an [IfBlock], because a conditional directive
// written under a target line selects which recipe lines the rule runs and so
// belongs to the rule rather than beside it.
//
// Comment is the comment ending the target line. A comment written there is
// not an object of its own: it shares a line with the rule, and a
// [CommentGroup] beside the rule would be printed on a line of its own.
//
// [Rule Syntax]: https://www.gnu.org/software/make/manual/html_node/Rule-Syntax.html
type Rule struct {
	Targets      []Expr      // rule targets
	Colon        token.Pos   // position of ':' separating targets and prerequisites
	PreReqs      []Expr      // rule pre-requisites
	Pipe         token.Pos   // position of '|' separating normal and order-only prerequisites
	OrderPreReqs []Expr      // order-only pre-requisites
	Comment      *Comment    // comment ending the target line, if any
	Recipes      []RecipeObj // recipe lines and the conditionals selecting them
}

func (*Rule) objNode() {}

// Pos implements Node
func (r *Rule) Pos() token.Pos {
	return r.Targets[0].Pos()
}

// End implements Node
func (r *Rule) End() token.Pos {
	for i := len(r.Recipes) - 1; i >= 0; i-- {
		if r.Recipes[i] != nil {
			return r.Recipes[i].End()
		}
	}
	if r.Comment != nil {
		return r.Comment.End()
	}
	for i := len(r.OrderPreReqs) - 1; i >= 0; i-- {
		if r.OrderPreReqs[i] != nil {
			return r.OrderPreReqs[i].End()
		}
	}
	for i := len(r.PreReqs) - 1; i >= 0; i-- {
		if r.PreReqs[i] != nil {
			return r.PreReqs[i].End()
		}
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

// A JuxtaposedExpr represents expressions written next to each other with
// nothing between them, as in $$(notdir or prefix$(FOO). make expands whatever
// is written without a blank separating it as a single value, so the parts
// belong to one expression rather than to several.
//
// Parts render back to back, so the node begins where its first part begins
// and ends where its last part ends. A juxtaposition of a single part carries
// no information the part does not, so the parser returns the part itself and
// only builds this node for two or more.
type JuxtaposedExpr struct {
	Parts []Expr // expressions written with no separator between them
}

func (*JuxtaposedExpr) exprNode() {}

// Pos implements Node.
//
// An empty juxtaposition occupies no source, so it has no position.
func (e *JuxtaposedExpr) Pos() token.Pos {
	if len(e.Parts) == 0 {
		return token.NoPos
	}

	return e.Parts[0].Pos()
}

// End implements Node
func (e *JuxtaposedExpr) End() token.Pos {
	if n := len(e.Parts); n > 0 {
		return e.Parts[n-1].End()
	}

	return token.NoPos
}

// String returns the source text of the juxtaposition.
//
// The parts are written with nothing between them, so the text is rebuilt from
// their positions the same way a [FuncCall] rebuilds its own.
func (e *JuxtaposedExpr) String() string {
	b := &textBuilder{pos: e.Pos()}
	for _, part := range e.Parts {
		if s, ok := part.(fmt.Stringer); ok {
			b.write(part.Pos(), s.String())
		}
	}

	return b.String()
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
	Prefix    token.Token // TAB, SEMI, or TEXT for a custom .RECIPEPREFIX
	PrefixPos token.Pos   // position of Prefix
	PrefixLit string      // source text of a TEXT prefix
}

// A recipe is an expression because its body is text make hands to the shell,
// and an object because a conditional directive holds objects and a
// conditional written inside a rule body holds recipes.
func (*Recipe) exprNode()      {}
func (*Recipe) objNode()       {}
func (*Recipe) recipeObjNode() {}

// PrefixText returns the source text of the prefix introducing the recipe.
//
// A recipe introduced by a custom .RECIPEPREFIX begins with an arbitrary
// character that has no token of its own, so it is recorded as [token.TEXT]
// carrying the character in PrefixLit. [token.ILLEGAL] marks an absent prefix
// rather than a literal, so it has no text.
func (r *Recipe) PrefixText() string {
	switch r.Prefix {
	case token.TEXT:
		return r.PrefixLit
	case token.ILLEGAL:
		return ""
	default:
		return r.Prefix.String()
	}
}

// Pos implements Node
func (r *Recipe) Pos() token.Pos {
	return r.PrefixPos
}

// End implements Node
func (r *Recipe) End() token.Pos {
	return r.PrefixPos + token.Pos(len(r.PrefixText())) + token.Pos(len(r.Value)) // Prefix + Value
}

// An Variable represents a make variable.
//
// Comment is the comment ending the assignment line. A comment written there
// is not an object of its own: it shares a line with the assignment, and a
// [CommentGroup] beside the variable would be printed on a line of its own.
type Variable struct {
	Name    Expr        // left-hand side of the assignment
	Op      token.Token // =, :=, ::=, :::=, !=, ?=
	OpPos   token.Pos   // position of Op
	Value   []Expr      // right-hand side of the assignment
	Comment *Comment    // comment ending the assignment line, if any
}

func (*Variable) objNode() {}

// Pos implements Node
func (s *Variable) Pos() token.Pos {
	return s.Name.Pos()
}

// End implements Node
func (s *Variable) End() token.Pos {
	if s.Comment != nil {
		return s.Comment.End()
	}
	for i := len(s.Value) - 1; i >= 0; i-- {
		if s.Value[i] != nil {
			return s.Value[i].End()
		}
	}
	return token.Pos(int(s.OpPos) + len(s.Op.String()))
}

// IfBlock represents a conditional directive and its parts.
//
// EndifComment is the comment ending the endif line. The block spans several
// lines and each one that can end in a comment records its own, so the comment
// on the opening directive is held by Directive and the one on an else by the
// [ElseBlock] it ends.
type IfBlock struct {
	Directive    IfDir        // conditional directive
	Text         []Obj        // text-if-true
	Else         []*ElseBlock // else directive blocks
	Endif        token.Pos    // position of ENDIF
	EndifComment *Comment     // comment ending the endif line, if any
}

func (*IfBlock) objNode()       {}
func (*IfBlock) dirNode()       {}
func (*IfBlock) recipeObjNode() {}

// Pos implements Node
func (b *IfBlock) Pos() token.Pos {
	return b.Directive.Pos()
}

// End implements Node
func (b *IfBlock) End() token.Pos {
	if b.EndifComment != nil {
		return b.EndifComment.End()
	}

	return b.Endif + 5 // pos + len("endif")
}

// ElseBlock represents an `else` clause in a conditional directive.
//
// Comment is the comment ending a bare else line. An else written with a
// condition ends its line with the condition, so the comment belongs to the
// [IfDir] in Condition rather than here, the same way a comment on the opening
// directive belongs to the directive.
type ElseBlock struct {
	Else      token.Pos // position of ELSE
	Condition IfDir     // condition, if it exists; nil otherwise
	Comment   *Comment  // comment ending a bare else line, if any
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
	}
	if b.Comment != nil {
		return b.Comment.End()
	}
	if b.Condition != nil {
		return b.Condition.End()
	}

	return b.Else + 4 // pos + len("else")
}

// IfeqDir represents a conditional directive block using `ifeq` or `ifneq`.
//
// Comment is the comment ending the directive line. A comment written there is
// not an object of its own: it shares a line with the directive, and a
// [CommentGroup] in the body of the block would be printed on a line of its
// own.
type IfeqDir struct {
	Tok     token.Token // IFEQ or IFNEQ
	TokPos  token.Pos   // position of Tok
	Open    token.Pos   // position of '(', if it exists
	Arg1    Expr        // first argument in the condition
	Comma   token.Pos   // position of ',', if it exists
	Arg2    Expr        // second argument in the condition
	Close   token.Pos   // position of ')', if it exists
	Comment *Comment    // comment ending the directive line, if any
}

func (*IfeqDir) ifDirNode() {}

// Pos implements Node
func (d *IfeqDir) Pos() token.Pos {
	return d.TokPos
}

// End implements node
func (d *IfeqDir) End() token.Pos {
	switch {
	case d.Comment != nil:
		return d.Comment.End()
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

// IfdefDir represents a conditional directive block using `ifdef` or `ifndef`.
//
// Comment is the comment ending the directive line. A comment written there is
// not an object of its own: it shares a line with the directive, and a
// [CommentGroup] in the body of the block would be printed on a line of its
// own.
type IfdefDir struct {
	Tok     token.Token // IFDEF or IFNDEF
	TokPos  token.Pos   // position of Tok
	VarName Expr        // variable-name
	Comment *Comment    // comment ending the directive line, if any
}

func (*IfdefDir) ifDirNode() {}

// Pos implements Node
func (d *IfdefDir) Pos() token.Pos {
	return d.TokPos
}

// End implements node
func (d *IfdefDir) End() token.Pos {
	if d.Comment != nil {
		return d.Comment.End()
	}
	if d.VarName != nil {
		return d.VarName.End()
	}

	// A directive with no variable name ends after the directive itself, the
	// way an ifeq with no arguments does. A name is missing from a directive
	// the parser recovered from, and that node is walked and printed like any
	// other.

	return d.TokPos + tokenLen(d.Tok)
}

// A DefineDir represents a multi-line variable definition, the block written
// between a define directive and the endef that terminates it. [Multi-Line]
//
// Body holds the lines between the two directives verbatim, one [Text] per
// line, excluding the line ending. make records the body of a define as the
// literal text it was written with rather than as make syntax, so a line that
// looks like a rule, a conditional, or another define is text of the value
// like any other line. A blank line inside the body is a line of the value as
// well, so it is recorded as an empty [Text] rather than left to the gap
// between two nodes the way a blank line between objects is.
//
// Op is the assignment operator written after the variable name, and
// [token.ILLEGAL] when the definition carries none. make accepts every
// operator an ordinary assignment accepts, so `define FOO +=` appends the body
// to FOO and `define FOO ?=` defines it only when FOO has no value.
//
// [Multi-Line]: https://www.gnu.org/software/make/manual/html_node/Multi_002dLine.html
type DefineDir struct {
	Define  token.Pos   // position of 'define'
	VarName Expr        // variable name, nil when the definition carries none
	Op      token.Token // =, :=, ::=, :::=, +=, ?=, !=, or ILLEGAL when absent
	OpPos   token.Pos   // position of Op
	Body    []*Text     // lines of the body, verbatim, excluding the line ending
	Endef   token.Pos   // position of 'endef'
}

func (*DefineDir) objNode() {}
func (*DefineDir) dirNode() {}

// Pos implements Node
func (d *DefineDir) Pos() token.Pos {
	return d.Define
}

// End implements Node
func (d *DefineDir) End() token.Pos {
	return d.Endef + tokenLen(token.ENDEF)
}

// An UndefineDir represents an undefine directive. [Undefine Directive]
//
// [Undefine Directive]: https://www.gnu.org/software/make/manual/html_node/Undefine-Directive.html
type UndefineDir struct {
	Undefine token.Pos // position of 'undefine'
	VarName  Expr      // variable name, nil when the directive carries none
}

func (*UndefineDir) objNode() {}
func (*UndefineDir) dirNode() {}

// Pos implements Node
func (d *UndefineDir) Pos() token.Pos {
	return d.Undefine
}

// End implements Node.
//
// A directive naming no variable ends after the directive itself, so the name
// is only read when there is one.
func (d *UndefineDir) End() token.Pos {
	if d.VarName != nil {
		return d.VarName.End()
	}

	return d.Undefine + tokenLen(token.UNDEFINE)
}

// tokenLen is the width of tok as the printer writes it. [token.ILLEGAL] marks
// an absent token rather than a literal, so it has no width.
func tokenLen(tok token.Token) token.Pos {
	if tok == token.ILLEGAL {
		return 0
	}

	return token.Pos(len(tok.String()))
}
