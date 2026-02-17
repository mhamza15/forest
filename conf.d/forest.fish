# Fish abbreviations for the forest CLI.
#
# Install via Fisher:
#   fisher install mhamza15/forest
#
# Or source manually:
#   source /path/to/forest/conf.d/forest.fish

# Root command.
abbr --add f forest

# tree: worktree lifecycle and browser.
abbr --add ft  "forest tree"
abbr --add fta "forest tree add"
abbr --add ftl "forest tree list"
abbr --add ftr "forest tree remove"
abbr --add fts "forest tree switch"
abbr --add ftp "forest tree prune"

# project: project registration.
abbr --add fp  "forest project"
abbr --add fpa "forest project add"
abbr --add fpl "forest project list"
abbr --add fpr "forest project remove"

# session: tmux session management.
abbr --add fs  "forest session"
abbr --add fsl "forest session list"
abbr --add fsk "forest session kill"

# config: open config in editor.
abbr --add fc "forest config"
