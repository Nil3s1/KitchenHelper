package recipes

import (
	"fmt"
	"math"
	"strings"
	"time"
)

type Ingredient struct {
	Name   string
	Amount float64
	Unit   string
}

type Recipe struct {
	ID          string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	Name        string
	Description string
	Author      string
	Version     int
	Text        string
	Tags        []string
	Ingredients []Ingredient
	Images      []string
}

func newRecipe(name, description, author, text string, tags []string, ingredients []Ingredient, images []string) (Recipe, error) {
	recipe := Recipe{
		Name:        name,
		Description: description,
		Author:      author,
		Version:     1,
		Text:        text,
		Tags:        nonNilStrings(tags),
		Ingredients: nonNilIngredients(ingredients),
		Images:      nonNilStrings(images),
	}
	if err := recipe.validate(); err != nil {
		return Recipe{}, err
	}
	id, err := newUUID()
	if err != nil {
		return Recipe{}, fmt.Errorf("generate recipe ID: %w", err)
	}
	now := time.Now().UTC()
	recipe.ID = id
	recipe.CreatedAt = now
	recipe.UpdatedAt = now
	return recipe, nil
}

func (r *Recipe) replace(name, description, author, text string, tags []string, ingredients []Ingredient, images []string) error {
	replacement := Recipe{
		ID:          r.ID,
		CreatedAt:   r.CreatedAt,
		Name:        name,
		Description: description,
		Author:      author,
		Version:     r.Version + 1,
		Text:        text,
		Tags:        nonNilStrings(tags),
		Ingredients: nonNilIngredients(ingredients),
		Images:      nonNilStrings(images),
	}
	if err := replacement.validate(); err != nil {
		return err
	}
	replacement.UpdatedAt = time.Now().UTC()
	*r = replacement
	return nil
}

func (r Recipe) validate() error {
	if strings.TrimSpace(r.Name) == "" {
		return fmt.Errorf("%w: recipe name must not be empty", ErrValidation)
	}
	ingredientNames := make(map[string]struct{}, len(r.Ingredients))
	for _, ingredient := range r.Ingredients {
		name := strings.TrimSpace(ingredient.Name)
		if name == "" {
			return fmt.Errorf("%w: ingredient name must not be empty", ErrValidation)
		}
		if strings.TrimSpace(ingredient.Unit) == "" {
			return fmt.Errorf("%w: ingredient unit must not be empty", ErrValidation)
		}
		if math.IsNaN(ingredient.Amount) || math.IsInf(ingredient.Amount, 0) || ingredient.Amount < 0 {
			return fmt.Errorf("%w: ingredient amount must be a finite non-negative number", ErrValidation)
		}
		normalizedName := strings.ToLower(name)
		if _, exists := ingredientNames[normalizedName]; exists {
			return fmt.Errorf("%w: ingredient names must be unique within a recipe", ErrValidation)
		}
		ingredientNames[normalizedName] = struct{}{}
	}
	return nil
}

func nonNilStrings(values []string) []string {
	if values == nil {
		return []string{}
	}
	return append([]string{}, values...)
}

func nonNilIngredients(values []Ingredient) []Ingredient {
	if values == nil {
		return []Ingredient{}
	}
	return append([]Ingredient{}, values...)
}
