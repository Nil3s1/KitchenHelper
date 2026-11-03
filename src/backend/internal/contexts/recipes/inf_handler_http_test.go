package recipes

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestHTTPQueryRoutesReturnDTOs(t *testing.T) {
	createdAt := time.Date(2026, 11, 3, 10, 0, 0, 0, time.UTC)
	recipe := RecipeReadModel{
		ID:          "recipe-id",
		CreatedAt:   createdAt,
		UpdatedAt:   createdAt.Add(time.Minute),
		Name:        "Suppe",
		Description: "Einfach",
		Author:      "Ada",
		Version:     2,
		Text:        "Kochen",
		Tags:        []string{"warm"},
		Ingredients: []IngredientReadModel{{Name: "Salz", Amount: 1.5, Unit: "g"}},
		Images:      []string{"recipes/suppe.jpg"},
	}
	other := recipe
	other.ID = "other-id"
	other.Tags = []string{}
	other.Ingredients = []IngredientReadModel{}
	other.Images = []string{}
	const recipeJSON = `{"id":"recipe-id","created_at":"2026-11-03T10:00:00Z","updated_at":"2026-11-03T10:01:00Z","name":"Suppe","description":"Einfach","author":"Ada","version":2,"text":"Kochen","tags":["warm"],"ingredients":[{"name":"Salz","amount":1.5,"unit":"g"}],"images":["recipes/suppe.jpg"]}`
	const otherJSON = `{"id":"other-id","created_at":"2026-11-03T10:00:00Z","updated_at":"2026-11-03T10:01:00Z","name":"Suppe","description":"Einfach","author":"Ada","version":2,"text":"Kochen","tags":[],"ingredients":[],"images":[]}`

	tests := []struct {
		name     string
		path     string
		all      []RecipeReadModel
		wantID   string
		wantJSON string
	}{
		{name: "by ID", path: "/api/v1/recipes/recipe-id", wantID: recipe.ID, wantJSON: recipeJSON},
		{name: "list preserves order", path: "/api/v1/recipes", all: []RecipeReadModel{recipe, other}, wantJSON: "[" + recipeJSON + "," + otherJSON + "]"},
		{name: "empty list", path: "/api/v1/recipes", all: []RecipeReadModel{}, wantJSON: "[]"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &testRecipeQueryRepository{byID: recipe, all: test.all}
			module := &Module{
				getRecipe:   newGetByIDRecipeHandler(repository),
				listRecipes: newGetAllRecipesHandler(repository),
			}
			handler := module.HTTPHandler()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			request := httptest.NewRequest(http.MethodGet, test.path, nil).WithContext(ctx)
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != http.StatusOK {
				t.Fatalf("status = %d, want 200; body = %s", response.Code, response.Body)
			}
			if contentType := response.Header().Get("Content-Type"); contentType != "application/json" {
				t.Errorf("Content-Type = %q, want application/json", contentType)
			}
			if repository.ctx != ctx || repository.id != test.wantID {
				t.Errorf("query received context %v, ID %q; want request context, ID %q", repository.ctx, repository.id, test.wantID)
			}
			var got, want any
			if err := json.Unmarshal(response.Body.Bytes(), &got); err != nil {
				t.Fatalf("decode response: %v", err)
			}
			if err := json.Unmarshal([]byte(test.wantJSON), &want); err != nil {
				t.Fatalf("decode expected response: %v", err)
			}
			if !reflect.DeepEqual(got, want) {
				t.Errorf("response = %s, want %s", response.Body, test.wantJSON)
			}
		})
	}
}

func TestHTTPQueryRoutesMapErrors(t *testing.T) {
	tests := []struct {
		name        string
		path        string
		err         error
		wantStatus  int
		wantMessage string
	}{
		{name: "missing recipe", path: "/api/v1/recipes/missing", err: ErrNotFound, wantStatus: http.StatusNotFound, wantMessage: "recipe not found"},
		{name: "failed lookup", path: "/api/v1/recipes/recipe-id", err: errors.New("private database error"), wantStatus: http.StatusInternalServerError, wantMessage: "internal server error"},
		{name: "failed list", path: "/api/v1/recipes", err: errors.New("private database error"), wantStatus: http.StatusInternalServerError, wantMessage: "internal server error"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			repository := &testRecipeQueryRepository{err: test.err}
			module := &Module{
				getRecipe:   newGetByIDRecipeHandler(repository),
				listRecipes: newGetAllRecipesHandler(repository),
			}
			handler := module.HTTPHandler()
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, httptest.NewRequest(http.MethodGet, test.path, nil))

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d", response.Code, test.wantStatus)
			}
			var body map[string]string
			if err := json.Unmarshal(response.Body.Bytes(), &body); err != nil {
				t.Fatalf("decode error response: %v", err)
			}
			if !reflect.DeepEqual(body, map[string]string{"error": test.wantMessage}) {
				t.Errorf("error response = %v, want only error %q", body, test.wantMessage)
			}
		})
	}
}

func TestHTTPCommandRoutesPreserveRecipeLifecycle(t *testing.T) {
	original, err := newRecipe("Suppe", "", "", "", nil, nil, nil)
	if err != nil {
		t.Fatalf("create original recipe: %v", err)
	}
	tests := []struct {
		method      string
		path        string
		wantStatus  int
		wantVersion int
	}{
		{method: http.MethodPost, path: "/api/v1/recipes", wantStatus: http.StatusCreated, wantVersion: 1},
		{method: http.MethodPut, path: "/api/v1/recipes/" + original.ID, wantStatus: http.StatusOK, wantVersion: 2},
	}
	for _, test := range tests {
		t.Run(test.method, func(t *testing.T) {
			repository := &testRecipeCommandRepository{recipe: original}
			module := &Module{
				addRecipe:    newAddRecipeHandler(repository),
				updateRecipe: newUpdateRecipeHandler(testUnitOfWork{repository: repository}),
			}
			handler := module.HTTPHandler()
			request := httptest.NewRequest(test.method, test.path, strings.NewReader(`{"name":"Neue Suppe","ingredients":[{"name":"Salz","amount":0,"unit":"g"}]}`))
			response := httptest.NewRecorder()

			handler.ServeHTTP(response, request)

			if response.Code != test.wantStatus {
				t.Fatalf("status = %d, want %d; body = %s", response.Code, test.wantStatus, response.Body)
			}
			var result recipeResponse
			if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
				t.Fatalf("decode command response: %v", err)
			}
			if result.ID == "" || result.CreatedAt.IsZero() || result.UpdatedAt.IsZero() ||
				result.Name != "Neue Suppe" || result.Version != test.wantVersion ||
				len(result.Ingredients) != 1 || result.Ingredients[0] != (ingredientResponse{Name: "Salz", Amount: 0, Unit: "g"}) ||
				result.Tags == nil || result.Images == nil {
				t.Fatalf("unexpected command response: %#v", result)
			}
			persisted := repository.created
			if test.method == http.MethodPut {
				persisted = repository.updated
				if result.ID != original.ID || !result.CreatedAt.Equal(original.CreatedAt) {
					t.Errorf("update changed recipe identity or creation time: %#v", result)
				}
			}
			if persisted.ID != result.ID || persisted.Version != result.Version {
				t.Errorf("persisted recipe = %#v, does not match response", persisted)
			}
		})
	}
}

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
