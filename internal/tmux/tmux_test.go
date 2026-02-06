package tmux

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSessionName(t *testing.T) {
	tests := []struct {
		project string
		branch  string
		want    string
	}{
		{
			project: "myapp",
			branch:  "feature",
			want:    "myapp-feature",
		},
		{
			project: "myapp",
			branch:  "feature/login",
			want:    "myapp-feature-login",
		},
		{
			project: "my.app",
			branch:  "v1.0",
			want:    "my_app-v1_0",
		},
		{
			project: "myapp",
			branch:  "fix:login:bug",
			want:    "myapp-fix-login-bug",
		},
	}

	for _, tt := range tests {
		t.Run(tt.project+"/"+tt.branch, func(t *testing.T) {
			assert.Equal(t, tt.want, SessionName(tt.project, tt.branch))
		})
	}
}

func TestIsRunning(t *testing.T) {
	t.Setenv("TMUX", "")
	assert.False(t, IsRunning())

	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	assert.True(t, IsRunning())
}

func TestRequireRunning(t *testing.T) {
	t.Setenv("TMUX", "")
	require.ErrorIs(t, RequireRunning(), ErrNotRunning)

	t.Setenv("TMUX", "/tmp/tmux-1000/default,12345,0")
	assert.NoError(t, RequireRunning())
}

func TestSessionExists_NoSuchSession(t *testing.T) {
	// This session name is unlikely to exist.
	assert.False(t, SessionExists("forest-test-nonexistent-session-xyz"))
}
