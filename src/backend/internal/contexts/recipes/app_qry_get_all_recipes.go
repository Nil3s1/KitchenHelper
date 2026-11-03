package recipes

import "context"

type GetAllRecipesQuery struct{}

type GetAllRecipesReadModel []RecipeReadModel

type GetAllRecipesHandler struct {
	repository recipeReadRepository
}

func newGetAllRecipesHandler(repository recipeReadRepository) *GetAllRecipesHandler {
	return &GetAllRecipesHandler{repository: repository}
}

func (h *GetAllRecipesHandler) Handle(ctx context.Context, _ GetAllRecipesQuery) (GetAllRecipesReadModel, error) {
	recipes, err := h.repository.getAll(ctx)
	if err != nil {
		return nil, err
	}
	result := make(GetAllRecipesReadModel, 0, len(recipes))
	for _, recipe := range recipes {
		result = append(result, newRecipeReadModel(recipe))
	}
	return result, nil
}
