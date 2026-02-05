package tui

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mhamza15/forest/internal/config"
	"github.com/mhamza15/forest/internal/git"
	"github.com/mhamza15/forest/internal/tmux"
)

// mode tracks the current interaction state.
type mode int

const (
	modeBrowse mode = iota
	modeConfirmDelete
	modeDeleting
	modeNewSelectProject
	modeNewInputBranch
)

// projectNode is a collapsible project with its worktrees.
type projectNode struct {
	name     string
	repo     string
	trees    []git.Worktree
	expanded bool
}

// deleteResultMsg carries the outcome of an async delete operation.
type deleteResultMsg struct {
	projectIdx int
	treeIdx    int
	project    string
	branch     string
	err        error
}

// Model is the bubbletea model for the inline tree browser.
type Model struct {
	projects []projectNode
	cursor   int
	mode     mode
	keys     keyMap
	help     help.Model
	spinner  spinner.Model

	// For the "new tree" flow: which project was selected.
	newProject string

	// Text input for branch name in new-tree mode.
	input textinput.Model

	// Status messages shown after actions.
	status string
	err    error

	// The chosen action to execute after the TUI exits. Nil means
	// the user quit without selecting anything.
	action func() error
}

// NewModel loads projects and their worktrees, returning a model
// ready to run.
func NewModel() (Model, error) {
	names, err := config.ListProjects()
	if err != nil {
		return Model{}, err
	}

	var projects []projectNode

	for _, name := range names {
		proj, loadErr := config.LoadProject(name)
		if loadErr != nil {
			continue
		}

		trees, listErr := git.List(proj.Repo)
		if listErr != nil {
			trees = nil
		}

		projects = append(projects, projectNode{
			name:     name,
			repo:     proj.Repo,
			trees:    trees,
			expanded: false,
		})
	}

	ti := textinput.New()
	ti.Placeholder = "branch name"
	ti.CharLimit = 128
	ti.Width = 40

	s := spinner.New(
		spinner.WithSpinner(spinner.Dot),
		spinner.WithStyle(lipgloss.NewStyle().Foreground(lipgloss.Color("11"))),
	)

	return Model{
		projects: projects,
		keys:     defaultKeyMap(),
		help:     help.New(),
		spinner:  s,
		input:    ti,
	}, nil
}

// Init satisfies tea.Model.
func (m Model) Init() tea.Cmd {
	return nil
}

// Update handles input and state transitions.
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.help.Width = msg.Width
		return m, nil

	case deleteResultMsg:
		return m.handleDeleteResult(msg)

	case spinner.TickMsg:
		if m.mode == modeDeleting {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Forward to text input when active.
	if m.mode == modeNewInputBranch {
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	// In branch input mode, handle text input first.
	if m.mode == modeNewInputBranch {
		return m.handleNewBranchInput(msg)
	}

	// In delete confirmation mode.
	if m.mode == modeConfirmDelete {
		return m.handleConfirmDelete(msg)
	}

	// In new-tree project selection mode.
	if m.mode == modeNewSelectProject {
		return m.handleNewSelectProject(msg)
	}

	switch {
	case key.Matches(msg, m.keys.Quit):
		return m, tea.Quit

	case key.Matches(msg, m.keys.Up):
		m.moveCursor(-1)

	case key.Matches(msg, m.keys.Down):
		m.moveCursor(1)

	case key.Matches(msg, m.keys.Toggle):
		m.toggleExpand()

	case key.Matches(msg, m.keys.Open):
		return m.openSelected()

	case key.Matches(msg, m.keys.Delete):
		return m.startDelete()

	case key.Matches(msg, m.keys.New):
		return m.startNew()

	case key.Matches(msg, m.keys.Help):
		m.help.ShowAll = !m.help.ShowAll
	}

	return m, nil
}

// View renders the inline tree browser.
func (m Model) View() string {
	if len(m.projects) == 0 {
		return styleDim.Render("No projects registered. Use 'forest project new' to add one.") + "\n"
	}

	var b strings.Builder

	row := 0

	for _, p := range m.projects {
		prefix := "  "
		if row == m.cursor {
			prefix = styleCursor.Render("> ")
		}

		arrow := ">"
		if p.expanded {
			arrow = "v"
		}

		line := fmt.Sprintf("%s%s %s", prefix, styleDim.Render(arrow), styleProject.Render(p.name))
		b.WriteString(line + "\n")
		row++

		if !p.expanded {
			continue
		}

		for _, t := range p.trees {
			treePrefix := "    "
			if row == m.cursor {
				treePrefix = styleCursor.Render("  > ")
			}

			branch := t.Branch
			if branch == "" {
				branch = "(detached)"
			}

			b.WriteString(treePrefix + styleTree.Render(branch) + "\n")
			row++
		}
	}

	// Mode-specific footer.
	switch m.mode {
	case modeConfirmDelete:
		b.WriteString("\n" + styleError.Render("Delete this tree? (y/N)") + "\n")

	case modeDeleting:
		b.WriteString("\n" + m.spinner.View() + " Deleting...\n")

	case modeNewSelectProject:
		b.WriteString("\n" + styleDim.Render("Select a project for the new tree, then press enter") + "\n")

	case modeNewInputBranch:
		b.WriteString("\n" + styleDim.Render("Branch name: ") + m.input.View() + "\n")
	}

	if m.status != "" {
		b.WriteString("\n" + styleDim.Render(m.status) + "\n")
	}

	if m.err != nil {
		b.WriteString("\n" + styleError.Render(m.err.Error()) + "\n")
	}

	if m.mode == modeBrowse {
		b.WriteString("\n" + m.help.View(m.keys) + "\n")
	}

	return b.String()
}

// visibleRows returns the total number of visible rows.
func (m Model) visibleRows() int {
	count := 0

	for _, p := range m.projects {
		count++
		if p.expanded {
			count += len(p.trees)
		}
	}

	return count
}

func (m *Model) moveCursor(delta int) {
	total := m.visibleRows()
	if total == 0 {
		return
	}

	m.cursor += delta

	if m.cursor < 0 {
		m.cursor = 0
	}

	if m.cursor >= total {
		m.cursor = total - 1
	}
}

// cursorTarget returns which project and optionally which tree the
// cursor is pointing at. treeIdx is -1 if the cursor is on a project
// row.
func (m Model) cursorTarget() (projectIdx int, treeIdx int) {
	row := 0

	for pi, p := range m.projects {
		if row == m.cursor {
			return pi, -1
		}
		row++

		if !p.expanded {
			continue
		}

		for ti := range p.trees {
			if row == m.cursor {
				return pi, ti
			}
			row++
		}
	}

	return 0, -1
}

func (m *Model) toggleExpand() {
	pi, _ := m.cursorTarget()

	if pi < len(m.projects) {
		m.projects[pi].expanded = !m.projects[pi].expanded
	}
}

// openSelected sets the action to switch to the selected tree's tmux
// session, then quits the TUI so the caller can execute it.
func (m Model) openSelected() (tea.Model, tea.Cmd) {
	pi, ti := m.cursorTarget()

	// If cursor is on a project, toggle expand instead of opening.
	if ti == -1 {
		m.toggleExpand()
		return m, nil
	}

	p := m.projects[pi]
	branch := p.trees[ti].Branch

	sessionName := tmux.SessionName(p.name, branch)

	m.action = func() error {
		if err := tmux.RequireRunning(); err != nil {
			return err
		}

		if !tmux.SessionExists(sessionName) {
			wtPath := p.trees[ti].Path
			if err := tmux.NewSession(sessionName, wtPath); err != nil {
				return err
			}
		}

		return tmux.SwitchTo(sessionName)
	}

	return m, tea.Quit
}

func (m Model) startDelete() (tea.Model, tea.Cmd) {
	_, ti := m.cursorTarget()

	// Only trees can be deleted, not projects.
	if ti == -1 {
		m.status = "move cursor to a tree to delete it"
		return m, nil
	}

	m.mode = modeConfirmDelete
	m.status = ""
	m.err = nil

	return m, nil
}

func (m Model) handleConfirmDelete(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Confirm):
		return m.executeDelete()

	default:
		m.mode = modeBrowse
		return m, nil
	}
}

// executeDelete starts the async delete operation and shows a spinner.
func (m Model) executeDelete() (tea.Model, tea.Cmd) {
	pi, ti := m.cursorTarget()
	p := m.projects[pi]
	branch := p.trees[ti].Branch

	m.mode = modeDeleting
	m.status = ""
	m.err = nil

	deleteCmd := func() tea.Msg {
		rc, err := config.Resolve(p.name)
		if err != nil {
			return deleteResultMsg{
				projectIdx: pi, treeIdx: ti,
				project: p.name, branch: branch, err: err,
			}
		}

		sessionName := tmux.SessionName(p.name, branch)
		wtPath := filepath.Join(rc.WorktreeDir, rc.Name, branch)

		_ = tmux.KillSession(sessionName)

		return deleteResultMsg{
			projectIdx: pi, treeIdx: ti,
			project: p.name, branch: branch,
			err: git.Remove(rc.Repo, wtPath),
		}
	}

	return m, tea.Batch(m.spinner.Tick, deleteCmd)
}

// handleDeleteResult processes the outcome of an async delete.
func (m Model) handleDeleteResult(msg deleteResultMsg) (tea.Model, tea.Cmd) {
	m.mode = modeBrowse

	if msg.err != nil {
		m.err = msg.err
		return m, nil
	}

	pi := msg.projectIdx
	ti := msg.treeIdx

	m.projects[pi].trees = append(
		m.projects[pi].trees[:ti],
		m.projects[pi].trees[ti+1:]...,
	)

	m.status = fmt.Sprintf("deleted %s/%s", msg.project, msg.branch)

	total := m.visibleRows()
	if m.cursor >= total && total > 0 {
		m.cursor = total - 1
	}

	return m, nil
}

func (m Model) startNew() (tea.Model, tea.Cmd) {
	m.mode = modeNewSelectProject
	m.status = ""
	m.err = nil

	return m, nil
}

func (m Model) handleNewSelectProject(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.keys.Quit), key.Matches(msg, m.keys.Cancel):
		m.mode = modeBrowse
		return m, nil

	case key.Matches(msg, m.keys.Up):
		m.moveCursor(-1)
		return m, nil

	case key.Matches(msg, m.keys.Down):
		m.moveCursor(1)
		return m, nil

	case key.Matches(msg, m.keys.Open):
		pi, _ := m.cursorTarget()

		if pi < len(m.projects) {
			m.newProject = m.projects[pi].name
			m.mode = modeNewInputBranch
			m.input.Reset()
			m.input.Focus()
			return m, textinput.Blink
		}
	}

	return m, nil
}

func (m Model) handleNewBranchInput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeBrowse
		m.input.Blur()
		return m, nil

	case tea.KeyEnter:
		return m.executeNewTree()
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)

	return m, cmd
}

func (m Model) executeNewTree() (tea.Model, tea.Cmd) {
	branch := strings.TrimSpace(m.input.Value())
	if branch == "" {
		m.err = fmt.Errorf("branch name cannot be empty")
		return m, nil
	}

	m.input.Blur()

	rc, err := config.Resolve(m.newProject)
	if err != nil {
		m.err = err
		m.mode = modeBrowse
		return m, nil
	}

	wtPath := filepath.Join(rc.WorktreeDir, rc.Name, branch)

	if err := git.Add(rc.Repo, wtPath, branch, rc.Branch); err != nil {
		m.err = err
		m.mode = modeBrowse
		return m, nil
	}

	// Add to our in-memory list.
	for i, p := range m.projects {
		if p.name == m.newProject {
			m.projects[i].trees = append(m.projects[i].trees, git.Worktree{
				Path:   wtPath,
				Branch: branch,
			})
			m.projects[i].expanded = true
			break
		}
	}

	m.status = fmt.Sprintf("created %s/%s", m.newProject, branch)
	m.mode = modeBrowse

	return m, nil
}

// Action returns the post-exit action, if any. The caller should
// execute this after the TUI program finishes.
func (m Model) Action() func() error {
	return m.action
}
