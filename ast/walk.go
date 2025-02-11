package ast

import (
	"go/ast"
	"iter"
)

type Visitor = ast.Visitor

func walkList[N Node](v Visitor, list []N) {
	for _, node := range list {
		Walk(v, node)
	}
}

func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *File:
		walkList(v, n.Contents)
	case *CommentGroup:
		walkList(v, n.List)
	case *Rule:
		walkList(v, n.Targets)
		walkList(v, n.PreReqs)
		walkList(v, n.OrderPreReqs)
		walkList(v, n.Recipes)
	case *Recipe:
		Walk(v, &n.Text)
	case *QuotedExpr:
		Walk(v, n.Value)
	case *Variable:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		walkList(v, n.Value)
	case *IfeqDir:
		Walk(v, n.Arg1)
		Walk(v, n.Arg2)
	case *IfdefDir:
		Walk(v, n.VarName)
	case *ElseBlock:
		Walk(v, n.Condition)
		walkList(v, n.Text)
	case *IfBlock:
		Walk(v, n.Directive)
		walkList(v, n.Text)
		walkList(v, n.Else)
	}
}

type inspector func(Node) bool

// Visit implements ast.Visitor.
func (i inspector) Visit(node ast.Node) (w ast.Visitor) {
	if i(node) {
		return i
	} else {
		return nil
	}
}

func Inspect(node Node, f func(Node) bool) {
	Walk(inspector(f), node)
}

func Preorder(root Node) iter.Seq[Node] {
	return func(yield func(Node) bool) {
		ok := true
		Inspect(root, func(n Node) bool {
			if n != nil {
				ok = ok && yield(n)
			}
			return ok
		})
	}
}
