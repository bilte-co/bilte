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

func TestCVRendersProductsAfterExperience(t *testing.T) {
	production := false
	resume := domain.Resume{
		Experience: []domain.Experience{{Name: "Example Company", Role: "Example Role"}},
		Products: []domain.Product{{
			Name:         "socky.flights",
			Role:         "Independent Product",
			Description:  "Real-time aviation intelligence.",
			Technologies: []string{"Go", "NATS"},
		}},
		OtherExperience: []domain.OtherExperience{{Name: "Earlier Company", Role: "Earlier Role"}},
	}

	var rendered strings.Builder
	if err := CV(&production, resume, nil).Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render CV: %v", err)
	}

	html := rendered.String()
	for _, text := range []string{
		"Selected Products &amp; Ventures",
		"socky.flights",
		"Independent Product",
		"Real-time aviation intelligence.",
		"[Go, NATS]",
	} {
		if !strings.Contains(html, text) {
			t.Errorf("expected rendered CV to contain %q", text)
		}
	}

	experienceIndex := strings.Index(html, ">Experience</h3>")
	productsIndex := strings.Index(html, ">Selected Products &amp; Ventures</h3>")
	otherExperienceIndex := strings.Index(html, ">Other Professional Experience</h3>")
	if experienceIndex < 0 || productsIndex < 0 || otherExperienceIndex < 0 {
		t.Fatal("expected experience, products, and other experience headings")
	}
	if !(experienceIndex < productsIndex && productsIndex < otherExperienceIndex) {
		t.Fatal("expected products section between experience and other experience")
	}
}

func TestCVOmitsProductsSectionWhenEmpty(t *testing.T) {
	production := false
	resume := domain.Resume{Name: "Example Name"}

	var rendered strings.Builder
	if err := CV(&production, resume, nil).Render(context.Background(), &rendered); err != nil {
		t.Fatalf("render CV: %v", err)
	}

	if strings.Contains(rendered.String(), "Selected Products") {
		t.Fatal("expected products section to be omitted when no products are present")
	}
}
