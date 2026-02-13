package github

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseLink_Issue(t *testing.T) {
	link, err := ParseLink("https://github.com/acme/widgets/issues/42")
	require.NoError(t, err)

	assert.Equal(t, KindIssue, link.Kind)
	assert.Equal(t, "acme", link.Owner)
	assert.Equal(t, "widgets", link.Repo)
	assert.Equal(t, 42, link.Number)
	assert.Equal(t, "acme/widgets", link.NWO())
}

func TestParseLink_PR(t *testing.T) {
	link, err := ParseLink("https://github.com/acme/widgets/pull/99")
	require.NoError(t, err)

	assert.Equal(t, KindPR, link.Kind)
	assert.Equal(t, "acme", link.Owner)
	assert.Equal(t, "widgets", link.Repo)
	assert.Equal(t, 99, link.Number)
}

func TestParseLink_TrailingSlash(t *testing.T) {
	link, err := ParseLink("https://github.com/acme/widgets/issues/7/")
	require.NoError(t, err)

	assert.Equal(t, KindIssue, link.Kind)
	assert.Equal(t, 7, link.Number)
}

func TestParseLink_Errors(t *testing.T) {
	tests := []struct {
		name string
		url  string
	}{
		{name: "wrong host", url: "https://gitlab.com/acme/widgets/issues/1"},
		{name: "missing number", url: "https://github.com/acme/widgets/issues"},
		{name: "bad kind", url: "https://github.com/acme/widgets/commits/123"},
		{name: "not a number", url: "https://github.com/acme/widgets/issues/abc"},
		{name: "too few segments", url: "https://github.com/acme/widgets"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseLink(tt.url)
			assert.Error(t, err)
		})
	}
}

func TestParseLink_ExtraPathSegments(t *testing.T) {
	// GitHub PR URLs often include /files, /checks, /commits, etc.
	link, err := ParseLink("https://github.com/acme/widgets/pull/12/files")
	require.NoError(t, err)

	assert.Equal(t, KindPR, link.Kind)
	assert.Equal(t, 12, link.Number)
}

func TestParseLink_QueryAndFragment(t *testing.T) {
	// GitHub URLs often have query params and fragments.
	link, err := ParseLink("https://github.com/acme/widgets/pull/55?diff=unified#discussion_r123")
	require.NoError(t, err)

	assert.Equal(t, KindPR, link.Kind)
	assert.Equal(t, 55, link.Number)
}
