package templates

import (
	"context"
	"strings"
	"testing"

	"github.com/bilte-co/bilte/internal/domain"
)

func TestCVRendersPublicationURL(t *testing.T) {
	production := false
	resume := domain.Resume{
		Name:  "Example Name",
		Title: "Example Title",
		Publications: []domain.Publication{
			{URL: "https://example.com/publication"},
		},
	}

	var rendered strings.Builder
	if err := CV(&production, resume, nil).Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render CV: %v", err)
	}

	const link = `<a href="https://example.com/publication" target="_blank">https://example.com/publication</a>`
	if !strings.Contains(rendered.String(), link) {
		t.Fatalf("expected rendered publication link %q", link)
	}
	if strings.Contains(rendered.String(), `{item.URL}`) {
		t.Fatal("rendered CV contains the literal publication URL expression")
	}
}
