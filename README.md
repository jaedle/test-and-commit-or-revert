# test && commit || revert - cli

This is a cli to apply [TCR](https://medium.com/@kentbeck_7670/test-commit-revert-870bbd756864) for your daily work.

## Usage

### Installation

To locally install the cli you may run:

```
task intall
```

This will build and copy the binary into `$HOME/bin`.
Please make sure to have this configured within your path.

### Configuration

Create a configuration file `tcr.json` in the root of your repository:

```json
{
  "test": "go test ./..."
}
```

The configuration needs to contain your test command you wish to run before you either commit your changes or revert
those

### Running

```sh
tcr
```

If you have a clean worktree the command will do nothing and exit with no error.

If you have a dirty worktree it will run the tests. It may:

- if the tests pass (zero exit code), it commit the changes with a work in progress commit (`[WIP] Refactoring`)
- if the tests fail (non-zero exit code), it will reset all changes to the repository including untracked files
- do nothing if there is an error on executing the test command
