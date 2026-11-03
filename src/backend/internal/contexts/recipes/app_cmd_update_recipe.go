package recipes

import "context"

type UpdateRecipeCommand struct {
	ID          string
	Name        string
	Description string
	Author      string
	Text        string
	Tags        []string
	Ingredients []IngredientInput
	Images      []string
}

type UpdateRecipeHandler struct {
	unitOfWork unitOfWork
}

func newUpdateRecipeHandler(unitOfWork unitOfWork) *UpdateRecipeHandler {
	return &UpdateRecipeHandler{unitOfWork: unitOfWork}
}

func (h *UpdateRecipeHandler) Handle(ctx context.Context, command UpdateRecipeCommand) (Recipe, error) {
	var updated Recipe
	err := h.unitOfWork.Do(ctx, func(ctx context.Context, repositories recipeRepositories) error {
		recipe, err := repositories.getByID(ctx, command.ID)
		if err != nil {
			return err
		}
		if err := recipe.replace(
			command.Name,
			command.Description,
			command.Author,
			command.Text,
			command.Tags,
			ingredientsFromInput(command.Ingredients),
			command.Images,
		); err != nil {
			return err
		}
		if err := repositories.update(ctx, recipe); err != nil {
			return err
		}
		updated = recipe
		return nil
	})
	if err != nil {
		return Recipe{}, err
	}
	return updated, nil
}
