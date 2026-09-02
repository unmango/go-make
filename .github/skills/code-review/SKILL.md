---
name: code-review
description: Reviews changes to the go-make Makefile parser, printer, and AST. Use when reviewing a pull request or a diff in this repository, so that the review checks position preservation, round-trip fidelity, nil-safety of optional AST fields, and the README support matrix rather than generic Go style.
---

# Reviewing `go-make`

`go-make` parses a Makefile into an AST and prints it back out unchanged.
Every review question comes back to that: does the change still read what make reads, and does it still print back what was written?

Start from `/AGENTS.md` for build commands and architecture.
This skill covers what to check in a diff and what to leave alone.

## The two contracts

**Round-trip fidelity.**
`testdata/roundtrip/*.mk` is the contract for syntax that parses and prints unchanged.
`e2e_test.go` walks the directory and registers one spec per file, so a new fixture needs no registration.
A parser change without a matching printer change, or the reverse, usually breaks a fixture.

**Position preservation.**
The printer recreates spacing from stored `token.Pos` data rather than normalizing it.
A node that loses or invents a position prints in the wrong place, and the specs assert exact positions rather than semantic equivalence.

## What to check

### Byte round-tripping is not proof

A fixture can round-trip byte for byte while the tree is wrong.
Two real cases from this repo's history:

- `CFLAGS=-DFOO=1` printed back unchanged as a `BadObj` holding the whole line, rather than as the variable it is.
- `ifeq (a,b)# c` printed back unchanged with the comment sitting in the block body, because the gap the printer fills was zero.

When a change alters how something is _read_, ask for a comparison of the parse trees before and after across the existing fixtures, not just a green suite.

### Optional fields are nil, and nil is not an error

Several `ast.Expr` fields are legitimately nil, and a node the parser recovered from can be missing almost anything:

- `IfeqDir.Arg1` and `Arg2` are nil for `ifeq (,)` and `ifeq ($(CI),)`.
- `IfdefDir.VarName` is nil for a recovered directive.
- `QuotedExpr.Value` is nil for `""`.
- `Rule.Targets` can be empty, and a `CommentGroup` can hold no comments.

Flag any new `x.Pos()`, `x.End()`, `expr.SetPos`, or `expr.Copy` call reached without a nil guard.
`builder/expr` panics on a nil expression rather than skipping it, so the panic surfaces at `obj.Copy`, far from the change.

### `builder/obj` copies must stay deep

`clone` and `cloneDir` copy nodes by value, so any new pointer field is shared with the original until it is copied explicitly.
A shared pointer is invisible to the round-trip fixtures: it only shows up when `Copy` is followed by `SetPos`, which then moves the original's node.
A new pointer field on an AST node needs a line in `clone`/`cloneDir` and a spec asserting the copy is not identical to the source.

### New AST fields have four other homes

A field added to a node in `ast/ast.go` usually needs:

1. `Pos()`/`End()` updated, checking fields in the order they appear in the source.
2. A case in `ast/walk.go`, guarded for nil, so `Walk` and `Inspect` reach it.
3. Layout in `builder/obj.SetPos` and width in `builder/obj.End`.
4. A row in the `README.md` support matrix.

A diff touching only the parser and printer is usually incomplete.

### The printer's line endings come from gaps

`fillSpace` and `fillLines` pad from where the printer stands to a stored position, and `fillLines` turns a byte gap into blank lines.
Adding an explicit `writeLine` where a gap already supplies one produces a stray blank line, and it will not show up until a fixture exercises the empty-body case.
Check that a new printer branch ends its line the same way the branch beside it does.

### Agreeing with GNU Make beats accepting more input

This package models what GNU Make 4.4.1 reads.
An input make rejects should stay a parse error rather than become a tree make would never build, and an input make accepts should not become an error.
When a change alters a reading, the reasoning is worth checking against `make -f` on the actual input, and the outcome belongs in the README matrix.

### Specs

Ginkgo and Gomega, using `Describe`, `It`, `Entry`, and `DescribeTable`.
A new behavior wants a round-trip fixture _and_ a spec asserting the tree, because the fixture alone only pins the bytes.
Prefer the existing helpers: `parseOne`, `onlyRule`, `onlyVariable`, `exprShape` in the root package, and `printed` in `builder/obj`.

## What not to flag

- **Long doc comments explaining why.** The house style states the reasoning behind a decision on the node or function it applies to. Comments that read as verbose elsewhere are the convention here.
- **Hard-coded positions in specs.** They are the point. A spec asserting `token.Pos(17)` is pinning layout, not being brittle.
- **The `builder` column of the README matrix being blank.** Not every supported syntax has a builder constructor, and that is recorded rather than missing.
- **Fixtures that look redundant.** Spacing variants (`# c`, `#c`, `)# c`) exercise different printer paths.
