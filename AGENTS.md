# AGENTS instructions for `go-make`

This is the canonical repository guide for coding agents in this repo.
`.github/copilot-instructions.md` should stay aligned with this file and point here for the full project context.

## Build, test, and formatting

- `make build` runs the standard compile check (`go build ./...`).
- `make test` runs Ginkgo only for changed packages and writes `.make/test`; use `make clean` if you need to force a rerun.
- `make test_all` runs the full Ginkgo suite across the repository.
- `go tool ginkgo run ./parser` runs a single package's suite.
- `go tool ginkgo --focus='single character variable reference and extra text' run ./parser` runs one focused spec.
- `make cover` runs the full suite with coverage and writes `cover.profile`.
- `make format` runs `go fmt` plus `dprint fmt`; Markdown formatting currently applies to `README.md`.
- If you change Nix files, CI also runs `nix flake check --all-systems` and `nix build`.

## High-level architecture

- `make.go` is the public facade. It re-exports the primary entry points from the lower-level packages: `parser.New`, `scanner.New`, `scanner.ScanTokens`, `printer.Fprint`, and `writer.New`.
- `scanner/` is the lexical layer. `ScanTokens` can be used directly as a `bufio.SplitFunc`, while `scanner.Scanner` wraps it to produce `token.Token` values plus `token.Pos` data backed by `go/token`-style filesets.
- `token/` and `ast/` define the shared model. `token/` mirrors `go/token` concepts and enumerates supported Make operators, directives, and built-in functions. `ast/` models files, rules, variables, comment groups, conditional blocks, text, quoted expressions, variable references, and recipes.
- `parser/` is a one-token-lookahead parser that consumes `scanner.Scanner` output and builds an `ast.File`. Its core job is turning token/position streams into a position-preserving AST for rules, variables, comments, and conditional directives.
- `printer/` is the round-trip output engine. It uses stored positions to recreate spacing and blank lines rather than normalizing formatting. `writer/` is mostly a thin convenience layer over `printer.Fprint`.
- `builder/` provides position-aware AST construction helpers (`builder/file`, `builder/rule`, `builder/text`, etc.) for programmatic AST creation and tests.
- `testdata/roundtrip/*.mk` is the contract for syntax that is expected to parse and print without modification.

## Key conventions

- Preserve `token.Pos` data when changing parser, printer, builder, or AST code. Tests assert exact positions, not just semantic equivalence.
- Treat round-trip fidelity as a primary behavior. Parser and printer changes usually need matching updates so fixtures in `testdata/roundtrip` still survive parse/print unchanged.
- `make test` is incremental, not exhaustive. For a guaranteed full run, use `make test_all`; for a fresh rerun of cached Make targets, use `make clean`.
- Tests are written with Ginkgo/Gomega. Follow the existing suite structure and use `Describe`, `It`, `Entry`, and `DescribeTable` patterns rather than switching styles within a package.
- Output behavior usually lives in `printer/`, not `writer/`; `writer/` mainly forwards to the printer package.
- Keep the root `make` package small and facade-like. Most implementation belongs in the underlying packages, with the root package re-exporting stable entry points.
