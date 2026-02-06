# forest

Forest is a CLI tool for managing git worktrees and tmux sessions.

It simplifies the workflow of creating and managing multiple working copies of a repository, each in its own directory with a dedicated tmux session.

## Install

```
go install github.com/mhamza15/forest@latest
```

## Quick start

Register a project:

```
forest project new ~/dev/myapp
```

Create a worktree and open it in a tmux session:

```
forest tree new myapp feature/login
```

Browse all projects and worktrees interactively:

```
forest tree
```

## Commands

### Projects

```
forest project new <path> [--name <name>]   Register a git repo as a project
forest project new                           Interactive mode with file picker
forest project remove <project>              Unregister a project
```

### Worktrees

```
forest tree new <project> <branch>           Create worktree + tmux session
forest tree remove <project> <branch>        Remove worktree + tmux session
forest tree                                  Interactive tree browser
```

The interactive tree browser supports vim keybindings (`j`/`k`, `ctrl+n`/`ctrl+p`, arrows), `enter` to open, `d` to delete, `n` to create a new tree, and `?` to toggle the full help view.

### Configuration

```
forest config                                Open global config in $EDITOR
forest config <project>                      Open project config in $EDITOR
```

## Configuration

Global config lives at `~/.config/forest/config.yaml`:

```yaml
# Default directory for storing worktrees.
# Organized as: <worktree_dir>/<project>/<branch>
worktree_dir: ~/.local/share/forest/worktrees

# Default branch to base new worktrees on.
branch: main

# Default tmux window layout for new sessions.
# Each entry creates a window. The first runs in the initial window.
layout:
  - command: nvim
  - command: ""    # plain shell
```

Per-project configs live at `~/.config/forest/projects/<name>.yaml` and can override any global setting:

```yaml
repo: /path/to/repo
worktree_dir: ""   # empty = use global default
branch: ""         # empty = use global default

# Files to copy from the repo root into each new worktree.
copy:
  - .env
  - config/local.yml

# Project-specific layout (overrides global layout entirely).
layout:
  - command: nvim
  - command: go test ./...
  - command: ""
```

## Shell completions

```
forest completion fish > ~/.config/fish/completions/forest.fish
forest completion bash > /etc/bash_completion.d/forest
forest completion zsh > "${fpath[1]}/_forest"
```

Project names and branch names complete dynamically.

## Requirements

- Go 1.25+
- git
- tmux
