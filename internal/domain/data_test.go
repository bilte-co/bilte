package domain

import (
	"encoding/json"
	"testing"
)

func TestResumeJSONUsesEducationAndPublicationsKeys(t *testing.T) {
	data, err := json.Marshal(Resume{
		Education:    []Education{{Name: "Example University"}},
		Publications: []Publication{{Title: "Example Publication"}},
	})
	if err != nil {
		t.Fatalf("marshal resume: %v", err)
	}

	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		t.Fatalf("unmarshal resume object: %v", err)
	}

	for _, key := range []string{"education", "publications"} {
		if _, ok := fields[key]; !ok {
			t.Errorf("expected JSON field %q", key)
		}
	}
	for _, key := range []string{"Education", "Publications"} {
		if _, ok := fields[key]; ok {
			t.Errorf("unexpected JSON field %q", key)
		}
	}
}
