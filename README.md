# Go Make

Makefile parsing and utilities in Go

## Usage

### Reading

The `make.Parser` is the primary way to read Makefiles.

```go
f := os.Open("Makefile")
p := make.NewParser(f, nil)

m, err := p.ParseFile()

fmt.Println(m.Rules)
```

The more primitive `make.Scanner` and `make.ScanTokens` used by `make.Parser` can be used individually.

Using `make.ScanTokens` with a `bufio.Scanner`

```go
f := os.Open("Makefile")
s := bufio.NewScanner(f)
s.Split(make.ScanTokens)

for s.Scan() {
  s.Bytes() // The current token byte slice i.e. []byte(":=")
  s.Text() // The current token as a string i.e. ":="
}
```

Using `make.Scanner`

```go
f := os.Open("Makefile")
s := make.NewScanner(f, nil)

for pos, tok, lit := s.Scan(); tok != token.EOF; {
  fmt.Println(pos) // The position of tok
  fmt.Println(tok) // The current token.Token i.e. token.SIMPLE_ASSIGN
  fmt.Println(lit) // Literal tokens as a string i.e. "identifier"
}

if err := s.Err(); err != nil {
  fmt.Println(err)
}
```

### Writing

Use `make.Fprint` to write ast nodes.

```go
var file *ast.File

err := make.Fprint(os.Stdout, file)
```
