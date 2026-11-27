<!-- generated from spec: libaql.spec.md sha256:2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f -->
# libaql (Go)

libaql is the front end of AQL (API Query Language), for MCP servers. AQL is
a small SQL dialect. A language model reads data with `SELECT` and
`DESCRIBE`, and it changes one record with `INSERT`, `UPDATE` or
`DELETE`. libaql reads the text of one statement. It returns the
statement, or an error whose message says what to write instead.

This directory is the Go binding: the module
`github.com/mge1512/aql/go/aql`, package `aql`, Version 0.6.0 of the
specification `specification/libaql.spec.md`.

## Public interface

```go
func ParseStatement(text string, allowedKinds []StatementKind, guide Guide) (Statement, error)
func ReadSet() []StatementKind   // select, describe: for the query tool
func WriteSet() []StatementKind  // insert, update, delete: for the execute tool
```

- `allowedKinds` is `ReadSet()` or `WriteSet()`.
- `guide.Entities` has the names of the entities of the server. It must
  not be empty. `guide.Example` is a read statement that the server
  accepts. The messages use both.
- The result is one of `ReadStatement`, `DescribeStatement`,
  `InsertStatement`, `UpdateStatement` and `DeleteStatement`, or an
  error of the type `*AQLError` with the kind `syntax`.
- The library looks up no name. An entity or a field stays as the statement
  writes it. A number keeps every digit, as text.

Example:

```go
guide := aql.Guide{
	Entities: []string{"tickets", "projects", "users"},
	Example:  "SELECT id, name, status FROM tickets WHERE project = 'P-1'",
}
st, err := aql.ParseStatement("SELECT id FROM tickets WHERE status = 'open'", aql.ReadSet(), guide)
if err != nil {
	return err.Error() // a message for the language model
}
read := st.(aql.ReadStatement)
```

The API reference comes from the documentation comments: `make doc`, or
`go doc -all github.com/mge1512/aql/go/aql`.

## Consumers

A server for one backend binds to libaql through the DEPENDENCIES section of
its own specification. It adds what is particular to the backend: the names
of the entities and fields, and the calls of the API. Every such server reads
a statement in the same way and answers a mistake with the same message.

## Version alignment

A consumer pins the Version of the specification (0.6.0) and its SHA-256
(`2fafcf9ca530438a6fb11f674789ae149aa9e4e48961b854a40a93e0c0cf6a6f`) in its DEPENDENCIES section, and records the hash in its
translation report. The build service builds the library from source and
holds all consumers on one Version. The release tag is `go/aql/v0.6.0`.

## Build

The build needs Go 1.24 or newer, lemongrass 0.1.1 and re2c 4.6. The
versions of the two tools are exact, because the build compares the
generated files `grammar.go` and `lexer.go` with what the tools write.

| Target | Purpose |
|---|---|
| `make build` | check the tools, the generated files and the parity with the specification; then `go build` |
| `make test` | `make build`, then `go test ./...` |
| `make generate` | write `grammar.go` and `lexer.go` again, after a change of `grammar.y`, `lexer.re` or `tokens.re` |
| `make dist` | the release source tarball |
| `make doc` | the API reference |
| `make clean` | remove build output |

Nobody edits `tokens.re`, `lexer.go` or `grammar.go` by hand.

## Installation

Signed packages come from the build service (build.opensuse.org), as a
dependency of the servers at build time. A user of a server needs no
package of the library. There is no installation with curl.

## License

MIT. See `LICENSE`.
