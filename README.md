# Go Make

Makefile parsing and utilities in Go

## Usage

At present the scanning utilities are the most tested.

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
s := make.NewScanner(f)

for s.Scan() {
  s.Token() // The current token.Token i.e. token.SIMPLE_ASSIGN
  s.Literal() // Literal tokens as a string i.e. "identifier"
}

if err := s.Err(); err != nil {
  fmt.Println(err)
}
```

## Future

- `make.Parser`
- `make.Parse(file)`
