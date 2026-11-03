package recipes

import (
	"testing"
	"time"
)

func TestReadModelFromPersistenceProjectsPayloadWithoutDomainAggregate(t *testing.T) {
	createdAt := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	readModel, err := readModelFromPersistence(
		"recipe-id",
		createdAt,
		updatedAt,
		[]byte(`{"name":"Suppe","description":"Einfach","author":"Ada","version":2,"text":"Kochen","tags":["warm"],"ingredients":[{"name":"Salz","amount":1.5,"unit":"g"}],"images":["recipes/suppe.jpg"]}`),
	)
	if err != nil {
		t.Fatalf("readModelFromPersistence() error = %v", err)
	}
	if readModel.ID != "recipe-id" ||
		!readModel.CreatedAt.Equal(createdAt) ||
		!readModel.UpdatedAt.Equal(updatedAt) ||
		readModel.Name != "Suppe" ||
		readModel.Version != 2 ||
		len(readModel.Ingredients) != 1 ||
		readModel.Ingredients[0] != (IngredientReadModel{Name: "Salz", Amount: 1.5, Unit: "g"}) {
		t.Fatalf("read model = %#v, want complete query projection", readModel)
	}
}

func TestReadModelFromPersistenceInitializesEmptyCollections(t *testing.T) {
	readModel, err := readModelFromPersistence("recipe-id", time.Time{}, time.Time{}, []byte(`{"name":"Suppe"}`))
	if err != nil {
		t.Fatalf("readModelFromPersistence() error = %v", err)
	}
	if readModel.Tags == nil || readModel.Ingredients == nil || readModel.Images == nil {
		t.Fatalf("empty read model collections = %#v, want non-nil slices", readModel)
	}
}
