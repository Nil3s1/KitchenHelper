package recipes

import (
	"context"
	"reflect"
	"testing"
)

type testRecipeQueryRepository struct {
	id     string
	byID   RecipeReadModel
	all    GetAllRecipesReadModel
}

func (r *testRecipeQueryRepository) getByID(_ context.Context, id string) (RecipeReadModel, error) {
	r.id = id
	return r.byID, nil
}

func (r *testRecipeQueryRepository) getAll(context.Context) (GetAllRecipesReadModel, error) {
	return r.all, nil
}

func TestGetByIDRecipeHandlerReturnsQueryProjection(t *testing.T) {
	expected := RecipeReadModel{ID: "recipe-id", Name: "Suppe", Version: 2}
	repository := &testRecipeQueryRepository{byID: expected}
	handler := newGetByIDRecipeHandler(repository)

	result, err := handler.Handle(context.Background(), GetByIDRecipeQuery{ID: expected.ID})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if repository.id != expected.ID || !reflect.DeepEqual(result, expected) {
		t.Fatalf("query result = %#v, repository ID = %q; want %#v for %q", result, repository.id, expected, expected.ID)
	}
}

func TestGetAllRecipesHandlerReturnsQueryProjections(t *testing.T) {
	expected := GetAllRecipesReadModel{{ID: "recipe-id", Name: "Suppe", Version: 2}}
	handler := newGetAllRecipesHandler(&testRecipeQueryRepository{all: expected})

	result, err := handler.Handle(context.Background(), GetAllRecipesQuery{})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("query result = %#v, want %#v", result, expected)
	}
}
