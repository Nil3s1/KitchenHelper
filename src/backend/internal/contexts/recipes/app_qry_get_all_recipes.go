package recipes

import "context"

type GetAllRecipesQuery struct{}

type GetAllRecipesReadModel []RecipeReadModel

type recipeQueryRepository interface {
	getByID(context.Context, string) (RecipeReadModel, error)
	getAll(context.Context) (GetAllRecipesReadModel, error)
}

type GetAllRecipesHandler struct {
	repository recipeQueryRepository
}

func newGetAllRecipesHandler(repository recipeQueryRepository) *GetAllRecipesHandler {
	return &GetAllRecipesHandler{repository: repository}
}

func (h *GetAllRecipesHandler) Handle(ctx context.Context, _ GetAllRecipesQuery) (GetAllRecipesReadModel, error) {
	return h.repository.getAll(ctx)
}
