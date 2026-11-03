package recipes

import (
	"database/sql"
	"net/http"
)

func NewModule(db *sql.DB) http.Handler {
	commandRepository := newPostgresCommandRepository(db)
	unitOfWork := newPostgresUnitOfWork(db)
	addRecipe := newAddRecipeHandler(commandRepository)
	updateRecipe := newUpdateRecipeHandler(unitOfWork)
	readRepository := newPostgresQueryRepository(db)
	getByID := newGetByIDRecipeHandler(readRepository)
	getAll := newGetAllRecipesHandler(readRepository)
	handler := newHTTPHandler(addRecipe, updateRecipe, getByID, getAll)

	mux := http.NewServeMux()
	handler.registerRoutes(mux)
	return mux
}
