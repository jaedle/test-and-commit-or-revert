# test && commit || revert - cli

This is a cli to apply [TCR](https://medium.com/@kentbeck_7670/test-commit-revert-870bbd756864) for your daily work.

## Usage

### Installation

```
task intall
```

This will:

1. Build the binary
2. Copy the build binary into `$HOME/bin`.

Please include `$HOME/bin` into your path configuration.

### Configuration

Create the file `tcr.json` in the git-repository root.

Example:

```json
{
  "test": "go test ./..."
}
```

Attributes:

- `test`: test command to run. Whitespaces within arguments (i.e. `task 'argument with space'`) are **not supported**.

### Run tcr

```sh
tcr
```

| worktree | result of test execution         | effect                               | exit code |    
|----------|----------------------------------|--------------------------------------|-----------|
| clean    | (will not be executed)           | (none)                               | zero      |
| dirty    | tests passed                     | a new commit is created with changes | zero      |
| dirty    | tests failed                     | worktree is reset to previous commit | non-zero  |
| dirty    | test command can not be executed | (none)                               | non-zero  |
