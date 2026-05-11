package render_test

import (
	"io/fs"
	"testing"

	"github.com/stretchr/testify/require"

	"findus/internal/transport/http/render"
	"findus/web"
)

func TestParseWebTemplates(t *testing.T) {
	sub, err := fs.Sub(web.Assets, "templates")
	require.NoError(t, err)
	_, err = render.Parse(sub)
	require.NoError(t, err)
}
