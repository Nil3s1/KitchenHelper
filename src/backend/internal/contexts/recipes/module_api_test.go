package recipes_test

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"kitchenhelper/backend/internal/contexts/recipes"
)

type unavailableDatabase struct {
	err error
}

func (d unavailableDatabase) Connect(context.Context) (driver.Conn, error) {
	return nil, d.err
}

func (d unavailableDatabase) Driver() driver.Driver { return d }

func (d unavailableDatabase) Open(string) (driver.Conn, error) {
	return nil, d.err
}

func TestPublicModuleWorksWithoutHTTP(t *testing.T) {
	databaseErr := errors.New("database unavailable")
	db := sql.OpenDB(unavailableDatabase{err: databaseErr})
	defer db.Close()
	module := recipes.NewModule(db)
	ctx := context.Background()

	if _, err := module.AddRecipe(ctx, recipes.AddRecipeCommand{
		Name:        "Suppe",
		Ingredients: []recipes.IngredientInput{{Name: "Salz", Amount: 1, Unit: "g"}},
	}); !errors.Is(err, databaseErr) {
		t.Fatalf("AddRecipe() error = %v, want database error", err)
	}
	if _, err := module.UpdateRecipe(ctx, recipes.UpdateRecipeCommand{ID: "recipe-id", Name: "Neu"}); !errors.Is(err, databaseErr) {
		t.Fatalf("UpdateRecipe() error = %v, want database error", err)
	}
	var recipe recipes.RecipeReadModel
	recipe, err := module.GetRecipe(ctx, "recipe-id")
	if !errors.Is(err, databaseErr) || recipe.ID != "" {
		t.Fatalf("GetRecipe() = %#v, %v; want zero result, database error", recipe, err)
	}
	var list []recipes.RecipeReadModel
	list, err = module.ListRecipes(ctx)
	if !errors.Is(err, databaseErr) || list != nil {
		t.Fatalf("ListRecipes() = %#v, %v; want nil, database error", list, err)
	}
	if _, err := module.AddRecipe(ctx, recipes.AddRecipeCommand{}); !errors.Is(err, recipes.ErrValidation) {
		t.Fatalf("invalid AddRecipe() error = %v, want validation error", err)
	}

	response := httptest.NewRecorder()
	module.HTTPHandler().ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/v1/recipes", nil))
	if response.Code != http.StatusInternalServerError {
		t.Fatalf("HTTP status = %d, want 500 for unavailable database", response.Code)
	}
}
