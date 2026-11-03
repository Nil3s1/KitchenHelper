package recipes

import "context"

type recipeReadRepository interface {
	getByID(context.Context, string) (Recipe, error)
	getAll(context.Context) ([]Recipe, error)
}

type recipeWriteRepository interface {
	create(context.Context, Recipe) error
	update(context.Context, Recipe) error
}

type recipeRepositories interface {
	recipeReadRepository
	recipeWriteRepository
}

type unitOfWork interface {
	Do(context.Context, func(context.Context, recipeRepositories) error) error
}
