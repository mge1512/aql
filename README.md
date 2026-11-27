# aql

AQL is the API Query Language: one small SQL dialect as the interface of
an MCP server.

<!-- variant: us -->

## What it is

A server for the Model Context Protocol (MCP) connects a language model to
the data behind an API. A server that offers AQL does not give the model
one tool per API call. It gives the model two tools and one language:

| Tool | Statements | Effect |
|---|---|---|
| `query` | `SELECT`, `DESCRIBE` | Reads |
| `execute` | `INSERT`, `UPDATE`, `DELETE` | Changes one record |

Some statements, for a backend with tickets and projects:

```sql
DESCRIBE tickets
SELECT id, name, status FROM tickets WHERE status = 'open' ORDER BY name LIMIT 20
SELECT status, COUNT(*) FROM tickets GROUP BY status HAVING COUNT(*) > 5
SELECT t.name, p.name AS project FROM tickets t JOIN projects p ON t.project = p.id
SELECT name FROM tickets WHERE project IN (SELECT id FROM projects WHERE owner = 'kim')
SELECT id FROM tickets WHERE body MATCH 'budget review'
UPDATE tickets SET status = 'closed' WHERE id = 'T-17'
```

This repository has the specification of `libaql`, the library that reads
and runs AQL. A server for one backend uses `libaql` and adds two things.
The first is a catalog with the entities and fields of that backend. The
second is an adapter that talks to its API. Such a server has the name
`mcp-server-<backend>-aql`.

## Why we do it

A server that maps an API call by call offers dozens of tools. Every tool
description uses context, and a model has to choose among all of them. One
query language has these effects:

- Two tools fit into the context of a small model that runs on your own
  hardware. Your data can stay in your network.
- Models already know SQL. A model finds the entities and their fields
  with `DESCRIBE`.
- A statement that is wrong gets a message that says what to write
  instead, so the model can correct itself.
- The server filters, joins, groups and counts. The model gets a small
  result.
- Writes are narrow. One statement changes one record, and it finds that
  record by its id. Reading and changing are two tools, so a host can
  allow the first and ask before the second.

One library for all servers has a further effect. Every server reads a
statement in the same way and answers a mistake with the same message.
The code that does this is reviewed once.

## The language in short

```
SELECT fields FROM entity [JOIN entity ON a.field = b.field]
    [WHERE condition] [GROUP BY field [HAVING condition]]
    [ORDER BY field [DESC]] [LIMIT n]
DESCRIBE [entity]
INSERT INTO entity (field, ...) VALUES ('value', ...)
UPDATE entity SET field = 'value', ... WHERE id = 'id'
DELETE FROM entity WHERE id = 'id'
```

- AQL uses only keywords that SQL also has.
- A condition compares with `=`, `!=`, `<`, `<=`, `>`, `>=`, `IS NULL`,
  `IN`, `BETWEEN`, `LIKE`, `CONTAINS` and `MATCH`. `AND`, `OR` and `NOT`
  join conditions.
- The functions are `COUNT`, `SUM`, `AVG`, `MIN` and `MAX`.
- A statement joins two entities with one `JOIN`, along a reference that
  the catalog declares.
- A subquery stands after `IN`, one level deep.
- `MATCH` is a search that the backend answers, for example a full text
  search. It works on the fields that the catalog marks for it.
- A write statement changes one record.

The language has no `OFFSET`, no outer join, no arithmetic and no other
function.

## How it is made

This repository is a project of
[Post-Coding Development (PCD)](https://github.com/mge1512/pcd). The
specification is the source of truth, and a language model translates it
into code. When the code is wrong, we change the specification and
translate again.

The front end of `libaql` is declarative. The specification has a token
table of regular expressions and a grammar in the notation of the Lemon
parser generator. A lexer generator builds the tokenizer from the table,
and a parser generator builds the parser from the grammar. The build fails
when the parser generator reports a conflict. No scanner and no parser is
written by hand.

The parser generator is
[lemongrass](https://github.com/mge1512/lemongrass), the Lemon parser
generator of SQLite with templates for Go and Rust. The lexer generator is
[re2c](https://re2c.org). The code that the two tools write needs no
library.

The specification names no target language. The first bindings are Go and
Rust. In code the library has the name `aql`. At run time a binding needs
the standard library of its language and no other library.

## Repository layout

| Path | Content |
|---|---|
| `specification/libaql.spec.md` | The specification of `libaql`, the source of truth |
| `documentation/architecture.md` | The design decisions and the order of work |
| `documentation/CHANGELOG.md` | The history of the specification |

## Status

- The specification has the front end so far: the tokens, the grammar, the
  limits, and the messages for a statement that cannot be read. Its Versions
  are in `documentation/CHANGELOG.md`.
- Later Versions add the rules for names and types, the evaluation, the
  write statements, the adapter contract and the MCP shell.
- There is no code and no package yet.
- The first server is
  [mcp-server-aha-aql](https://github.com/mge1512/aha-aql) for Aha!. It
  has its own engine today, and it moves to `libaql` when the library
  exists.

## Where AQL comes from

The language started as the Aha Query Language of the open source project
[aha-studio](https://github.com/grokify/aha-studio) (MIT license), and
`mcp-server-aha-aql` implements a dialect of it. This repository takes the
language out of that one server. The letters stay, with a new meaning: API
Query Language. Other products use the same letters for other languages.

## License

MIT. See [`LICENSE`](LICENSE).
