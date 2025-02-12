package build

import "go/token"

type File interface {
	Start(token.Pos)
	Rule(func(Expr), ...func(Rule))
}

type Rule interface {
	Target(func(Expr))
}

type Expr interface {
	Text(string)
	VarRef(string)
}
