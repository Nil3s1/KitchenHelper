package recipes

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

func TestModuleCommandsPreserveDomainBehavior(t *testing.T) {
	repository := &testRecipeCommandRepository{}
	module := &Module{
		addRecipe:    newAddRecipeHandler(repository),
		updateRecipe: newUpdateRecipeHandler(testUnitOfWork{repository: repository}),
	}
	ctx := context.Background()
	created, err := module.AddRecipe(ctx, AddRecipeCommand{
		Name:        "Suppe",
		Description: "Einfach",
		Author:      "Ada",
		Text:        "Kochen",
		Tags:        []string{"warm"},
		Ingredients: []IngredientInput{{Name: "Salz", Amount: 0, Unit: "g"}},
		Images:      []string{"recipes/suppe.jpg"},
	})
	if err != nil {
		t.Fatalf("AddRecipe() error = %v", err)
	}
	if !reflect.DeepEqual(created, repository.created) || created.ID == "" ||
		created.Version != 1 || created.Name != "Suppe" || created.CreatedAt.IsZero() {
		t.Fatalf("created recipe = %#v, want persisted version-one recipe", created)
	}
	repository.recipe = created
	updated, err := module.UpdateRecipe(ctx, UpdateRecipeCommand{
		ID:          created.ID,
		Name:        created.Name,
		Description: created.Description,
		Author:      created.Author,
		Text:        created.Text,
		Tags:        created.Tags,
		Ingredients: []IngredientInput{{Name: "Salz", Amount: 0, Unit: "g"}},
		Images:      created.Images,
	})
	if err != nil {
		t.Fatalf("UpdateRecipe() error = %v", err)
	}
	if !reflect.DeepEqual(updated, repository.updated) || updated.ID != created.ID ||
		!updated.CreatedAt.Equal(created.CreatedAt) || updated.Version != 2 {
		t.Fatalf("updated recipe = %#v, want same identity and incremented version", updated)
	}
}

func TestModuleCommandsReturnErrors(t *testing.T) {
	writeErr := errors.New("write failed")
	original, err := newRecipe("Suppe", "", "", "", nil, nil, nil)
	if err != nil {
		t.Fatalf("create original: %v", err)
	}
	repository := &testRecipeCommandRepository{recipe: original, createErr: writeErr, updateErr: writeErr}
	module := &Module{
		addRecipe:    newAddRecipeHandler(repository),
		updateRecipe: newUpdateRecipeHandler(testUnitOfWork{repository: repository}),
	}
	tests := []struct {
		name string
		call func() (Recipe, error)
		want error
	}{
		{name: "invalid create", call: func() (Recipe, error) {
			return module.AddRecipe(context.Background(), AddRecipeCommand{})
		}, want: ErrValidation},
		{name: "invalid update", call: func() (Recipe, error) {
			return module.UpdateRecipe(context.Background(), UpdateRecipeCommand{ID: original.ID})
		}, want: ErrValidation},
		{name: "failed create", call: func() (Recipe, error) {
			return module.AddRecipe(context.Background(), AddRecipeCommand{Name: "Suppe"})
		}, want: writeErr},
		{name: "failed update", call: func() (Recipe, error) {
			return module.UpdateRecipe(context.Background(), UpdateRecipeCommand{ID: original.ID, Name: "Neu"})
		}, want: writeErr},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := test.call()
			if !errors.Is(err, test.want) || !reflect.DeepEqual(result, Recipe{}) {
				t.Fatalf("result = %#v, error = %v; want zero result, error %v", result, err, test.want)
			}
		})
	}
}

func TestModuleQueriesReturnDTOsAndForwardContext(t *testing.T) {
	expected := RecipeReadModel{
		ID: "recipe-id", Name: "Suppe", Version: 2,
		Tags: []string{}, Ingredients: []IngredientReadModel{}, Images: []string{},
	}
	repository := &testRecipeQueryRepository{byID: expected, all: []RecipeReadModel{expected}}
	module := &Module{queries: repository}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	recipe, err := module.GetRecipe(ctx, expected.ID)
	if err != nil || !reflect.DeepEqual(recipe, expected) {
		t.Fatalf("GetRecipe() = %#v, %v; want %#v, nil", recipe, err, expected)
	}
	if repository.ctx != ctx || repository.id != expected.ID {
		t.Fatal("GetRecipe did not forward request context and ID")
	}
	repository.ctx = nil
	list, err := module.ListRecipes(ctx)
	if err != nil || !reflect.DeepEqual(list, repository.all) || repository.ctx != ctx {
		t.Fatalf("ListRecipes() = %#v, %v; want projections with request context", list, err)
	}
	repository.all = []RecipeReadModel{}
	list, err = module.ListRecipes(ctx)
	if err != nil || list == nil || len(list) != 0 {
		t.Fatalf("empty ListRecipes() = %#v, %v; want non-nil empty slice, nil", list, err)
	}
}

func TestModuleQueriesReturnErrors(t *testing.T) {
	for _, expected := range []error{ErrNotFound, errors.New("query failed")} {
		repository := &testRecipeQueryRepository{err: expected}
		module := &Module{queries: repository}
		recipe, err := module.GetRecipe(context.Background(), "missing")
		if !errors.Is(err, expected) || !reflect.DeepEqual(recipe, RecipeReadModel{}) {
			t.Fatalf("GetRecipe() = %#v, %v; want zero result, %v", recipe, err, expected)
		}
		list, err := module.ListRecipes(context.Background())
		if !errors.Is(err, expected) || list != nil {
			t.Fatalf("ListRecipes() = %#v, %v; want nil, %v", list, err, expected)
		}
	}
}
