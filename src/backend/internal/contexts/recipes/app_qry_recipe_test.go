package recipes

import (
	"context"
	"reflect"
	"testing"
)

type testGetByIDRecipeRepository struct {
	id     string
	result RecipeReadModel
}

func (r *testGetByIDRecipeRepository) getByID(_ context.Context, id string) (RecipeReadModel, error) {
	r.id = id
	return r.result, nil
}

type testGetAllRecipesRepository struct {
	result GetAllRecipesReadModel
}

func (r testGetAllRecipesRepository) getAll(context.Context) (GetAllRecipesReadModel, error) {
	return r.result, nil
}

func TestGetByIDRecipeHandlerReturnsQueryProjection(t *testing.T) {
	expected := RecipeReadModel{ID: "recipe-id", Name: "Suppe", Version: 2}
	repository := &testGetByIDRecipeRepository{result: expected}
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
	handler := newGetAllRecipesHandler(testGetAllRecipesRepository{result: expected})

	result, err := handler.Handle(context.Background(), GetAllRecipesQuery{})
	if err != nil {
		t.Fatalf("Handle() error = %v", err)
	}
	if !reflect.DeepEqual(result, expected) {
		t.Fatalf("query result = %#v, want %#v", result, expected)
	}
}
