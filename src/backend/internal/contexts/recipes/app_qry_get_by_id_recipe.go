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

type getByIDRecipeRepository interface {
	getByID(context.Context, string) (RecipeReadModel, error)
}

type GetByIDRecipeHandler struct {
	repository getByIDRecipeRepository
}

func newGetByIDRecipeHandler(repository getByIDRecipeRepository) *GetByIDRecipeHandler {
	return &GetByIDRecipeHandler{repository: repository}
}

func (h *GetByIDRecipeHandler) Handle(ctx context.Context, query GetByIDRecipeQuery) (RecipeReadModel, error) {
	return h.repository.getByID(ctx, query.ID)
}
