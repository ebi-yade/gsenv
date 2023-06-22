# gsenv
A simple CLI tool that allows you to execute a command with secrets stored in Google Cloud Secret Manager mapped to environment variables.

## Installation

Just download an appropriate binary from releases, or run the command below:

```console
$ go install github.com/ebi-yade/gsenv/cmd/gsenv@latest
```

## Usage

The only top level command is currently supported.  Note that `--project` is required.

```console
$ gsenv --project my-project --filter 'name:KISS OR name:SLACK' -- my_command
```

## Acknowledgements

This project is inspired by [ssmwrap](https://github.com/handlename/ssmwrap).  Thanks! [@handlename](https://github.com/handlename)
