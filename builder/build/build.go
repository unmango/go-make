package build

import (
	"github.com/unmango/go-make/token"
)

type File interface {
	Start(token.Pos)
	Rule(target func(Expr), fs ...func(Rule))
}

type Rule interface {
	Target(func(Expr))
}

type Expr interface {
	Text(text string)
	VarRef(name string)
}
