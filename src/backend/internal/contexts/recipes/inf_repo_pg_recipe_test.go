package recipes

import (
	"encoding/json"
	"testing"
	"time"

	"kitchenhelper/backend/internal/shared"
)

func TestPersistenceMappingRoundTripsRecipe(t *testing.T) {
	createdAt := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	original := Recipe{
		DomainModel: shared.DomainModel{
			ID:        "recipe-id",
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
		Name:        "Suppe",
		Description: "Einfach",
		Author:      "Ada",
		Version:     3,
		Text:        "Kochen",
		Tags:        []string{"warm"},
		Ingredients: []Ingredient{{Name: "Salz", Amount: 1.5, Unit: "g"}},
		Images:      []string{"recipes/suppe.jpg"},
	}

	model := persistenceFromDomain(original)
	encoded, err := json.Marshal(model.Payload)
	if err != nil {
		t.Fatalf("marshal persistence payload: %v", err)
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(encoded, &fields); err != nil {
		t.Fatalf("unmarshal persistence payload: %v", err)
	}
	for _, field := range []string{"name", "description", "author", "version", "text", "tags", "ingredients", "images"} {
		if _, exists := fields[field]; !exists {
			t.Errorf("persistence payload is missing %q", field)
		}
	}
	if _, exists := fields["id"]; exists {
		t.Fatal("recipe ID must not be stored in the JSON document")
	}

	restored := domainFromPersistence(model)
	if restored.ID != original.ID ||
		!restored.CreatedAt.Equal(original.CreatedAt) ||
		!restored.UpdatedAt.Equal(original.UpdatedAt) ||
		restored.Name != original.Name ||
		restored.Version != original.Version ||
		len(restored.Ingredients) != 1 ||
		restored.Ingredients[0] != original.Ingredients[0] {
		t.Fatalf("domain mapping did not round-trip: got %#v", restored)
	}
}
