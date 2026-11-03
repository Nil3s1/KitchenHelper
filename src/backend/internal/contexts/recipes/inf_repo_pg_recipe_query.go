package recipes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type postgresQueryRepository struct {
	db *sql.DB
}

func newPostgresQueryRepository(db *sql.DB) *postgresQueryRepository {
	return &postgresQueryRepository{db: db}
}

func (r *postgresQueryRepository) getByID(ctx context.Context, id string) (RecipeReadModel, error) {
	var readModel RecipeReadModel
	var payload []byte
	err := r.db.QueryRowContext(
		ctx,
		`SELECT id, created_at, updated_at, payload FROM recipes.recipes WHERE id = $1`,
		id,
	).Scan(&readModel.ID, &readModel.CreatedAt, &readModel.UpdatedAt, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return RecipeReadModel{}, ErrNotFound
	}
	if err != nil {
		return RecipeReadModel{}, fmt.Errorf("get recipe query projection: %w", err)
	}
	return readModelFromPersistence(readModel.ID, readModel.CreatedAt, readModel.UpdatedAt, payload)
}

func (r *postgresQueryRepository) getAll(ctx context.Context) (GetAllRecipesReadModel, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT id, created_at, updated_at, payload FROM recipes.recipes ORDER BY created_at DESC, id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list recipe query projections: %w", err)
	}
	defer rows.Close()

	readModels := make(GetAllRecipesReadModel, 0)
	for rows.Next() {
		var readModel RecipeReadModel
		var payload []byte
		if err := rows.Scan(&readModel.ID, &readModel.CreatedAt, &readModel.UpdatedAt, &payload); err != nil {
			return nil, fmt.Errorf("scan recipe query projection: %w", err)
		}
		readModel, err = readModelFromPersistence(readModel.ID, readModel.CreatedAt, readModel.UpdatedAt, payload)
		if err != nil {
			return nil, err
		}
		readModels = append(readModels, readModel)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recipe query projections: %w", err)
	}
	return readModels, nil
}

func readModelFromPersistence(id string, createdAt, updatedAt time.Time, payload []byte) (RecipeReadModel, error) {
	var readModel RecipeReadModel
	if err := json.Unmarshal(payload, &readModel); err != nil {
		return RecipeReadModel{}, fmt.Errorf("decode recipe query projection: %w", err)
	}
	readModel.ID = id
	readModel.CreatedAt = createdAt
	readModel.UpdatedAt = updatedAt
	if readModel.Tags == nil {
		readModel.Tags = []string{}
	}
	if readModel.Ingredients == nil {
		readModel.Ingredients = []IngredientReadModel{}
	}
	if readModel.Images == nil {
		readModel.Images = []string{}
	}
	return readModel, nil
}
