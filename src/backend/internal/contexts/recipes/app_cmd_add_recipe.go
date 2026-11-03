package recipes

import "context"

type AddRecipeCommand struct {
	Name        string
	Description string
	Author      string
	Text        string
	Tags        []string
	Ingredients []IngredientInput
	Images      []string
}

type addRecipeHandler struct {
	repository recipeCommandRepository
}

func newAddRecipeHandler(repository recipeCommandRepository) *addRecipeHandler {
	return &addRecipeHandler{repository: repository}
}

func (h *addRecipeHandler) Handle(ctx context.Context, command AddRecipeCommand) (Recipe, error) {
	recipe, err := newRecipe(
		command.Name,
		command.Description,
		command.Author,
		command.Text,
		command.Tags,
		ingredientsFromInput(command.Ingredients),
		command.Images,
	)
	if err != nil {
		return Recipe{}, err
	}
	if err := h.repository.create(ctx, recipe); err != nil {
		return Recipe{}, err
	}
	return recipe, nil
}
