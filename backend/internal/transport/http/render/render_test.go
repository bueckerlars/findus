package render_test

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"

	"findus/backend/internal/transport/http/render"
	"findus/frontend"
)

func TestParseWebTemplates(t *testing.T) {
	sub, err := fs.Sub(frontend.Assets, "templates")
	require.NoError(t, err)
	_, err = render.Parse(sub)
	require.NoError(t, err)
}
