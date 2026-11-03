package recipes

import (
	"context"
	"errors"
	"reflect"
	"testing"
)

type testRecipeQueryRepository struct {
	ctx  context.Context
	id   string
	byID RecipeReadModel
	all  []RecipeReadModel
	err  error
}

func (r *testRecipeQueryRepository) getByID(ctx context.Context, id string) (RecipeReadModel, error) {
	r.ctx = ctx
	r.id = id
	return r.byID, r.err
}

func (r *testRecipeQueryRepository) getAll(ctx context.Context) ([]RecipeReadModel, error) {
	r.ctx = ctx
	return r.all, r.err
}

func TestGetByIDRecipeHandlerReturnsDTOAndForwardsInput(t *testing.T) {
	expected := RecipeReadModel{
		ID: "recipe-id", Name: "Suppe", Version: 2,
		Tags:        []string{"warm"},
		Ingredients: []IngredientReadModel{{Name: "Salz", Amount: 0, Unit: "g"}},
		Images:      []string{},
	}
	repository := &testRecipeQueryRepository{byID: expected}
	handler := newGetByIDRecipeHandler(repository)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	result, err := handler.Handle(ctx, expected.ID)
	if err != nil || !reflect.DeepEqual(result, expected) {
		t.Fatalf("Handle() = %#v, %v; want %#v, nil", result, err, expected)
	}
	if repository.ctx != ctx || repository.id != expected.ID {
		t.Fatal("handler did not forward context and ID")
	}
}

func TestGetAllRecipesHandlerReturnsOrderedDTOs(t *testing.T) {
	tests := []struct {
		name string
		all  []RecipeReadModel
	}{
		{name: "ordered list", all: []RecipeReadModel{{ID: "second"}, {ID: "first"}}},
		{name: "empty list", all: []RecipeReadModel{}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &testRecipeQueryRepository{all: test.all}
			handler := newGetAllRecipesHandler(repository)
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			result, err := handler.Handle(ctx)
			if err != nil || !reflect.DeepEqual(result, test.all) {
				t.Fatalf("Handle() = %#v, %v; want %#v, nil", result, err, test.all)
			}
			if repository.ctx != ctx {
				t.Fatal("handler did not forward context")
			}
		})
	}
}

func TestQueryHandlersPropagateErrors(t *testing.T) {
	for _, expected := range []error{ErrNotFound, errors.New("query failed")} {
		repository := &testRecipeQueryRepository{err: expected}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		result, err := newGetByIDRecipeHandler(repository).Handle(ctx, "missing")
		if !errors.Is(err, expected) || !reflect.DeepEqual(result, RecipeReadModel{}) ||
			repository.ctx != ctx || repository.id != "missing" {
			t.Fatalf("lookup = %#v, %v; want zero result, %v with original inputs", result, err, expected)
		}
		repository.ctx = nil
		list, err := newGetAllRecipesHandler(repository).Handle(ctx)
		if !errors.Is(err, expected) || list != nil || repository.ctx != ctx {
			t.Fatalf("list = %#v, %v; want nil, %v with original context", list, err, expected)
		}
	}
}
