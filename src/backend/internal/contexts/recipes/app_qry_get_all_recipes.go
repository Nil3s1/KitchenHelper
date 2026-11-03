package recipes

import "context"

type GetAllRecipesQuery struct{}

type GetAllRecipesReadModel []RecipeReadModel

type getAllRecipesRepository interface {
	getAll(context.Context) (GetAllRecipesReadModel, error)
}

type GetAllRecipesHandler struct {
	repository getAllRecipesRepository
}

func newGetAllRecipesHandler(repository getAllRecipesRepository) *GetAllRecipesHandler {
	return &GetAllRecipesHandler{repository: repository}
}

func (h *GetAllRecipesHandler) Handle(ctx context.Context, _ GetAllRecipesQuery) (GetAllRecipesReadModel, error) {
	return h.repository.getAll(ctx)
}
