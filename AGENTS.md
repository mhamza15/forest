# Repository Guidelines

## Project Overview

Forest is a Go CLI tool for managing git worktrees paired with tmux sessions. It lets users register git repositories as "projects," create worktrees for branches (including from GitHub issue/PR links), and automatically open each worktree in a dedicated tmux session with a configured layout. An interactive TUI browser provides a tree view of all projects and worktrees.

Module: `github.com/mhamza15/forest`
Go version: 1.25.7

## Architecture & Data Flow

```
main.go → cmd.Execute()
             │
             ├── cmd/root.go          (Cobra root, logging, version, subcommand registration)
             │     ├── cmd/tree/       (worktree lifecycle: add, list, remove, switch, prune, browser)
             │     ├── cmd/project/    (project registration: add, list, remove)
             │     ├── cmd/session/    (tmux session management: list, kill)
             │     └── cmd/config/     (open config in $EDITOR)
             │
             └── internal/
                   ├── forest/         Orchestration layer combining config + git + tmux
                   ├── config/         YAML config loading/saving, XDG paths, config merging
                   ├── git/            Git worktree/branch/remote ops via exec.Command
                   ├── tmux/           Tmux session lifecycle via exec.Command
                   ├── tui/            Bubbletea interactive tree browser
                   ├── completion/     Cobra shell completion functions
                   └── github/         GitHub URL parsing + PR metadata via gh CLI
```

### Data flow for worktree creation

1. User invokes `forest tree add <branch>` (or a GitHub link, or `--project` override)
2. `config.Resolve(project)` loads global + project YAML configs, merges into `ResolvedConfig`
3. `forest.AddTree(rc, branch)` checks for existing worktree via `git.FindByBranch`, creates via `git.Add`, copies files via `git.CopyFiles`, symlinks files via `git.SymlinkFiles`
4. `forest.OpenSession(rc, branch, path)` creates tmux session via `tmux.NewSession`, applies layout via `tmux.ApplyLayout`
5. `tmux.SwitchTo(sessionName)` attaches the user to the new session

### Configuration hierarchy

- Global config: `$XDG_CONFIG_HOME/forest/config.yaml`
- Per-project configs: `$XDG_CONFIG_HOME/forest/projects/{name}.yaml`
- `config.Resolve()` merges both; project-level fields override global defaults

## Key Directories

| Path | Purpose |
|---|---|
| `cmd/` | Cobra command definitions, one file per leaf command |
| `cmd/tree/` | Worktree commands (add, list, remove, switch, prune, browser) |
| `cmd/project/` | Project registration commands (add, list, remove) |
| `cmd/session/` | Tmux session commands (list, kill) |
| `cmd/config/` | Config editor command |
| `internal/forest/` | High-level orchestration: `AddTree`, `OpenSession`, `RemoveTree` |
| `internal/config/` | Config types, YAML I/O, XDG path resolution, config merging |
| `internal/git/` | Git worktree, branch, remote, and file-copy operations |
| `internal/tmux/` | Tmux session, window, layout, and session-name utilities |
| `internal/tui/` | Bubbletea browser model, key bindings, lipgloss styles |
| `internal/completion/` | Shell completion for projects and branches |
| `internal/github/` | GitHub link parsing, PR head fetching via `gh` CLI |

## Development Commands

```sh
make build     # Build binary with version from git tags: -ldflags "-X ...cmd.version=..."
make test      # Run tests via gotestsum ./...
make lint      # Run golangci-lint run
make fmt       # Run golangci-lint run --fix (gofumpt formatting)
```

**Single-test example:**
```sh
gotestsum -- -run TestSessionName ./internal/tmux/
```

## Changelog

Update `CHANGELOG.md` with every user-facing change. Follow [Keep a Changelog](https://keepachangelog.com/en/1.1.0/) format. Group entries under `Added`, `Changed`, `Deprecated`, `Removed`, `Fixed`, or `Security`. Write entries from the user's perspective, not the implementation's.

## Code Conventions & Common Patterns

### Import grouping

Three groups separated by blank lines: stdlib, external, internal.

```go
import (
    "fmt"
    "log/slog"

    "github.com/spf13/cobra"

    "github.com/mhamza15/forest/internal/config"
    "github.com/mhamza15/forest/internal/forest"
)
```

### Formatting

- **gofumpt** (stricter gofmt). Run via `make fmt`.
- Octal literals use `0o` prefix: `0o755`, `0o644`.

### Command definitions

Each command group has a `root.go` with a `Command()` function that returns a `*cobra.Command` and registers subcommands. Leaf commands are defined in separate files via factory functions (`addCmd()`, `listCmd()`, etc.) returning `*cobra.Command`.

```go
func addCmd() *cobra.Command {
    return &cobra.Command{
        Use:   "add ...",
        Short: "...",
        Args:  cobra.ExactArgs(1),
        RunE:  runAdd,
        ValidArgsFunction: completion.Branches,
    }
}
func runAdd(cmd *cobra.Command, args []string) error {
    // ...
    return nil
}
```

- All commands use `RunE` (returns `error`), never `Run`.
- Flags are defined on the `*cobra.Command` in the factory function, stored in package-level vars.
- `SilenceErrors: true` on root; `SilenceUsage` set in `PersistentPreRun` after command resolution.
- `--project` / `-p` persistent flag on root for project override; commands infer from cwd when omitted.

### Error handling

- Early return on error. No `else` after error check.
- Wrap with context: `fmt.Errorf("creating worktree parent dir: %w", err)`
- Sentinel errors for flow control: `git.ErrWorktreeDirty`, `tmux.ErrNotRunning`
- Errors bubble up to `cmd.Execute()` which prints to stderr and exits 1.

### Logging

- `log/slog` with the standard `TextHandler` writing to stderr.
- Debug-level only, gated behind `--verbose` flag.
- Pattern: `slog.Debug("message", slog.String("key", value), slog.Any("err", err))`

### External commands

Git and tmux operations are thin wrappers around `exec.Command`. No client libraries.

```go
cmd := exec.Command("git", "worktree", "add", ...)
cmd.Dir = repoPath
out, err := cmd.CombinedOutput()
```

### Output formatting

- `text/tabwriter` for aligned tabular output.
- `lipgloss` styles for colored terminal output (Catppuccin palette).
- `charmbracelet/huh` for interactive prompts (confirmations, file pickers, text input).
- User-facing messages via `fmt.Printf` / `fmt.Println` to stdout.

### TUI (Bubbletea)

- `internal/tui/browser.go` implements the Elm architecture: `Model`, `Init`, `Update`, `View`.
- State machine with modes: browse, delete confirmation, force-delete, new branch input, deleting (spinner).
- Post-exit actions returned via `Model.Action()` and executed after TUI teardown.
- Key bindings in `internal/tui/keys.go` (vim-style + arrows).
- Styles in `internal/tui/styles.go`.

### Naming

- Packages: lowercase single words (`forest`, `config`, `git`, `tmux`, `tui`).
- Files: lowercase, one concept per file (`worktree.go`, `copy.go`, `remote.go`).
- Functions: `PascalCase` exports, `camelCase` internal. Descriptive verbs (`AddTree`, `OpenSession`, `FindByBranch`).
- Test functions: `Test<Function>[_<Scenario>]` (e.g., `TestLoadGlobal_MissingFile`, `TestParseLink_Errors`).
- Command aliases: `configcmd`, `projectcmd`, `sessioncmd`, `treecmd` when importing `cmd/` subpackages.

## Important Files

| File | Role |
|---|---|
| `main.go` | Entry point; calls `cmd.Execute()` |
| `cmd/root.go` | Root Cobra command, logging init, version resolution, subcommand wiring |
| `internal/forest/forest.go` | `AddTree`, `OpenSession`, `RemoveTree` orchestration |
| `internal/config/config.go` | `GlobalConfig` type, `LoadGlobal`, `SaveGlobal` |
| `internal/config/project.go` | `ProjectConfig`, `ResolvedConfig`, `Resolve`, `FindProjectByRemote` |
| `internal/config/paths.go` | XDG path helpers: `ConfigDir`, `DataDir`, `GlobalConfigPath`, `ProjectsDir` |
| `internal/config/schema.go` | YAML schema modeline generation |
| `internal/git/worktree.go` | `Add`, `Remove`, `List`, `FindByBranch`, `IsPrunable`, porcelain parsing |
| `internal/git/copy.go` | `CopyFiles` from repo root to new worktrees |
| `internal/git/remote.go` | `Remotes`, `RemoteURL`, `NormalizeRemoteURL`, `Fetch` |
| `internal/tmux/tmux.go` | `NewSession`, `SwitchTo`, `KillSession`, `ApplyLayout`, `SessionName` |
| `internal/tui/browser.go` | Bubbletea `Model` for interactive tree browser |
| `internal/github/github.go` | `ParseLink`, `FetchPRHead`, GitHub URL/PR handling via `gh` CLI |
| `Makefile` | Build, test, lint, fmt targets |
| `.golangci.yml` | Linter config (modernize, testifylint, gofumpt) |

## Runtime & Tooling

- **Go 1.25.7** (minimum; set in `go.mod`)
- **golangci-lint** for linting and formatting (gofumpt)
- **gotestsum** as test runner (wraps `go test`)
- **git** and **tmux** required at runtime (invoked via `exec.Command`)
- **gh** CLI required for GitHub PR metadata fetching (optional feature)
- Version injected at build time: `-ldflags "-X github.com/mhamza15/forest/cmd.version=$(VERSION)"`
- CI: GitHub Actions running `go test ./...` and `golangci-lint run` on Go 1.25.x

## Testing

- **Framework**: Go standard `testing` package + `github.com/stretchr/testify` (`assert` and `require`).
- **Runner**: `gotestsum ./...`
- **Pattern**: Table-driven tests with `t.Run()` subtests.

```go
tests := []struct {
    name string
    input string
    want  string
}{
    {name: "case one", input: "a", want: "b"},
}
for _, tt := range tests {
    t.Run(tt.name, func(t *testing.T) {
        assert.Equal(t, tt.want, SomeFunc(tt.input))
    })
}
```

- **Assertions**: `require.*` for preconditions (fail-fast), `assert.*` for outcomes.
- **Setup**: `t.TempDir()` for filesystem isolation, `t.Setenv()` for environment vars. No global setup/teardown.
- **Helpers**: Marked with `t.Helper()`. Example: `initTestRepo()` creates a real git repo with an initial commit.
- **Mocking**: None. Tests use real git repos and filesystem operations.
- **Test data**: Inline in test files (YAML strings, porcelain output literals). No fixture files.
- **Naming**: `Test<Function>` or `Test<Function>_<Scenario>`. Subtest names are lowercase descriptive phrases.
- **No cmd/ tests**: Tests exist only for `internal/` packages.

### Key linters

| Linter | Purpose |
|---|---|
| `modernize` | Enforce modern Go idioms |
| `testifylint` | Correct testify usage patterns |
| `gofumpt` | Strict formatting (superset of gofmt) |
