# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Added

- Symlink files from repo root into new worktrees via `symlink` project config field.
- `--project` / `-p` flag to override project on any command; project is inferred from working directory when omitted.
- `--branch` / `-b` flag on `tree add` to override the base branch for a new worktree.
- `--no-session` flag on `tree add` to create a worktree without opening a tmux session.
- Attach to existing tmux session when invoked outside tmux.
- Support GitHub URLs in `project add` to register a project by repository URL.
- Support GitHub issue and pull request URLs in `tree add` to create worktrees directly from links.
- Fork branch fetching for pull requests from external contributors.
- `tree prune` command to remove worktrees whose branches have been merged or deleted.
- Squash-merge detection in `tree prune` by checking remote branch existence.
- `--dry-run` flag for `tree prune`.
- `tree switch` command to switch to an existing worktree's tmux session.
- `tree list` command to list worktrees for one or all projects.
- `tree remove` command with automatic detection of current worktree when run with no arguments.
- `--force` flag for `tree remove` with interactive prompt on dirty worktrees.
- `project list` command.
- `project remove` command.
- `session list` command.
- `session kill` command.
- `config` command to open global or project config in `$EDITOR`.
- Interactive TUI browser for navigating projects and worktrees.
- Loading spinner in TUI during worktree deletion.
- Tmux window layout configuration via `layout` field in global and project configs.
- Optional `name` field for layout windows.
- Copy files from repo root into new worktrees via `copy` project config field.
- JSON Schema files for global and project configs with editor autocomplete via modelines.
- `projects_dir` global config for `project add` file picker starting directory.
- Dynamic shell completions for project names and branch names.
- `--version` flag with `debug.ReadBuildInfo` fallback.
- `--verbose` flag for debug logging.
- Styled terminal output with lipgloss (Catppuccin palette) and tabwriter alignment.
- XDG Base Directory support for config and data paths.
- CI workflow with test and lint jobs.
- Dependabot for Go modules and GitHub Actions.

### Changed

- Replace `huh` prompts with plain stdin confirmation in `tree remove`.
- Rename `project new` and `tree new` to `project add` and `tree add`.
- Use GitHub raw URLs for schema modelines instead of local files.
- Reject unknown subcommands on `tree` instead of launching TUI.
- Show help for unknown subcommands but not for runtime errors.
- Make `worktree_dir` optional with XDG default and fallback.
- Commands no longer accept `<project>` as a positional argument; use `--project` flag or cwd inference instead.

### Removed

- Remove `tint` and `go-isatty` dependencies from debug logging handler.

### Fixed

- Skip fork fetch when branch already exists locally.
- Detect deleted working directory in `tree remove`.
- Block key input in TUI while delete is in progress.
- Fix layout selecting wrong window when tmux `base-index` is non-zero.
- Fix `ExpandPath` matching `~user` paths incorrectly.
- Kill tmux session after worktree removal, not before.
- Sanitize colons in tmux session names.
- Fix list command column alignment inconsistencies.
- Check out existing branches instead of failing on `tree add`.
- Flatten branch names containing slashes in worktree directory paths.
- Switch to last session before killing the current one.
- Look up worktree paths from git instead of constructing them.
