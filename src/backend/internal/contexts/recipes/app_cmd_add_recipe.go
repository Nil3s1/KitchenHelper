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
	unitOfWork unitOfWork
}

func newAddRecipeHandler(unitOfWork unitOfWork) *AddRecipeHandler {
	return &AddRecipeHandler{unitOfWork: unitOfWork}
}

func (h *AddRecipeHandler) Handle(ctx context.Context, command AddRecipeCommand) (Recipe, error) {
	var created Recipe
	err := h.unitOfWork.Do(ctx, func(ctx context.Context, repositories recipeRepositories) error {
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
			return err
		}
		if err := repositories.create(ctx, recipe); err != nil {
			return err
		}
		created = recipe
		return nil
	})
	if err != nil {
		return Recipe{}, err
	}
	return created, nil
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
