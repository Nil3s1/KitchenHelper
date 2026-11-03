package recipes

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDecodeRecipeRequestRequiresIngredientAmount(t *testing.T) {
	request := httptest.NewRequest(
		"POST",
		"/api/v1/recipes",
		strings.NewReader(`{"name":"Suppe","ingredients":[{"name":"Salz","unit":"g"}]}`),
	)

	_, err := decodeRecipeRequest(httptest.NewRecorder(), request)
	if !errors.Is(err, ErrValidation) {
		t.Fatalf("decodeRecipeRequest() error = %v, want validation error", err)
	}
}

func TestDecodeRecipeRequestAcceptsZeroIngredientAmount(t *testing.T) {
	request := httptest.NewRequest(
		"POST",
		"/api/v1/recipes",
		strings.NewReader(`{"name":"Suppe","ingredients":[{"name":"Salz","amount":0,"unit":"g"}]}`),
	)

	decoded, err := decodeRecipeRequest(httptest.NewRecorder(), request)
	if err != nil {
		t.Fatalf("decodeRecipeRequest() error = %v", err)
	}
	command := addCommandFromRequest(decoded)
	if len(command.Ingredients) != 1 || command.Ingredients[0].Amount != 0 {
		t.Fatalf("mapped ingredients = %#v, want one ingredient with amount 0", command.Ingredients)
	}
}

func TestAddCommandMappingCopiesTransportFields(t *testing.T) {
	request := recipeRequest{
		Name: "Suppe",
		Ingredients: []ingredientRequest{{
			Name: "Salz",
			Unit: "g",
			Amount: func() *float64 {
				amount := 2.5
				return &amount
			}(),
		}},
	}

	command := addCommandFromRequest(request)
	if command.Name != request.Name ||
		len(command.Ingredients) != 1 ||
		command.Ingredients[0].Amount != 2.5 {
		t.Fatalf("mapped command = %#v, fields not copied from transport DTO", command)
	}
}
