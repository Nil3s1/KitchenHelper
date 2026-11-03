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

type IngredientInput struct {
	Name   string
	Amount float64
	Unit   string
}

type AddRecipeHandler struct {
	repository recipeCommandRepository
}

func newAddRecipeHandler(repository recipeCommandRepository) *AddRecipeHandler {
	return &AddRecipeHandler{repository: repository}
}

func (h *AddRecipeHandler) Handle(ctx context.Context, command AddRecipeCommand) (Recipe, error) {
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

func ingredientsFromInput(inputs []IngredientInput) []Ingredient {
	ingredients := make([]Ingredient, 0, len(inputs))
	for _, input := range inputs {
		ingredients = append(ingredients, Ingredient{
			Name:   input.Name,
			Amount: input.Amount,
			Unit:   input.Unit,
		})
	}
	return ingredients
}
