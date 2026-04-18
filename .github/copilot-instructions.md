# Copilot instructions for `go-make`

Start with `/AGENTS.md`. It is the canonical repository guide for build/test commands, architecture, and codebase conventions.

Keep these Copilot-specific reminders in mind while using it:

- Prefer `make build` for compile checks and `make test_all` for a full suite run.
- `make test` is incremental and writes `.make/test`; use `make clean` before relying on it for a fresh rerun.
- For focused testing, use Ginkgo directly, for example `go tool ginkgo run ./parser` or `go tool ginkgo --focus='single character variable reference and extra text' run ./parser`.
- Preserve `token.Pos` data when changing parser, printer, builder, or AST code. Tests check exact positions.
- Treat `testdata/roundtrip/*.mk` as the round-trip contract for parser/printer behavior.
