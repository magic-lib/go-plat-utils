package templates

import (
	"bytes"
	"context"
	"github.com/hairyhenderson/gomplate/v3"
	"github.com/hairyhenderson/gomplate/v3/conv"
	"github.com/hairyhenderson/gomplate/v3/env"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	tl "text/template"
)

func testTemplate(t *testing.T, tr *gomplate.Renderer, tmpl string) string {
	t.Helper()

	var out bytes.Buffer
	err := tr.Render(context.Background(), "testtemplate", tmpl, &out)
	assert.NoError(t, err)

	return out.String()
}

func TestGetenvTemplates(t *testing.T) {
	tr := gomplate.NewRenderer(gomplate.Options{
		Funcs: tl.FuncMap{
			"getenv": env.Getenv,
			"bool":   conv.Bool,
		},
	})
	assert.Empty(t, testTemplate(t, tr, `{{getenv "BLAHBLAHBLAH"}}`))
	assert.Equal(t, os.Getenv("USER"), testTemplate(t, tr, `{{getenv "USER"}}`))
	assert.Equal(t, "default value", testTemplate(t, tr, `{{getenv "BLAHBLAHBLAH" "default value"}}`))
}
