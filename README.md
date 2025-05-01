## kk - kubectl extension

An extension to kubectl.

## Features

1. Create & use custom aliases for context names.

```bash
# Run this to switch context interactively.
kk context
```

## Installation

Requires go to be installed.

```bash
go install github.com/piotrszyma/kk@latest
```

## Development

This package is built on top of [cobra-cli](https://github.com/spf13/cobra-cli/blob/main/README.md).

### Usage

```bash
# Run cobra-cli to add command "foo".
go tool cobra-cli add foo
```
