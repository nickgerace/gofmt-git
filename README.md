# gofmt-git

[![tag](https://img.shields.io/github/v/tag/nickgerace/gofmt-git?label=version&style=flat-square)](https://github.com/nickgerace/gofmt-git/releases/latest)
[![license](https://img.shields.io/github/license/nickgerace/gofmt-git?style=flat-square)](./LICENSE)

Want to avoid formatting unchanged Go files?
This CLI application formats Go files in a Git repository's worktree.
You can execute this within any directory inside of the repository.

```sh
% git status -s
 M main.go
 M pkg/cmd/run.go
 M pkg/cmd.go

% gofmt-git
main.go
pkg/cmd/run.go
pkg/cmd.go
```

## Limitations

This tool runs with the default settings of `go fmt` (e.g. `go fmt <filename>` or `go fmt ./...`).
It does not offer any formatting options at this time.

## Compatibility

`gofmt-git` should work on any primary platform that [Go](https://golang.org/) supports.
Please [file an issue](https://github.com/nickgerace/gofmt-git/issues) if your platform is unsupported, or is not working as expected.

## Code of Conduct

This repository follows and enforces the Go programming language's [Code of Conduct](https://golang.org/conduct).

## Maintainers

- [@nickgerace](https://nickgerace.dev)
