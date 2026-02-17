# 🌳 forest

Forest is a simple CLI tool for managing git worktrees and tmux sessions.

It was created to support my development workflow, which involves multiple AI agents working in the same repository on different tasks, isolated with Git worktrees. I would find myself often creating a new tmux session, navigating to the repo I'd like to work on, creating a Git worktree, navigating to the worktree, then setting up my typical layout of a coding agent, an editor, and other miscellaneous windows depending on the project. With the proliferation of worktrees/tmux sessions, it got a little challenging to manage.

Forest does essentially just that, but in an automated, configurable fashion. It optionally adds a friendly TUI on top of that to hopefully make things easier to manage.

Disclaimer: This tool was created entirely with AI, with minimal direct code review from me. I needed a tool that served a certain purpose, but I also don't have the time to treat it as more mature software project. It's functional, but as you can imagine, there's probably tons of slop all around.

Inspired by [workmux](https://github.com/raine/workmux) and [twig](https://github.com/andersonkrs/twig).

## Install

```
go install github.com/mhamza15/forest@latest
```

## Quick start

First, register a git repository as a project:

```
forest project add ~/dev/myapp
```

Next, create a worktree and tmux session for a branch:

```
forest tree add myapp feature/login
```

This creates a worktree at `<worktree_dir>/myapp/feature-login` (`worktree_dir` defaults to `~/.local/share/forest/worktrees`), starts a tmux session named `myapp-feature-login`, and switches to it.

To browse all projects and their worktrees interactively:

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
  prune       Remove worktrees for branches merged into the base branch
  remove      Remove a worktree and its tmux session
  switch      Switch to an existing worktree's tmux session
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

Global config lives at `$XDG_CONFIG_HOME/forest/config.yaml`, or `~/.config/forest/config.yaml`:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/mhamza15/forest/main/internal/config/schema/config.schema.json

# Default directory for storing worktrees. Organized as: <worktree_dir>/<project>/<branch>.
# Defaults to $XDG_DATA_HOME/forest/worktrees, or ~/.local/share/forest/worktrees.
worktree_dir: /path/to/worktrees

# Default branch to base new worktrees on.
branch: main

# Default tmux window layout for new sessions across all projects. Each entry creates a separate window.
layout:
  - command: opencode

  - name: editor
    command: nvim

  - name: shell
    command: ""
```

Per-project configs live at `$XDG_CONFIG_HOME/forest/projects/<project>.yaml`, or `~/.config/forest/projects/<project>.yaml`. and can override any global setting:

```yaml
# yaml-language-server: $schema=https://raw.githubusercontent.com/mhamza15/forest/main/internal/config/schema/project.schema.json

# The path to the repo.
repo: /path/to/repo

# The directory to store this project's worktrees. Leave empty or omit
# to use the global default.
worktree_dir: /path/to/worktrees

# The default branch to base new worktrees on. Leave empty or omit to
# use the global default.
branch: main

# Default directory that your projects live in. Used as starting
# point in ` + "`" + `project add` + "`" + ` directory picker.
projects_dir: ~/dev

# Files to copy from the repo root into each new worktree.
copy:
  - .env
  - config/local.yml

# Project-specific layout (overrides global layout).
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

## Fish abbreviations

Forest ships as a [Fisher](https://github.com/jorgebucaran/fisher)-compatible fish plugin with short abbreviations for every command:

| Abbreviation | Expansion |
|---|---|
| `f` | `forest` |
| `ft` | `forest tree` |
| `fta` | `forest tree add` |
| `ftl` | `forest tree list` |
| `ftr` | `forest tree remove` |
| `fts` | `forest tree switch` |
| `ftpr` | `forest tree prune` |
| `ftp` | `forest tree --project` |
| `ftap` | `forest tree add --project` |
| `ftlp` | `forest tree list --project` |
| `ftrp` | `forest tree remove --project` |
| `ftsp` | `forest tree switch --project` |
| `ftprp` | `forest tree prune --project` |
| `fp` | `forest project` |
| `fpa` | `forest project add` |
| `fpl` | `forest project list` |
| `fpr` | `forest project remove` |
| `fs` | `forest session` |
| `fsl` | `forest session list` |
| `fsk` | `forest session kill` |
| `fsp` | `forest session --project` |
| `fslp` | `forest session list --project` |
| `fskp` | `forest session kill --project` |
| `fc` | `forest config` |

Install with Fisher:

```
fisher install mhamza15/forest
```

Or source directly:

```
source /path/to/forest/conf.d/forest.fish
```
