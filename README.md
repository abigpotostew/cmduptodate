# cmduptodate
Command line utility to check if a go command is up to date by checking all source files and imported source files modified date in comparison to the respective target generated binary.

In general, if no error is returned (exit status 0), the go binary is considered up-to-date.

The goal of this project is to prevent unnecessary work such as recompiling your command which can be extended to preventing unnecessary downstream work such as building a docker container that consumes the binary.

## Dependencies
`go` available on your PATH. `go list` is invoked during this command. 
environment variable `GOPATH` is set

## Installation
`go install github.com/abigpotostew/cmduptodate`

## Usage
`cmduptodate -c github.com/you/yourproject -g path/to/yourbinary [-p github.com/you/yourproject]`

If no error is returned (exit status 0), the go binary is considered up-to-date:

Flags:
* `-c` (required) your go fully qualified package for target command
* `-g` (required) path to the target compiled command binary (does not need to exist)
* `-p` (optional) for multi command projects, this is the go project's base path. This helps exclude go standard library source files and vendor source. Defaults to the `-c` value

## Examples
Single command project
`cmduptodate -c github.com/you/yourproject/cmd/yourbinary -g path/to/yourbinary`
Possible outputs:
* If the `path/to/yourbinary` does not exist the output is `path/to/yourbinary is out of date because it does not exist` (exit code 1)
* If a source file `code.go` used by `yourbinary` is modified after `path/to/yourbinary` the output is `path/to/yourbinary is out of date with /$(GOPATH)/github.com/you/yourproject/code.go`  (exit code 1)
* If no source file is updated after `path/to/yourbinary` modified time the output is `path/to/yourbinary is up to date` (exit code 0)


Multi-command project
`cmduptodate -c github.com/you/yourproject/cmd/yourbinary -g path/to/yourbinary -p github.com/you/yourproject`
Possible outputs:
* If the `path/to/yourbinary` does not exist the output is `path/to/yourbinary is out of date because it does not exist` (exit code 1)
* If a source file `code.go` used by `yourbinary` is modified after `path/to/yourbinary` the output is `path/to/yourbinary is out of date with /$(GOPATH)/github.com/you/yourproject/yourpackage/code.go`  (exit code 1)
* If no source file is updated after `path/to/yourbinary` modified time the output is `path/to/yourbinary is up to date` (exit code 0)


## Applications
This is originally intended to be used with `github.com/go-task/task`. For example a task file could be defined as follows:
```yaml
version: '2'

tasks:
  build:
    cmds:
      - go build
    status:
      - cmduptodate -c github.com/you/yourproject -g bin/yourbinary
```



## Limitations
### Go Modules
Go Modules are not supported at this time.

### Vendor imports
This does not properly check vendor imports at this time. If you run `dep ensure` you need to force recompile your binary. I would like to support vendor in the future.
