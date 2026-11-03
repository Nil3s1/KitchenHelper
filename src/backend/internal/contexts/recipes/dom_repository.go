package recipes

import "context"

type recipeCommandRepository interface {
	create(context.Context, Recipe) error
	getByID(context.Context, string) (Recipe, error)
	update(context.Context, Recipe) error
}

type unitOfWork interface {
	Do(context.Context, func(context.Context, recipeCommandRepository) error) error
}
