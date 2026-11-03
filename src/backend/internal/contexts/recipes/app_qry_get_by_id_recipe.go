package recipes

import (
	"context"
	"time"
)

type GetByIDRecipeQuery struct {
	ID string
}

type IngredientReadModel struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}

type RecipeReadModel struct {
	ID          string                 `json:"id"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Author      string                 `json:"author"`
	Version     int                    `json:"version"`
	Text        string                 `json:"text"`
	Tags        []string               `json:"tags"`
	Ingredients []IngredientReadModel  `json:"ingredients"`
	Images      []string               `json:"images"`
}

type GetByIDRecipeHandler struct {
	repository recipeReadRepository
}

func newGetByIDRecipeHandler(repository recipeReadRepository) *GetByIDRecipeHandler {
	return &GetByIDRecipeHandler{repository: repository}
}

func (h *GetByIDRecipeHandler) Handle(ctx context.Context, query GetByIDRecipeQuery) (RecipeReadModel, error) {
	recipe, err := h.repository.getByID(ctx, query.ID)
	if err != nil {
		return RecipeReadModel{}, err
	}
	return newRecipeReadModel(recipe), nil
}

func newRecipeReadModel(recipe Recipe) RecipeReadModel {
	ingredients := make([]IngredientReadModel, 0, len(recipe.Ingredients))
	for _, ingredient := range recipe.Ingredients {
		ingredients = append(ingredients, IngredientReadModel{
			Name:   ingredient.Name,
			Amount: ingredient.Amount,
			Unit:   ingredient.Unit,
		})
	}
	return RecipeReadModel{
		ID:          recipe.ID,
		CreatedAt:   recipe.CreatedAt,
		UpdatedAt:   recipe.UpdatedAt,
		Name:        recipe.Name,
		Description: recipe.Description,
		Author:      recipe.Author,
		Version:     recipe.Version,
		Text:        recipe.Text,
		Tags:        recipe.Tags,
		Ingredients: ingredients,
		Images:      recipe.Images,
	}
}
