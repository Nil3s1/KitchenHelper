package recipes

import (
	"context"
	"database/sql"
)

type Module struct {
	addRecipe    *addRecipeHandler
	updateRecipe *updateRecipeHandler
	queries      recipeQueryRepository
}

// NewModule wires the application API to PostgreSQL without starting a transport.
// The caller owns the database connection and its lifetime.
func NewModule(db *sql.DB) *Module {
	commandRepository := newPostgresCommandRepository(db)
	unitOfWork := newPostgresUnitOfWork(db)
	return &Module{
		addRecipe:    newAddRecipeHandler(commandRepository),
		updateRecipe: newUpdateRecipeHandler(unitOfWork),
		queries:      newPostgresQueryRepository(db),
	}
}

func (m *Module) AddRecipe(ctx context.Context, command AddRecipeCommand) (Recipe, error) {
	return m.addRecipe.Handle(ctx, command)
}

func (m *Module) UpdateRecipe(ctx context.Context, command UpdateRecipeCommand) (Recipe, error) {
	return m.updateRecipe.Handle(ctx, command)
}

func (m *Module) GetRecipe(ctx context.Context, id string) (RecipeReadModel, error) {
	return m.queries.getByID(ctx, id)
}

func (m *Module) ListRecipes(ctx context.Context) ([]RecipeReadModel, error) {
	return m.queries.getAll(ctx)
}
