package builder

import "github.com/unmango/go-make/token"

type (
	RuleBuilder = func(Rule)
	ExprBuilder = func(Expr)
	FileBuilder = func(File)
)

type File interface {
	Start(token.Pos)
	Rule(ExprBuilder, ...RuleBuilder)
}

type Rule interface {
	Target(ExprBuilder)
}

type Expr interface {
	Text(string)
	VarRef(string)
}
