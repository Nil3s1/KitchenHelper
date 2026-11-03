package recipes

import (
	"context"
	"testing"
)

type testRecipeCommandRepository struct {
	recipe    Recipe
	created   Recipe
	updated   Recipe
	createErr error
	updateErr error
}

func (r *testRecipeCommandRepository) getByID(context.Context, string) (Recipe, error) {
	return r.recipe, nil
}

func (r *testRecipeCommandRepository) create(_ context.Context, recipe Recipe) error {
	r.created = recipe
	return r.createErr
}

func (r *testRecipeCommandRepository) update(_ context.Context, recipe Recipe) error {
	r.updated = recipe
	return r.updateErr
}

type testUnitOfWork struct {
	repository recipeCommandRepository
}

func (u testUnitOfWork) Do(
	ctx context.Context,
	work func(context.Context, recipeCommandRepository) error,
) error {
	return work(ctx, u.repository)
}

func TestAddRecipeHandlerCreatesAndPersistsRecipe(t *testing.T) {
	repository := &testRecipeCommandRepository{}
	handler := newAddRecipeHandler(repository)

	created, err := handler.Handle(context.Background(), AddRecipeCommand{
		Name:        "Suppe",
		Ingredients: []IngredientInput{{Name: "Salz", Amount: 1, Unit: "g"}},
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if created.ID == "" || created.Version != 1 {
		t.Fatalf("created recipe = %#v, want an identified version-one recipe", created)
	}
	if repository.created.ID != created.ID || repository.created.Name != created.Name {
		t.Fatalf("persisted recipe = %#v, want created recipe %#v", repository.created, created)
	}
}

func TestUpdateRecipeHandlerLoadsAndPersistsReplacement(t *testing.T) {
	original, err := newRecipe("Suppe", "", "", "", nil, nil, nil)
	if err != nil {
		t.Fatalf("newRecipe() error = %v", err)
	}
	repository := &testRecipeCommandRepository{recipe: original}
	handler := newUpdateRecipeHandler(testUnitOfWork{repository: repository})

	updated, err := handler.Handle(context.Background(), UpdateRecipeCommand{
		ID:   original.ID,
		Name: "Neue Suppe",
	})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if updated.ID != original.ID || !updated.CreatedAt.Equal(original.CreatedAt) ||
		updated.Name != "Neue Suppe" || updated.Version != original.Version+1 {
		t.Fatalf("updated recipe = %#v, does not preserve identity or increment version", updated)
	}
	if repository.updated.ID != updated.ID || repository.updated.Version != updated.Version {
		t.Fatalf("persisted recipe = %#v, want updated recipe %#v", repository.updated, updated)
	}
}
