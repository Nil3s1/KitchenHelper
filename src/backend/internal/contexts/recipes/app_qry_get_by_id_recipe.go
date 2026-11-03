package recipes

import "context"

type getByIDRecipeHandler struct {
	repository recipeQueryRepository
}

func newGetByIDRecipeHandler(repository recipeQueryRepository) *getByIDRecipeHandler {
	return &getByIDRecipeHandler{repository: repository}
}

func (h *getByIDRecipeHandler) Handle(ctx context.Context, id string) (RecipeReadModel, error) {
	return h.repository.getByID(ctx, id)
}
