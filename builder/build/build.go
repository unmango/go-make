package build

import (
	"github.com/unmango/go-make/token"
)

type File interface {
	Start(token.Pos)
	AddRule(func(Rule), ...func(Rule))
	InsertRule(int, func(Rule), ...func(Rule))
}

type Rule interface {
	Target(func(Expr), ...func(Expr))
}

type Expr interface {
	Text(text string)
	VarRef(name string)
}
