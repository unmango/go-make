package builder

import "github.com/unmango/go-make/token"

type (
	RuleFunc = func(Rule)
	ExprFunc = func(Expr)
	FileFunc = func(File)
)

type File interface {
	Start(token.Pos)
	Rule(ExprFunc, ...RuleFunc)
}

type Rule interface {
	Target(ExprFunc)
}

type Expr interface {
	Text(string)
	VarRef(string)
}
