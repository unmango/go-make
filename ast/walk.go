package ast

import (
	"go/ast"
	"iter"
)

type Visitor = ast.Visitor

// isNil reports whether node is nil, covering both a nil interface value and a
// nil pointer stored in a list of a concrete node type.
func isNil[N Node](node N) bool {
	var zero N
	return any(node) == any(zero)
}

func walkList[N Node](v Visitor, list []N) {
	for _, node := range list {
		if !isNil(node) {
			Walk(v, node)
		}
	}
}

// Walk traverses an AST in depth-first order: it starts by calling v.Visit(node).
// If the visitor w returned by v.Visit(node) is not nil, Walk is invoked
// recursively with visitor w for each of the non-nil children of node. Nil
// children are skipped, so a Visitor is never handed a nil child node.
func Walk(v Visitor, node Node) {
	if v = v.Visit(node); v == nil {
		return
	}

	switch n := node.(type) {
	case *BadObj:
		// nothing to do
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
		if n.Value != nil {
			Walk(v, n.Value)
		}
	case *Variable:
		if n.Name != nil {
			Walk(v, n.Name)
		}
		walkList(v, n.Value)
	case *IfeqDir:
		if n.Arg1 != nil {
			Walk(v, n.Arg1)
		}
		if n.Arg2 != nil {
			Walk(v, n.Arg2)
		}
	case *IfdefDir:
		if n.VarName != nil {
			Walk(v, n.VarName)
		}
	case *ElseBlock:
		if n.Condition != nil {
			Walk(v, n.Condition)
		}
		walkList(v, n.Text)
	case *IfBlock:
		if n.Directive != nil {
			Walk(v, n.Directive)
		}
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
