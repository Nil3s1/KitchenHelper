package recipes

import "context"

type IngredientInput struct {
	Name   string
	Amount float64
	Unit   string
}

type recipeCommandRepository interface {
	create(context.Context, Recipe) error
	getByID(context.Context, string) (Recipe, error)
	update(context.Context, Recipe) error
}

type unitOfWork interface {
	Do(context.Context, func(context.Context, recipeCommandRepository) error) error
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
