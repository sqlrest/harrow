# harrow

`harrow` formats PostgreSQL SQL. It reads SQL, lays it out in a canonical style, and writes it back — and it **never changes what a statement means and never drops a comment**.

A harrow is the implement dragged across a plowed field to break up the clods and level the surface into a fine, even tilth. `harrow` does the same to SQL: rough, uneven input in, a smooth, canonical layout out.

## Why it's safe

Most SQL formatters re-emit your query from a parse tree and silently corrupt anything they don't fully understand. `harrow` doesn't. Every rendering is checked against the original through [`gomatic/go-sql`](https://github.com/gomatic/go-sql): if the formatted SQL doesn't have the **same meaning** (identical PostgreSQL fingerprint) and the **same comments** as the input, `harrow` discards it and emits the original untouched. Formatting that can't be proven faithful never ships.

It parses with PostgreSQL's own parser ([`pg_query`](https://github.com/pganalyze/pg_query_go)), so it understands exactly the SQL that PostgreSQL does.

## Install

```sh
go install github.com/sqlrest/harrow/cmd/harrow@latest
```

`harrow` links PostgreSQL's parser through cgo, so a C toolchain is required to build it.

## Usage

```sh
# Format standard input to standard output (composes in a pipe)
cat schema.sql | harrow

# Print each formatted file to standard output
harrow migrations/*.sql

# Rewrite changed files in place
harrow --write migrations/*.sql

# List the files whose formatting would change (exit-code-friendly in CI)
harrow --list migrations/*.sql
```

| Flag | Description |
|---|---|
| `--write`, `-w` | Rewrite each changed file in place instead of writing to stdout. |
| `--list`, `-l` | Print the paths of files whose formatting would change. |
| `--log-level` | `debug`, `info`, `warn` (default), `error`. |
| `--log-format` | `text` (default) or `json`. |

## Example

```sql
-- input
SELECT   a,b  FROM t WHERE x::int=1 and y in (1,2);
```

```sql
-- harrow
select a, b from t where x::int = 1 and y in (1, 2)
```

A statement carrying comments is preserved verbatim until comment-aware reformatting lands, so a comment is never lost.

## Built on go-sql

All the formatting, parsing, comparison, and verification live in the reusable [`gomatic/go-sql`](https://github.com/gomatic/go-sql) library; `harrow` is a thin CLI over it.

## License

MIT. See [LICENSE](LICENSE).
