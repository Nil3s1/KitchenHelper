package recipes

import "context"

type recipeCreateRepository interface {
	create(context.Context, Recipe) error
}

type recipeUpdateRepository interface {
	getByID(context.Context, string) (Recipe, error)
	update(context.Context, Recipe) error
}

type unitOfWork interface {
	Do(context.Context, func(context.Context, recipeUpdateRepository) error) error
}
