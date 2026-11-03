package recipes

import (
	"context"
	"time"
)

type recipeQueryRepository interface {
	getByID(context.Context, string) (RecipeReadModel, error)
	getAll(context.Context) ([]RecipeReadModel, error)
}

type IngredientReadModel struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}

type RecipeReadModel struct {
	ID          string                `json:"id"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Author      string                `json:"author"`
	Version     int                   `json:"version"`
	Text        string                `json:"text"`
	Tags        []string              `json:"tags"`
	Ingredients []IngredientReadModel `json:"ingredients"`
	Images      []string              `json:"images"`
}
