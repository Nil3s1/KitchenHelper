package recipes

import (
	"database/sql"
	"net/http"
)

func NewModule(db *sql.DB) http.Handler {
	unitOfWork := newPostgresUnitOfWork(db)
	addRecipe := newAddRecipeHandler(unitOfWork)
	updateRecipe := newUpdateRecipeHandler(unitOfWork)
	readRepository := newPostgresRepository(db)
	getByID := newGetByIDRecipeHandler(readRepository)
	getAll := newGetAllRecipesHandler(readRepository)
	handler := newHTTPHandler(addRecipe, updateRecipe, getByID, getAll)

	mux := http.NewServeMux()
	handler.registerRoutes(mux)
	return mux
}
