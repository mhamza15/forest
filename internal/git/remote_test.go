package git

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeRemoteURL(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want string
	}{
		{
			name: "https with .git",
			raw:  "https://github.com/acme/widgets.git",
			want: "acme/widgets",
		},
		{
			name: "https without .git",
			raw:  "https://github.com/acme/widgets",
			want: "acme/widgets",
		},
		{
			name: "https trailing slash",
			raw:  "https://github.com/acme/widgets/",
			want: "acme/widgets",
		},
		{
			name: "ssh with .git",
			raw:  "git@github.com:acme/widgets.git",
			want: "acme/widgets",
		},
		{
			name: "ssh without .git",
			raw:  "git@github.com:acme/widgets",
			want: "acme/widgets",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, NormalizeRemoteURL(tt.raw))
		})
	}
}
