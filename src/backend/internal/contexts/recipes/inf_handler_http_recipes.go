package recipes

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"time"
)

type httpHandler struct {
	application *Module
}

func newHTTPHandler(application *Module) *httpHandler {
	return &httpHandler{application: application}
}

// HTTPHandler exposes the application API through HTTP without starting a server.
func (m *Module) HTTPHandler() http.Handler {
	mux := http.NewServeMux()
	newHTTPHandler(m).registerRoutes(mux)
	return mux
}

func (h *httpHandler) registerRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/recipes", h.add)
	mux.HandleFunc("GET /api/v1/recipes", h.getAll)
	mux.HandleFunc("GET /api/v1/recipes/{id}", h.getByID)
	mux.HandleFunc("PUT /api/v1/recipes/{id}", h.update)
}

func (h *httpHandler) add(w http.ResponseWriter, r *http.Request) {
	request, err := decodeRecipeRequest(w, r)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			writeCommandError(w, err)
			return
		}
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	created, err := h.application.AddRecipe(r.Context(), addCommandFromRequest(request))
	if err != nil {
		writeCommandError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, recipeResponseFromDomain(created))
}

func (h *httpHandler) getByID(w http.ResponseWriter, r *http.Request) {
	recipe, err := h.application.GetRecipe(r.Context(), r.PathValue("id"))
	if err != nil {
		writeQueryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recipe)
}

func (h *httpHandler) getAll(w http.ResponseWriter, r *http.Request) {
	recipes, err := h.application.ListRecipes(r.Context())
	if err != nil {
		writeQueryError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recipes)
}

func (h *httpHandler) update(w http.ResponseWriter, r *http.Request) {
	request, err := decodeRecipeRequest(w, r)
	if err != nil {
		if errors.Is(err, ErrValidation) {
			writeCommandError(w, err)
			return
		}
		writeError(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	updated, err := h.application.UpdateRecipe(r.Context(), updateCommandFromRequest(r.PathValue("id"), request))
	if err != nil {
		writeCommandError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, recipeResponseFromDomain(updated))
}

type recipeRequest struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Author      string              `json:"author"`
	Version     int                 `json:"version"`
	Text        string              `json:"text"`
	Tags        []string            `json:"tags"`
	Ingredients []ingredientRequest `json:"ingredients"`
	Images      []string            `json:"images"`
}

type ingredientRequest struct {
	Name   string   `json:"name"`
	Amount *float64 `json:"amount"`
	Unit   string   `json:"unit"`
}

type recipeResponse struct {
	ID          string               `json:"id"`
	CreatedAt   time.Time            `json:"created_at"`
	UpdatedAt   time.Time            `json:"updated_at"`
	Name        string               `json:"name"`
	Description string               `json:"description"`
	Author      string               `json:"author"`
	Version     int                  `json:"version"`
	Text        string               `json:"text"`
	Tags        []string             `json:"tags"`
	Ingredients []ingredientResponse `json:"ingredients"`
	Images      []string             `json:"images"`
}

type ingredientResponse struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}

func decodeRecipeRequest(w http.ResponseWriter, r *http.Request) (recipeRequest, error) {
	var request recipeRequest
	if err := decodeBody(w, r, &request); err != nil {
		return recipeRequest{}, err
	}
	for _, ingredient := range request.Ingredients {
		if ingredient.Amount == nil {
			return recipeRequest{}, ErrValidation
		}
	}
	return request, nil
}

func addCommandFromRequest(request recipeRequest) AddRecipeCommand {
	return AddRecipeCommand{
		Name:        request.Name,
		Description: request.Description,
		Author:      request.Author,
		Text:        request.Text,
		Tags:        request.Tags,
		Ingredients: ingredientInputsFromRequest(request.Ingredients),
		Images:      request.Images,
	}
}

func updateCommandFromRequest(id string, request recipeRequest) UpdateRecipeCommand {
	return UpdateRecipeCommand{
		ID:          id,
		Name:        request.Name,
		Description: request.Description,
		Author:      request.Author,
		Text:        request.Text,
		Tags:        request.Tags,
		Ingredients: ingredientInputsFromRequest(request.Ingredients),
		Images:      request.Images,
	}
}

func ingredientInputsFromRequest(requests []ingredientRequest) []IngredientInput {
	ingredients := make([]IngredientInput, 0, len(requests))
	for _, request := range requests {
		ingredients = append(ingredients, IngredientInput{
			Name:   request.Name,
			Amount: *request.Amount,
			Unit:   request.Unit,
		})
	}
	return ingredients
}

func recipeResponseFromDomain(recipe Recipe) recipeResponse {
	ingredients := make([]ingredientResponse, 0, len(recipe.Ingredients))
	for _, ingredient := range recipe.Ingredients {
		ingredients = append(ingredients, ingredientResponse{
			Name:   ingredient.Name,
			Amount: ingredient.Amount,
			Unit:   ingredient.Unit,
		})
	}
	return recipeResponse{
		ID:          recipe.ID,
		CreatedAt:   recipe.CreatedAt,
		UpdatedAt:   recipe.UpdatedAt,
		Name:        recipe.Name,
		Description: recipe.Description,
		Author:      recipe.Author,
		Version:     recipe.Version,
		Text:        recipe.Text,
		Tags:        append([]string{}, recipe.Tags...),
		Ingredients: ingredients,
		Images:      append([]string{}, recipe.Images...),
	}
}

func decodeBody(w http.ResponseWriter, r *http.Request, destination any) error {
	decoder := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(destination); err != nil {
		return err
	}
	if err := decoder.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON value")
	}
	return nil
}

func writeCommandError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrValidation) {
		writeError(w, http.StatusBadRequest, "recipe document violates validation rules")
		return
	}
	writeQueryError(w, err)
}

func writeQueryError(w http.ResponseWriter, err error) {
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "recipe not found")
		return
	}
	log.Printf("handle recipe request: %v", err)
	writeError(w, http.StatusInternalServerError, "internal server error")
}

func writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("encode HTTP response: %v", err)
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
