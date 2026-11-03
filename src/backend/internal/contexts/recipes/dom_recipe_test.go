package recipes

import (
	"errors"
	"reflect"
	"testing"
)

func TestNewRecipeCreatesValidVersionOneAggregate(t *testing.T) {
	recipe, err := newRecipe("Suppe", "", "", "", nil, []Ingredient{{Name: "Salz", Amount: 1, Unit: "g"}}, nil)
	if err != nil {
		t.Fatalf("newRecipe() error = %v", err)
	}
	if recipe.ID == "" {
		t.Fatal("new recipe must have an ID")
	}
	if recipe.Version != 1 {
		t.Fatalf("new recipe version = %d, want 1", recipe.Version)
	}
	if recipe.Tags == nil || recipe.Ingredients == nil || recipe.Images == nil {
		t.Fatal("recipe collections must be initialized")
	}
}

func TestRecipeReplaceIncrementsVersion(t *testing.T) {
	recipe, err := newRecipe("Suppe", "", "", "", nil, nil, nil)
	if err != nil {
		t.Fatalf("newRecipe() error = %v", err)
	}

	if err := recipe.replace("Neue Suppe", "", "", "", nil, nil, nil); err != nil {
		t.Fatalf("replace() error = %v", err)
	}
	if recipe.Name != "Neue Suppe" || recipe.Version != 2 {
		t.Fatalf("updated recipe = (%q, version %d), want (Neue Suppe, version 2)", recipe.Name, recipe.Version)
	}
}

func TestRecipeValidateEnforcesDomainRules(t *testing.T) {
	tests := []struct {
		name   string
		recipe Recipe
	}{
		{name: "missing recipe name", recipe: Recipe{}},
		{name: "missing ingredient name", recipe: Recipe{Name: "Suppe", Ingredients: []Ingredient{{Amount: 1, Unit: "g"}}}},
		{name: "missing ingredient unit", recipe: Recipe{Name: "Suppe", Ingredients: []Ingredient{{Name: "Salz", Amount: 1}}}},
		{name: "negative amount", recipe: Recipe{Name: "Suppe", Ingredients: []Ingredient{{Name: "Salz", Amount: -1, Unit: "g"}}}},
		{name: "duplicate ingredient names", recipe: Recipe{Name: "Suppe", Ingredients: []Ingredient{
			{Name: "Salz", Amount: 1, Unit: "g"},
			{Name: " salz ", Amount: 2, Unit: "g"},
		}}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			if err := test.recipe.validate(); !errors.Is(err, ErrValidation) {
				t.Fatalf("validate() error = %v, want validation error", err)
			}
		})
	}
}

func TestDomainEntitiesHaveNoStructTags(t *testing.T) {
	for _, entity := range []reflect.Type{reflect.TypeOf(Recipe{}), reflect.TypeOf(Ingredient{})} {
		for i := 0; i < entity.NumField(); i++ {
			if tag := entity.Field(i).Tag; tag != "" {
				t.Errorf("%s.%s has struct tag %q", entity.Name(), entity.Field(i).Name, tag)
			}
		}
	}
}
