# Go-pgsanity

A Golang port of [markdrago/pgsanity](https://github.com/markdrago/pgsanity)
to run sanity checks on SQL scripts with focus on PostgreSQL syntax.

Under the hood, pgsanity uses
[ecpg](https://www.postgresql.org/docs/current/app-ecpg.html).

> ecpg is the embedded SQL preprocessor for C programs.
> It converts C programs with embedded SQL statements to
> normal C code by replacing the SQL invocations with
> special function calls. The output files can then be
> processed with any C compiler tool chain.

## License

See [LICENSE](./LICENSE) for details.

## Installation

```
go get github.com/erstam/go-pgsanity/pgsanity
```

You can control where the executable gets installed
using the %GOBIN% env var.

**Note:** Make sure that you have PostgreSQL installed and that the
`ecpg` tool is in %PATH%.  Check with `where ecpg` to confirm.
You can install [PostgreSQL here](https://www.postgresql.org/download/).

## Usage

At the cmd prompt / linux shell:

```
> pgsanity <path_to_file_or_folder>
```

- If a file path is provided, it must end with `.sql`.
- If a folder path is provided, the folder is scanned for all `.sql` files,
and those are processed, others are skipped.
  
**Note:** tested on Windows only!

## Contribute

PRs are welcome! :)
