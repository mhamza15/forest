# forest

Forest is a CLI tool for managing git worktrees and tmux sessions.

It simplifies the workflow of creating and managing multiple working copies of a repository, each in its own directory with a dedicated tmux session.

Check the installed version with `forest --version`.

## Install

```
go install github.com/mhamza15/forest@latest
```

## Quick start

Register a project:

```
forest project add ~/dev/myapp
```

Create a worktree and open it in a tmux session:

```
forest tree add myapp feature/login
```

Browse all projects and worktrees interactively:

```
forest tree
```

## Commands

```
Forest is a CLI tool for managing git worktrees and tmux sessions.

It simplifies the workflow of creating and managing multiple working
copies of a repository, each in its own directory with a dedicated
tmux session.

Usage:
  forest [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  config      Open configuration in your editor
  help        Help about any command
  project     Manage projects
  session     Manage tmux sessions
  tree        Manage and browse worktrees

Flags:
  -h, --help      help for forest
      --verbose   enable debug logging
  -v, --version   version for forest

Use "forest [command] --help" for more information about a command.
```

### Projects

```
Register, list, and manage forest projects.

Usage:
  forest project [command]

Available Commands:
  add         Register a new project
  list        List registered projects
  remove      Unregister a project
```

### Worktrees

```
Create, remove, and browse worktrees. Run without a subcommand
to open the interactive tree browser.

Usage:
  forest tree [flags]
  forest tree [command]

Available Commands:
  add         Create a new worktree and tmux session
  list        List worktrees for one or all projects
  remove      Remove a worktree and its tmux session
```

The interactive tree browser (`forest tree`) supports vim keybindings (`j`/`k`, `ctrl+n`/`ctrl+p`, arrows), `enter`/`l` to open, `tab` to expand/collapse, `d` to delete, `n` to create a new tree, `?` to toggle help, and `q` to quit.

### Sessions

```
Manage tmux sessions without affecting worktrees.

Usage:
  forest session [command]

Available Commands:
  kill        Kill a tmux session without removing its worktree
  list        List active tmux sessions
```

### Configuration

```
Opens the global config in $EDITOR. If a project name is given,
opens that project's config instead.

Usage:
  forest config [project] [flags]
```

## Configuration

Global config lives at `$XDG_CONFIG_HOME/forest/config.yaml` (defaults to `~/.config/forest/config.yaml`):

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/mhamza15/forest/main/internal/config/schema/config.schema.json

# Default directory for storing worktrees.
# Organized as: <worktree_dir>/<project>/<branch>
worktree_dir: ~/.local/share/forest/worktrees

# Default branch to base new worktrees on.
branch: main

# Default tmux window layout for new sessions.
# Each entry creates a window.
layout:
  - command: opencode
  - name: editor
    command: nvim
  - name: shell
    command: ""
```

Per-project configs live at `~/.config/forest/projects/<name>.yaml` and can override any global setting:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/mhamza15/forest/main/internal/config/schema/project.schema.json

repo: /path/to/repo
worktree_dir: ""   # empty = use global default
branch: ""         # empty = use global default

# Files to copy from the repo root into each new worktree.
copy:
  - .env
  - config/local.yml

# Project-specific layout (overrides global layout entirely).
layout:
  - command: opencode
  - name: editor
    command: nvim
  - name: shell
    command: ""
```

## Shell completions

```
forest completion fish > ~/.config/fish/completions/forest.fish
forest completion bash > /etc/bash_completion.d/forest
forest completion zsh > "${fpath[1]}/_forest"
```

Project names and branch names complete dynamically.
