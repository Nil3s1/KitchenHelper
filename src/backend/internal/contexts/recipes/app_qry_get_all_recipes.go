package recipes

import "context"

type getAllRecipesHandler struct {
	repository recipeQueryRepository
}

func newGetAllRecipesHandler(repository recipeQueryRepository) *getAllRecipesHandler {
	return &getAllRecipesHandler{repository: repository}
}

func (h *getAllRecipesHandler) Handle(ctx context.Context) ([]RecipeReadModel, error) {
	return h.repository.getAll(ctx)
}
