package templates

import (
	"context"
	"net/http"

	"github.com/a-h/templ"
)

func Render(ctx context.Context, w http.ResponseWriter, status int, component templ.Component) error {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(status)
	if component == nil {
		return nil
	}
	return component.Render(ctx, w)
}
