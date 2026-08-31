package parser

import (
	"fmt"
	"io"
	"math"
	"slices"
	"strings"

	"github.com/unmango/go-make/ast"
	"github.com/unmango/go-make/scanner"
	"github.com/unmango/go-make/token"
)

type Parser struct {
	s      *scanner.Scanner
	file   *token.File
	errors scanner.ErrorList

	pos token.Pos
	tok token.Token // one token look-ahead
	lit string      // token literal

	// recipePrefix is the source text of the character introducing a recipe
	// line, rebound by an assignment to .RECIPEPREFIX.
	recipePrefix string

	// inRule reports whether the line about to be parsed follows a target
	// line, where a prefixed line is a recipe of that rule. make reports a
	// prefixed line written anywhere else as a recipe commencing before a
	// target, so the prefix only introduces a recipe here.
	inRule bool
}

// defaultRecipePrefix introduces a recipe line when .RECIPEPREFIX is empty.
const defaultRecipePrefix = "\t"

// recipePrefixVar is the variable that rebinds the character introducing a
// recipe line.
const recipePrefixVar = ".RECIPEPREFIX"

func New(r io.Reader, file *token.File) *Parser {
	if file == nil {
		file = token.NewFileSet().AddFile("", 1, math.MaxInt-2)
	}

	p := &Parser{
		s:    scanner.New(r, file),
		file: file,

		recipePrefix: defaultRecipePrefix,
	}
	p.next()

	return p
}

// recipePrefixToken returns the token recording a recipe introduced by the
// .RECIPEPREFIX character prefix.
//
// A tab is the character make introduces a recipe with by default and has a
// token of its own. Every other character is recorded as text carrying the
// character, including a ';', which has a token of its own but means something
// else: [token.SEMI] marks the semicolon that introduces a recipe on a target
// line, and the two are written in different places and printed differently.
func recipePrefixToken(prefix string) token.Token {
	if prefix == token.TAB.String() {
		return token.TAB
	}

	return token.TEXT
}

// isRecipePrefix reports whether the current token introduces a recipe line.
//
// The scanner has no notion of .RECIPEPREFIX, so a custom prefix is not a
// token of its own. It arrives as the leading character of the token that
// begins the line, and the token text is what identifies it.
func (p *Parser) isRecipePrefix() bool {
	if p.tok == token.NEWLINE || p.tok == token.EOF {
		return false
	}
	if tok := recipePrefixToken(p.recipePrefix); tok != token.TEXT {
		return p.tok == tok
	}

	return strings.HasPrefix(p.recipeTokenText(), p.recipePrefix)
}

// setRecipePrefix rebinds the recipe prefix from an assignment to
// .RECIPEPREFIX. make introduces a recipe with the first character of the
// value of the variable, and with a tab when the value is empty.
//
// The value is only known while parsing when it is written literally. A value
// that is a variable reference or a function call is expanded by make, so the
// prefix in effect is left alone rather than guessed at.
func (p *Parser) setRecipePrefix(v *ast.Variable) {
	switch v.Op {
	case token.RECURSIVE_ASSIGN, token.SIMPLE_ASSIGN,
		token.POSIX_ASSIGN, token.IMMEDIATE_ASSIGN:
	case token.APPEND_ASSIGN:
		// Appending changes the first character of the value only when the
		// value is empty, which is the case exactly when the prefix is a tab.
		if p.recipePrefix != defaultRecipePrefix {
			return
		}
	default:
		// '?=' assigns nothing because .RECIPEPREFIX always has a value, and
		// the output of '!=' is not known until make runs the command.
		return
	}

	if len(v.Value) == 0 {
		p.recipePrefix = defaultRecipePrefix
		return
	}
	if t, ok := v.Value[0].(*ast.Text); ok && len(t.Value) > 0 {
		p.recipePrefix = t.Value[:1]
	}
}

// consumeRecipePrefix advances past a custom recipe prefix. The scanner does
// not know the prefix, so it reads it as the leading character of a token that
// holds the beginning of the recipe body as well. The remainder of that token
// is left in place for the body.
func (p *Parser) consumeRecipePrefix(prefix string) {
	if text := p.recipeTokenText(); len(text) > len(prefix) {
		p.tok, p.lit = token.TEXT, text[len(prefix):]
		p.pos += token.Pos(len(prefix))
	} else {
		p.next()
	}
}

func (p *Parser) error(pos token.Pos, msg string) {
	epos := p.file.Position(pos)
	p.errors.Add(epos, msg)
}

func (p *Parser) errorExpected(pos token.Pos, msg string) {
	msg = "expected " + msg
	if p.pos == pos {
		switch {
		case p.tok.IsLiteral():
			msg += ", found " + p.lit
		default:
			msg += ", found '" + p.tok.String() + "'"
		}
	}

	p.error(pos, msg)
}

func (p *Parser) expect(tok token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok {
		p.errorExpected(pos, "'"+tok.String()+"'")
	}

	p.next()
	return pos
}

func (p *Parser) expectOneOf(tok ...token.Token) token.Pos {
	pos := p.pos
	if p.tok != tok[0] {
		ts := make([]string, len(tok))
		for i, t := range tok {
			ts[i] = fmt.Sprint("'", t, "'")
		}

		p.errorExpected(pos, "one of "+strings.Join(ts, ", "))
	}

	p.next()
	return pos
}

func (p *Parser) next() {
	p.pos, p.tok, p.lit = p.s.Scan()
}

// skipNewlines consumes newlines without consuming a recipe prefix.
// Object parsing starts at the first non-newline token so a leading tab
// remains part of the object it introduces.
func (p *Parser) skipNewlines() {
	for p.tok == token.NEWLINE {
		p.next()
	}
}

// parseText parses the text a value was written with. Text that is not there
// is not invented: a token that is not text records an error and yields no
// node, so no placeholder reaches a caller or the printer.
func (p *Parser) parseText() *ast.Text {
	if p.tok != token.TEXT {
		p.expect(token.TEXT)
		return nil
	}

	pos, name := p.pos, p.lit
	p.next()

	return &ast.Text{
		ValuePos: pos,
		Value:    name,
	}
}

// assignOps are the operators that define a variable. make reads the first one
// on a line and takes the rest of the line as the value, so a later one is
// ordinary text.
var assignOps = []token.Token{
	token.RECURSIVE_ASSIGN,
	token.SIMPLE_ASSIGN,
	token.POSIX_ASSIGN,
	token.IMMEDIATE_ASSIGN,
	token.IFNDEF_ASSIGN,
	token.SHELL_ASSIGN,
	token.APPEND_ASSIGN,
}

// isAssignOp reports whether tok is an operator that defines a variable.
func isAssignOp(tok token.Token) bool {
	return slices.Contains(assignOps, tok)
}

// joinsName reports whether tok continues a name written against it inside an
// expansion. The characters that delimit make syntax elsewhere spell
// themselves in a name, and the scanner splits on each of them, so the pieces
// of $(a;b) and $(a!=b) arrive separately and are put back together.
func joinsName(tok token.Token) bool {
	switch tok {
	case token.TEXT, token.SEMI, token.PIPE:
		return true
	default:
		return isAssignOp(tok)
	}
}

// joinDelimited extends value with the delimiters written against it, and with
// the text written against those.
//
// Each of those characters is a token of its own wherever it appears, and make
// gives them a meaning of their own in particular places only, so text that
// holds one elsewhere is reassembled from the pieces the scanner split it
// into: $(a;b) refers to the variable named "a;b", `ifeq 'a|b' 'c'` compares
// against "a|b", and $(a=b) refers to the variable named "a=b".
//
// pos is where value was written, and the scanner drops the blanks between
// tokens, so a gap in the positions is what reports one. A piece written apart
// from value belongs to whatever the caller reads next.
func (p *Parser) joinDelimited(pos token.Pos, value string) string {
	end := pos + token.Pos(len(value))
	for p.pos == end && joinsName(p.tok) {
		text := p.recipeTokenText()
		value += text
		end = p.pos + token.Pos(len(text))
		p.next()
	}

	return value
}

// nameStop holds the tokens parseObj reads itself. A rule and a variable are
// both a run of expressions followed by the character that says which one the
// line is, so that character ends the run rather than joining the expression
// written in front of it.
var nameStop = append([]token.Token{token.COLON}, assignOps...)

// maxCallArgs is the number of arguments make passes to each built-in
// function. Once a call has that many arguments make stops splitting, so any
// further comma belongs to the final argument: $(subst a,b,c,d) substitutes
// into the text "c,d". A function that takes any number of arguments, and a
// name that is not a built-in, are absent from the map and split on every
// top-level comma.
var maxCallArgs = map[token.Token]int{
	token.SUBST:      3,
	token.PATSUBST:   3,
	token.STRIP:      1,
	token.FINDSTRING: 2,
	token.FILTER:     2,
	token.FILTER_OUT: 2,
	token.SORT:       1,
	token.WORD:       2,
	token.WORDS:      1,
	token.WORDLIST:   3,
	token.FIRSTWORD:  1,
	token.LASTWORD:   1,
	token.DIR:        1,
	token.NOTDIR:     1,
	token.SUFFIX:     1,
	token.BASENAME:   1,
	token.ADDSUFFIX:  2,
	token.ADDPREFIX:  2,
	token.JOIN:       2,
	token.WILDCARD:   1,
	token.REALPATH:   1,
	token.ABSPATH:    1,
	token.ERROR:      1,
	token.WARNING:    1,
	token.SHELL:      1,
	token.ORIGIN:     1,
	token.FLAVOR:     1,
	token.LET:        3,
	token.FOREACH:    3,
	token.IF:         3,
	token.INTCMP:     5,
	token.EVAL:       1,
	token.FILE:       2,
	token.VALUE:      1,
}

// closeDelim is the delimiter that terminates an expansion opened by open.
func closeDelim(open token.Token) token.Token {
	if open == token.LBRACE {
		return token.RBRACE
	}

	return token.RPAREN
}

// opensExpansion reports whether tok begins the expansion a '$' introduces:
// the '$' of an escaped '$$', a delimited '$(name)' or '${name}', or the name
// of an undelimited '$name'.
func opensExpansion(tok token.Token) bool {
	switch tok {
	case token.TEXT, token.DOLLAR, token.LPAREN, token.LBRACE:
		return true
	default:
		return false
	}
}

// parseRef parses a variable reference or a function call. A '$' that opens
// neither is text: an escaped '$$' returns an ast.Text holding both characters,
// and a '$' with no expansion after it returns one holding the character
// itself.
//
// A name that is not there is not invented. An expansion whose delimiter is
// followed by no name records an error and yields no node, so no placeholder
// reaches a caller or the printer.
func (p *Parser) parseRef() ast.Expr {
	if p.tok != token.DOLLAR {
		p.expect(token.DOLLAR)
		return nil
	}

	dollar := p.pos
	p.next()

	// A '$' opens an expansion only when an expansion follows it immediately.
	// Anywhere else the character stands for itself, so "a$" at the end of a
	// line and the "$" of "a$ b" reach the printer as the text they were
	// written with.
	if p.pos != dollar+1 || !opensExpansion(p.tok) {
		return &ast.Text{
			ValuePos: dollar,
			Value:    token.DOLLAR.String(),
		}
	}

	open, name := token.ILLEGAL, ""
	switch p.tok {
	case token.DOLLAR:
		// '$$' escapes a literal '$'. Both characters belong to the value so
		// the printer writes them back unchanged.
		p.next()
		return &ast.Text{
			ValuePos: dollar,
			Value:    "$$",
		}
	case token.LPAREN, token.LBRACE:
		open = p.tok
		p.next()
		if p.tok == token.TEXT || p.tok.IsBuiltinFunction() {
			builtin, namePos := p.tok.IsBuiltinFunction(), p.pos
			name = p.lit
			p.next()
			if p.startsCallArgs(namePos+token.Pos(len(name)), open, builtin) {
				return p.parseFuncCall(dollar, open, &ast.Text{
					ValuePos: namePos,
					Value:    name,
				})
			}

			name = p.joinDelimited(namePos, name)
		} else {
			// A delimiter with no name after it names no variable, so the
			// error is recorded and no node is built rather than one carrying
			// a name that was never written.
			p.errorExpected(p.pos, "'"+token.TEXT.String()+"'")
			p.skipRef(open)
			return nil
		}
	case token.TEXT:
		if len(p.lit) == 1 {
			name = p.lit
			p.next()
		} else {
			// TODO: This should occur in the scanner
			name = p.lit[:1]
			p.lit = p.lit[1:]
			p.pos++
		}
	default:
		// A token that opens no expansion names no variable, so the error is
		// recorded and no node is built.
		p.expectOneOf(token.TEXT, token.DOLLAR, token.LPAREN, token.LBRACE)
		return nil
	}

	close := token.ILLEGAL
	if open != token.ILLEGAL {
		switch p.tok {
		case token.RPAREN, token.RBRACE:
			close = p.tok
			p.next()
		default:
			p.expectOneOf(token.RPAREN, token.RBRACE)
		}
	}

	return &ast.VarRef{
		Dollar: dollar,
		Open:   open,
		Name:   name,
		Close:  close,
	}
}

// skipRef consumes what is left of an expansion whose name could not be
// parsed, up to and including the delimiter that closes it, so the caller
// resumes after the expansion rather than inside it. An expansion left
// unterminated ends at the line it was written on.
func (p *Parser) skipRef(open token.Token) {
	close := closeDelim(open)
	for p.tok != close && p.tok != token.NEWLINE && p.tok != token.EOF {
		p.next()
	}
	if p.tok == close {
		p.next()
	}
}

// startsCallArgs reports whether the token following a name inside an
// expansion begins an argument list. make reads an expansion as a call when
// the name is separated from what follows it by a blank, and reads a built-in
// name as a call when it is followed by a blank, a comma, or the closing
// delimiter. So $(dir) calls dir with no arguments, $(info text) calls a name
// make does not know, and $(dir:x) is a reference to a variable named "dir:x".
func (p *Parser) startsCallArgs(nameEnd token.Pos, open token.Token, builtin bool) bool {
	// The scanner drops the spaces between tokens, so a gap in the positions
	// is what reports one. A tab survives as a token of its own.
	if p.pos > nameEnd || p.tok == token.TAB {
		return true
	}

	return builtin && (p.tok == token.COMMA || p.tok == closeDelim(open))
}

// parseFuncCall parses the arguments and closing delimiter of a function call.
// The '$', the opening delimiter, and the name have already been consumed.
func (p *Parser) parseFuncCall(dollar token.Pos, open token.Token, name *ast.Text) *ast.FuncCall {
	call := &ast.FuncCall{
		Dollar: dollar,
		Open:   open,
		Name:   name,
		Close:  closeDelim(open),
	}
	if p.tok != call.Close && p.tok != token.NEWLINE && p.tok != token.EOF {
		p.parseCallArgs(call)
	}

	call.ClosePos = p.pos
	if p.tok == call.Close {
		p.next()
	} else {
		// A call is not a recovery boundary, the line is. Leaving the token
		// that ended the arguments unconsumed lets the caller resynchronize on
		// the newline the way it does for every other unfinished construct.
		p.errorExpected(p.pos, "'"+call.Close.String()+"'")
	}

	return call
}

// parseCallArgs parses the comma separated arguments of call, stopping at the
// closing delimiter that matches the one the call was opened with.
//
// A comma only separates arguments at the top level of the call. Commas nested
// in parentheses belong to the argument that contains them, so $(if
// $(findstring a,b),x,y) has three arguments rather than four. Nesting through
// another expansion needs no counting, because parsing it consumes its
// delimiters, but a parenthesis written as ordinary text does, so the depth is
// tracked as well.
func (p *Parser) parseCallArgs(call *ast.FuncCall) {
	maxArgs := maxCallArgs[token.Lookup(call.Name.Value)]
	arg, depth := &ast.FuncArg{From: p.pos}, 0

	for p.tok != token.NEWLINE && p.tok != token.EOF {
		switch p.tok {
		case call.Close:
			if depth == 0 {
				arg.To = p.pos
				call.Args = append(call.Args, arg)
				return
			}
			depth--
		case call.Open:
			depth++
		case token.COMMA:
			// make stops splitting once a function has all the arguments it
			// takes, so a comma past that count is part of the last argument.
			if depth == 0 && (maxArgs == 0 || len(call.Args) < maxArgs-1) {
				arg.To = p.pos
				call.Args = append(call.Args, arg)
				call.Commas = append(call.Commas, p.pos)
				arg = &ast.FuncArg{From: p.pos + 1}
				p.next()
				continue
			}
		}

		if part := p.parseCallArgPart(); part != nil {
			arg.Parts = append(arg.Parts, part)
		}
	}

	arg.To = p.pos
	call.Args = append(call.Args, arg)
}

// parseCallArgPart parses one piece of a function argument. Only '$' has a
// meaning to make inside a call, so every other token is kept as the text it
// was written with and reaches the printer unchanged.
func (p *Parser) parseCallArgPart() ast.Expr {
	switch p.tok {
	case token.DOLLAR:
		return p.parseRef()
	case token.TEXT:
		if text := p.parseText(); text != nil {
			return text
		}
		return nil
	default:
		text := &ast.Text{ValuePos: p.pos, Value: p.recipeTokenText()}
		p.next()
		return text
	}
}

// juxtaposable reports whether tok continues the expression that precedes it
// when the two are written with nothing between them. make expands an
// unseparated run of text and expansions as a single value, so a delimiter
// that only means something inside an expansion is ordinary text out here:
// $$(notdir is the escape '$$' joined to the text '(notdir'.
//
// A ';' and a '|' mean something on a target line only, a ':' and an
// assignment operator mean something once per line, and the caller that reads
// one names it in stop. Out here they join the run the way every other
// delimiter does: "a;b" is one value, "a|b" is one variable value, and
// "CFLAGS=-DFOO=1" holds one '=' that assigns and one that is text.
func juxtaposable(tok token.Token) bool {
	switch tok {
	case token.TEXT, token.DOLLAR,
		token.LPAREN, token.RPAREN,
		token.LBRACE, token.RBRACE,
		token.COMMA, token.SEMI, token.PIPE,
		token.COLON:
		return true
	default:
		// A built-in function name is a name only inside an expansion. Written
		// anywhere else it is the text it spells, which is what $$(notdir x)
		// makes of "notdir".
		return isAssignOp(tok) || tok.IsBuiltinFunction()
	}
}

// startsExpression reports whether tok can begin a whitespace-delimited
// expression. Text and the '$' of an expansion always can, and so can a ';'
// or a '|' written as a word of its own, because make gives those two
// characters a meaning on a target line only. A caller that reads them itself
// never reaches this, because it stops on them before parsing another
// expression, so "X = a ; b" is three words of a value rather than an error.
func startsExpression(tok token.Token) bool {
	switch tok {
	case token.TEXT, token.DOLLAR, token.SEMI, token.PIPE:
		return true
	default:
		return false
	}
}

// startsValue reports whether tok can begin a whitespace-delimited expression
// of a variable value. An assignment operator can, because make reads the
// first one on a line and takes the rest of the line as the value, so
// "X = a = b" holds the three words "a", "=" and "b".
//
// Nowhere else does an operator begin an expression. A target line reads one
// as the definition of a target-specific variable, which this package does not
// model, so "target: VAR = value" is an error rather than three prerequisites.
func startsValue(tok token.Token) bool {
	return startsExpression(tok) || isAssignOp(tok)
}

// parseExpression parses one whitespace-delimited expression.
//
// Expressions written with nothing between them are one expression, collected
// into an [ast.JuxtaposedExpr]. A run of a single part carries nothing the
// part does not, so it is returned as itself and the common case keeps the
// node it has always had.
//
// stop holds tokens the caller reads itself. An [ast.IfeqDir] ends an argument
// on ',' and on ')', and an [ast.Rule] ends a prerequisite on '|' and on ';',
// all of which are ordinary text anywhere else.
func (p *Parser) parseExpression(stop ...token.Token) ast.Expr {
	return p.parseExpr(startsExpression, stop)
}

// parseValue parses one whitespace-delimited expression of a variable value,
// where an assignment operator is ordinary text.
func (p *Parser) parseValue() ast.Expr {
	return p.parseExpr(startsValue, nil)
}

// parseExpr parses one whitespace-delimited expression beginning with a token
// starts accepts.
func (p *Parser) parseExpr(starts func(token.Token) bool, stop []token.Token) ast.Expr {
	if !starts(p.tok) {
		p.expectOneOf(token.TEXT, token.DOLLAR)
		return nil
	}

	first := p.parseExprPart()
	if first == nil {
		return nil
	}

	// The scanner drops the blanks between tokens, so a gap in the positions
	// is what reports one, the same way an expansion tells a call from a
	// reference.
	parts, end := []ast.Expr{first}, first.End()
	for p.pos == end && juxtaposable(p.tok) && !slices.Contains(stop, p.tok) {
		part := p.parseExprPart()
		if part == nil {
			break
		}

		parts = append(parts, part)
		end = part.End()
	}
	if len(parts) == 1 {
		return first
	}

	return &ast.JuxtaposedExpr{Parts: parts}
}

// parseExprPart parses one piece of an expression: an expansion introduced by
// '$', or the text everything else was written with. Only '$' means anything
// to make in a value, so a delimiter reached while joining pieces together
// reaches the printer as the character it was written with.
func (p *Parser) parseExprPart() ast.Expr {
	switch p.tok {
	case token.TEXT:
		if text := p.parseText(); text != nil {
			return text
		}
		return nil
	case token.DOLLAR:
		return p.parseRef()
	default:
		text := &ast.Text{ValuePos: p.pos, Value: p.recipeTokenText()}
		p.next()
		return text
	}
}

func (p *Parser) parseComment() *ast.Comment {
	// The scanner reports everything between the pound and the end of the line,
	// whitespace included. Storing it verbatim keeps "#foo" distinct from
	// "# foo" so the printer can write either one back unchanged.
	pos, lit := p.pos, p.lit
	p.next()

	return &ast.Comment{
		Pound: pos,
		Text:  lit,
	}
}

func (p *Parser) parseCommentGroup() *ast.CommentGroup {
	g := &ast.CommentGroup{}
	for p.tok == token.COMMENT {
		g.List = append(g.List, p.parseComment())
		if p.tok == token.NEWLINE {
			p.next() // potentially more comments in group
		}
	}

	return g
}

func (p *Parser) parseIfdefDir() *ast.IfdefDir {
	pos, tok := p.pos, p.tok
	p.next()
	arg := p.parseExpression()

	return &ast.IfdefDir{
		Tok:     tok,
		TokPos:  pos,
		VarName: arg,
	}
}

// parseQuotedExpr parses one quoted argument of an [ast.IfeqDir].
//
// A token that is not a quote yields no node. There is no quote to record and
// no position to report, so a node built there would claim to start at offset
// zero and print its value without the quotes it never had. The return type is
// [ast.Expr] so that the absent argument reaches the caller as a nil
// interface rather than a nil *[ast.QuotedExpr], which [ast.Walk] would
// dereference.
func (p *Parser) parseQuotedExpr() ast.Expr {
	var quote token.Token
	switch p.tok {
	case token.APOS, token.QUOTE:
		quote = p.tok
	default:
		p.expectOneOf(token.APOS, token.QUOTE)
		return nil
	}

	open := p.expect(quote)

	// `ifeq "" "b"` quotes an empty argument. The closing quote immediately
	// follows the opening one, and a nil Value records it. Text that fails to
	// parse is absent for the same reason.
	var value ast.Expr
	if p.tok != quote {
		if text := p.parseText(); text != nil {
			text.Value = p.joinDelimited(text.ValuePos, text.Value)
			value = text
		}
	}

	close := p.expect(quote)

	return &ast.QuotedExpr{
		Quote: quote,
		Open:  open,
		Value: value,
		Close: close,
	}
}

func (p *Parser) parseIfeqDir() *ast.IfeqDir {
	pos, tok := p.pos, p.tok
	p.next() // consume ifeq or ifneq

	var (
		lparen, rparen token.Pos
		arg1, arg2     ast.Expr
		comma          token.Pos
	)

	switch p.tok {
	case token.LPAREN:
		lparen = p.expect(token.LPAREN)
		// An argument may be empty, `ifeq ($(CI),)` compares against the empty
		// string. The delimiter that follows marks the absence, and a nil Arg
		// records it.
		if p.tok != token.COMMA {
			arg1 = p.parseExpression(token.COMMA, token.RPAREN)
		}
		comma = p.expect(token.COMMA)
		if p.tok != token.RPAREN {
			arg2 = p.parseExpression(token.COMMA, token.RPAREN)
		}
		rparen = p.expect(token.RPAREN)
	case token.APOS, token.QUOTE:
		arg1 = p.parseQuotedExpr()
		arg2 = p.parseQuotedExpr()
	default:
		p.expectOneOf(token.LPAREN, token.APOS, token.QUOTE)
	}

	return &ast.IfeqDir{
		Tok:    tok,
		TokPos: pos,
		Open:   lparen,
		Arg1:   arg1,
		Comma:  comma,
		Arg2:   arg2,
		Close:  rparen,
	}
}

func (p *Parser) parseIfDir() (d ast.IfDir) {
	switch p.tok {
	case token.IFDEF, token.IFNDEF:
		d = p.parseIfdefDir()
	case token.IFEQ, token.IFNEQ:
		d = p.parseIfeqDir()
	}

	return
}

func (p *Parser) parseElseBlock() *ast.ElseBlock {
	pos := p.expect(token.ELSE)
	condition := p.parseIfDir()

	p.skipNewlines()
	text := p.parseObjList()

	return &ast.ElseBlock{
		Else:      pos,
		Condition: condition,
		Text:      text,
	}
}

func (p *Parser) parseIfBlock() *ast.IfBlock {
	ifdir := p.parseIfDir()

	// Every branch of the block begins in the context the block itself began
	// in, so a rule opened before the block owns the prefixed lines of each
	// branch rather than only those of the first.
	inRule := p.inRule
	p.skipNewlines()
	text := p.parseObjList()

	var eblocks []*ast.ElseBlock
	for p.tok == token.ELSE {
		p.inRule = inRule
		b := p.parseElseBlock()
		eblocks = append(eblocks, b)
		p.skipNewlines()

		if b.Condition == nil {
			break
		}
	}

	endif := p.expect(token.ENDIF)

	return &ast.IfBlock{
		Directive: ifdir,
		Text:      text,
		Else:      eblocks,
		Endif:     endif,
	}
}

// isAssignOp reports whether tok is an assignment operator. A define directive
// accepts every operator an ordinary assignment accepts: `define FOO +=`
// appends its body to FOO and `define FOO ?=` defines FOO only when it has no
// value.
func isAssignOp(tok token.Token) bool {
	switch tok {
	case token.RECURSIVE_ASSIGN, token.SIMPLE_ASSIGN, token.POSIX_ASSIGN,
		token.IMMEDIATE_ASSIGN, token.IFNDEF_ASSIGN, token.SHELL_ASSIGN,
		token.APPEND_ASSIGN:
		return true
	default:
		return false
	}
}

// newlineWidth is the width of the newline the parser is looking at. The
// scanner reports the text of a CRLF ending in the literal and reports an LF
// ending with none, so an absent literal is the single byte of an LF.
func (p *Parser) newlineWidth() token.Pos {
	if p.lit != "" {
		return token.Pos(len(p.lit))
	}

	return 1
}

// appendExprs adds the expressions of src to dst, leaving out the ones that
// are not there. A recovering parse yields no node for an expression it could
// not read, and a nil carries no position for [Parser.parseBadObj] to write
// its text at.
func appendExprs(dst, src []ast.Expr) []ast.Expr {
	for _, e := range src {
		if e != nil {
			dst = append(dst, e)
		}
	}

	return dst
}

// parseDefineDir parses a multi-line variable definition, the block written
// between a define directive and the endef terminating it.
//
// The name is one expression, so a define written with blanks in its name is
// left to [Parser.parseBadObj] the way an assignment written with several
// names is. Anything else left on the line after the name and the optional
// assignment operator is unsupported syntax, and the line reaches the printer
// as the text it was written with rather than as a block that would print
// back missing it.
func (p *Parser) parseDefineDir() ast.Obj {
	pos := p.pos
	p.next()

	// The directive is kept as text so a line that turns out to be
	// unsupported reaches parseBadObj as the whole line it was written as.
	parsed := []ast.Expr{&ast.Text{ValuePos: pos, Value: token.DEFINE.String()}}

	var names []ast.Expr
	for p.tok == token.TEXT || p.tok == token.DOLLAR {
		names = append(names, p.parseExpression())
	}
	if len(names) > 1 {
		p.error(p.pos, "variable may have only one name")
		return p.parseBadObj(appendExprs(parsed, names))
	}

	d := &ast.DefineDir{Define: pos, Op: token.ILLEGAL}
	if len(names) == 1 {
		d.VarName = names[0]
	}
	if isAssignOp(p.tok) {
		d.Op, d.OpPos = p.tok, p.pos
		p.next()
	}
	if p.tok != token.NEWLINE && p.tok != token.EOF {
		parsed = appendExprs(parsed, names)
		if d.Op != token.ILLEGAL {
			parsed = append(parsed, &ast.Text{
				ValuePos: d.OpPos,
				Value:    d.Op.String(),
			})
		}

		return p.parseBadObj(parsed)
	}

	p.parseDefineBody(d)

	return d
}

// parseDefineBody reads the lines between a define directive and the endef
// terminating it, and records the position of that endef.
//
// make records the body of a define as the literal text it was written with,
// so every line is kept verbatim rather than parsed as make syntax. A line
// holds no node the printer could pad up to, so the blanks a line begins with
// are part of the text it carries, and a blank line is a line of the value
// like any other rather than a gap between two nodes.
//
// A line holding nothing but endef terminates the block. A define written in
// the body opens a block of its own that the endef matching it closes, which
// make counts the same way, so the value of the outer variable holds the inner
// block as text. A line whose endef carries text after it, or whose endef
// follows a tab, terminates nothing, matching make for the second and keeping
// the line intact for the first, which make warns about and this parser does
// not read.
func (p *Parser) parseDefineBody(d *ast.DefineDir) {
	depth := 1
	for p.tok == token.NEWLINE {
		lineStart := p.pos + p.newlineWidth()
		p.next()

		first, firstPos, n := p.tok, p.pos, 0
		b, nextPos := &strings.Builder{}, lineStart
		for p.tok != token.NEWLINE && p.tok != token.EOF {
			text := p.recipeTokenText()
			for range int(p.pos - nextPos) {
				b.WriteByte(' ')
			}

			b.WriteString(text)
			nextPos = p.pos + token.Pos(len(text))
			n++
			p.next()
		}

		// The newline that ends the file opens no line of its own, so a body
		// that runs to the end of the file ends with the last line written
		// rather than with an empty one.
		if n == 0 && p.tok == token.EOF {
			break
		}

		if first == token.ENDEF && n == 1 {
			if depth--; depth == 0 {
				d.Endef = firstPos
				return
			}
		} else if first == token.DEFINE {
			depth++
		}

		d.Body = append(d.Body, &ast.Text{
			ValuePos: lineStart,
			Value:    b.String(),
		})
	}

	// A block that runs to the end of the file was never terminated, which
	// make reports as well. The error is recorded the way a conditional
	// records a missing endif.
	d.Endef = p.expect(token.ENDEF)
}

// parseUndefineDir parses an undefine directive. The name is read the way a
// define reads its own, so a directive naming several variables, or carrying
// syntax the parser does not understand, is left to [Parser.parseBadObj].
func (p *Parser) parseUndefineDir() ast.Obj {
	pos := p.pos
	p.next()

	parsed := []ast.Expr{&ast.Text{ValuePos: pos, Value: token.UNDEFINE.String()}}

	var names []ast.Expr
	for p.tok == token.TEXT || p.tok == token.DOLLAR {
		names = append(names, p.parseExpression())
	}
	if len(names) > 1 {
		p.error(p.pos, "variable may have only one name")
		return p.parseBadObj(appendExprs(parsed, names))
	}
	if p.tok != token.NEWLINE && p.tok != token.EOF {
		return p.parseBadObj(appendExprs(parsed, names))
	}

	d := &ast.UndefineDir{Undefine: pos}
	if len(names) == 1 {
		d.VarName = names[0]
	}

	return d
}

func (p *Parser) parseObj() ast.Obj {
	// A conditional directive inside a rule body holds recipe lines, and the
	// body of the directive is an object list like any other, so a prefixed
	// line reached while the parser is inside a rule is read as a recipe.
	if p.inRule && p.isRecipePrefix() {
		return p.parseRecipe(recipePrefixToken(p.recipePrefix), p.recipePrefix)
	}

	switch p.tok {
	case token.COMMENT:
		return p.parseCommentGroup()
	case token.IFDEF, token.IFNDEF, token.IFEQ, token.IFNEQ:
		return p.parseIfBlock()
	case token.DEFINE:
		return p.parseDefineDir()
	case token.UNDEFINE:
		return p.parseUndefineDir()
	}

	// TODO: refactor to improve the error message
	// we expect one expression, then we expect one
	// of (Expr | COLON | *_ASSIGN)
	var l []ast.Expr
	for p.tok == token.TEXT || p.tok == token.DOLLAR {
		l = append(l, p.parseExpression(nameStop...))
	}

	switch p.tok {
	case token.COLON:
		return p.parseRule(l)
	case token.SIMPLE_ASSIGN, token.POSIX_ASSIGN, token.IMMEDIATE_ASSIGN,
		token.IFNDEF_ASSIGN, token.RECURSIVE_ASSIGN, token.SHELL_ASSIGN,
		token.APPEND_ASSIGN:
		switch len(l) {
		case 1:
			return p.parseVar(l[0])
		case 0:
			// An assignment operator is a token of its own wherever it is
			// written, so a line opening with one arrives here rather than as
			// the text of a name. make rejects a definition with no name.
			p.error(p.pos, "variable name is empty")
		default:
			p.error(p.pos, "variable may have only one name")
		}

		fallthrough
	default:
		return p.parseBadObj(l)
	}
}

// exprText reproduces the source text of an expression parsed from a line
// that turned out to be unsupported.
func exprText(e ast.Expr) string {
	if s, ok := e.(fmt.Stringer); ok {
		return s.String()
	}

	return ""
}

// parseBadObj consumes the remainder of the current line and returns a
// BadObj spanning it, including the expressions already parsed from the
// line. The line is the recovery boundary because make itself is
// line-oriented, so the next line is the first point the parser can be
// confident it is synchronized again.
func (p *Parser) parseBadObj(parsed []ast.Expr) *ast.BadObj {
	from := p.pos
	if len(parsed) > 0 {
		from = parsed[0].Pos()
	}

	b := &strings.Builder{}
	nextPos := from
	write := func(pos token.Pos, text string) {
		for range int(pos - nextPos) {
			b.WriteByte(' ')
		}

		b.WriteString(text)
		nextPos = pos + token.Pos(len(text))
	}

	for _, e := range parsed {
		write(e.Pos(), exprText(e))
	}
	for p.tok != token.NEWLINE && p.tok != token.EOF {
		write(p.pos, p.recipeTokenText())
		p.next()
	}

	return &ast.BadObj{
		From: from,
		To:   nextPos,
		Text: b.String(),
	}
}

func (p *Parser) parseObjList() (l []ast.Obj) {
	for p.tok != token.EOF && p.tok != token.ENDIF && p.tok != token.ELSE {
		o := p.parseObj()
		l = appendObj(l, o)
		p.trackRule(o)
		p.skipNewlines()
	}

	return
}

// appendObj adds o to l, moving into the preceding rule what belongs to its
// body.
//
// make reads a conditional at the point its line appears, so a conditional
// written under a target line and holding prefixed lines selects which recipes
// that rule runs. The conditional belongs in the recipe list rather than
// beside the rule, and so does a prefixed line that follows one. A conditional
// holding no recipe line wraps ordinary make syntax and is left where it was
// written.
//
// A conditional holding both recipe lines and a target line of its own is
// attached in full to the rule that precedes it. make would end the rule at
// the target line, but a block cannot be split between a recipe list and the
// object list around it, and the whole block still prints back unchanged.
//
// A recipe with no rule directly in front of it, such as one written after a
// conditional that stayed where it was, is left as an object of its own. It
// prints the line it was written as, and the rule it belongs to is not in
// reach.
func appendObj(l []ast.Obj, o ast.Obj) []ast.Obj {
	if len(l) == 0 {
		return append(l, o)
	}

	r, ok := l[len(l)-1].(*ast.Rule)
	if !ok {
		return append(l, o)
	}

	switch n := o.(type) {
	case *ast.Recipe:
		r.Recipes = append(r.Recipes, n)
		return l
	case *ast.IfBlock:
		if holdsRecipe(n) {
			r.Recipes = append(r.Recipes, n)
			return l
		}
	}

	return append(l, o)
}

// holdsRecipe reports whether any branch of b holds a recipe line, directly or
// through a conditional nested in one of them.
func holdsRecipe(b *ast.IfBlock) bool {
	if objsHoldRecipe(b.Text) {
		return true
	}
	for _, e := range b.Else {
		if objsHoldRecipe(e.Text) {
			return true
		}
	}

	return false
}

func objsHoldRecipe(l []ast.Obj) bool {
	for _, o := range l {
		switch n := o.(type) {
		case *ast.Recipe:
			return true
		case *ast.IfBlock:
			if holdsRecipe(n) {
				return true
			}
		}
	}

	return false
}

// trackRule records whether o leaves the parser inside a rule. A target line
// opens a rule body and a recipe keeps it open. A comment is ignored by make
// and leaves the context alone, and a conditional leaves the context its own
// contents left. Every other object is written without a prefix and closes the
// body.
func (p *Parser) trackRule(o ast.Obj) {
	switch o.(type) {
	case *ast.Rule, *ast.Recipe:
		p.inRule = true
	case *ast.CommentGroup, *ast.IfBlock:
		// leave the context as the object found it
	default:
		p.inRule = false
	}
}

func (p *Parser) parseVar(name ast.Expr) ast.Obj {
	op, opPos := p.tok, p.pos
	p.next()

	var rhs []ast.Expr
	for p.tok != token.COMMENT && p.tok != token.NEWLINE && p.tok != token.EOF {
		rhs = append(rhs, p.parseValue())
	}

	// A comment ends the value the way it ends a target line, and is recorded
	// on the variable for the same reason: it shares the assignment's line.
	var comment *ast.Comment
	if p.tok == token.COMMENT {
		comment = p.parseComment()
	}

	v := &ast.Variable{
		Name:    name,
		Op:      op,
		OpPos:   opPos,
		Value:   rhs,
		Comment: comment,
	}
	// .RECIPEPREFIX is an ordinary variable that the parser reads as well,
	// because it rebinds the character introducing a recipe for the rest of
	// the file.
	if t, ok := name.(*ast.Text); ok && t.Value == recipePrefixVar {
		p.setRecipePrefix(v)
	}

	return v
}

// recipeTokenText returns the source text of the current token. Tokens such
// as TEXT and UNSUPPORTED stringify to their name rather than to any source
// text, and the scanner reports the text it read in the literal, so the
// literal wins whenever the token carries one. Operators and directives
// carry no literal and stringify to the syntax they represent.
func (p *Parser) recipeTokenText() string {
	if p.tok == token.COMMENT {
		return "#" + p.lit
	}
	if p.lit != "" {
		return p.lit
	}

	return p.tok.String()
}

// parseRecipe reads a recipe introduced by prefixTok, written as the source
// text prefix. A rule may introduce its first recipe with a semicolon on the
// target line, so both are parameters rather than always [token.TEXT] and
// [Parser.recipePrefix]. The token is what tells that semicolon from a ';'
// that .RECIPEPREFIX bound to introduce a recipe line of its own, because the
// two are written with the same character.
func (p *Parser) parseRecipe(prefixTok token.Token, prefix string) *ast.Recipe {
	prefixPos := p.pos
	prefixWidth := token.Pos(len(prefix))
	if prefixTok == token.TEXT {
		p.consumeRecipePrefix(prefix)
	} else {
		p.expect(prefixTok)
	}
	b := &strings.Builder{}
	nextPos := prefixPos + prefixWidth
	for p.tok != token.NEWLINE && p.tok != token.EOF {
		if gap := int(p.pos - nextPos); gap > 0 {
			for range gap {
				b.WriteByte(' ')
			}
		}

		text := p.recipeTokenText()
		b.WriteString(text)
		nextPos = p.pos + token.Pos(len(text))
		p.next()
	}
	if p.tok == token.NEWLINE {
		p.next()
	}

	r := &ast.Recipe{
		Prefix:    prefixTok,
		PrefixPos: prefixPos,
		Text: ast.Text{
			Value:    b.String(),
			ValuePos: prefixPos + prefixWidth,
		},
	}
	if prefixTok == token.TEXT {
		r.PrefixLit = prefix
	}

	return r
}

func (p *Parser) parseRule(targets []ast.Expr) *ast.Rule {
	colon := p.expect(token.COLON)
	prereqs := []ast.Expr{}
	// A '|' and a ';' end the prerequisite list wherever they are written, so
	// they end a prerequisite written against them as well: "target: a|b" has
	// one normal prerequisite and one order-only prerequisite, the same as
	// "target: a | b".
	//
	// A comment ends the line, and so ends the prerequisite written against
	// it: make removes the comment before reading the line, so
	// "target: a# text" names the prerequisite "a".
	for p.tok != token.PIPE && p.tok != token.SEMI && p.tok != token.COMMENT &&
		p.tok != token.NEWLINE && p.tok != token.EOF {
		prereqs = append(prereqs, p.parseExpression(token.PIPE, token.SEMI, token.COLON))
	}

	pipe, oprereqs := token.NoPos, []ast.Expr{}
	if p.tok == token.PIPE {
		pipe = p.pos
		p.next()
		// Only the first '|' separates. make reads a later one as a character
		// of the prerequisite holding it, so "target: a|b|c" has the single
		// order-only prerequisite "b|c" and the '|' is absent from stop here.
		for p.tok != token.SEMI && p.tok != token.COMMENT &&
			p.tok != token.NEWLINE && p.tok != token.EOF {
			oprereqs = append(oprereqs, p.parseExpression(token.SEMI, token.COLON))
		}
	}

	// A comment runs to the end of the line, so it is the last thing the
	// target line can hold and a ';' written after one belongs to its text.
	// It is recorded on the rule rather than beside it because an
	// [ast.CommentGroup] is written on a line of its own.
	var comment *ast.Comment
	if p.tok == token.COMMENT {
		comment = p.parseComment()
	}

	// A semicolon ends the prerequisite list and introduces a recipe on the
	// target line itself, so it is read before the prefixed recipes on the
	// lines below. It consumes the rest of the line, the newline included,
	// which is otherwise skipped here.
	recipes := make([]ast.RecipeObj, 0)
	if p.tok == token.SEMI {
		recipes = append(recipes, p.parseRecipe(token.SEMI, token.SEMI.String()))
	} else if p.tok == token.NEWLINE {
		p.next()
	}

	// Blank lines do not end a recipe list. make ignores them and keeps
	// attaching prefixed lines to the rule, so the loop looks past any run of
	// newlines for another prefix. Newlines consumed here without a recipe
	// following them are the same newlines the caller would have skipped, and
	// the blank lines survive a round trip because the printer reconstructs
	// them from the position of whatever node comes next.
	for p.isRecipePrefix() || p.tok == token.NEWLINE {
		if p.tok == token.NEWLINE {
			p.skipNewlines()
		} else {
			recipes = append(recipes, p.parseRecipe(recipePrefixToken(p.recipePrefix), p.recipePrefix))
		}
	}

	return &ast.Rule{
		Targets:      targets,
		Colon:        colon,
		PreReqs:      prereqs,
		Pipe:         pipe,
		OrderPreReqs: oprereqs,
		Comment:      comment,
		Recipes:      recipes,
	}
}

func (p *Parser) parseFile() *ast.File {
	var content []ast.Obj
	for p.tok != token.EOF {
		if p.skipNewlines(); p.tok == token.EOF {
			break
		}

		o := p.parseObj()
		content = appendObj(content, o)
		p.trackRule(o)
	}

	return &ast.File{
		Contents:   content,
		FileStart:  token.Pos(p.file.Base()),
		FileEnd:    token.Pos(p.file.Base() + p.file.Size()),
		LineEnding: p.s.LineEnding(),
	}
}

func (p *Parser) ParseFile() (*ast.File, error) {
	f := p.parseFile()
	if p.errors.Len() > 0 {
		p.errors.Sort()
		return nil, p.errors.Err()
	} else {
		return f, nil
	}
}
