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

type updateRecipeHandler struct {
	unitOfWork unitOfWork
}

func newUpdateRecipeHandler(unitOfWork unitOfWork) *updateRecipeHandler {
	return &updateRecipeHandler{unitOfWork: unitOfWork}
}

func (h *updateRecipeHandler) Handle(ctx context.Context, command UpdateRecipeCommand) (Recipe, error) {
	var updated Recipe
	err := h.unitOfWork.Do(ctx, func(ctx context.Context, repository recipeCommandRepository) error {
		recipe, err := repository.getByID(ctx, command.ID)
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
		if err := repository.update(ctx, recipe); err != nil {
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
