package recipes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"kitchenhelper/backend/internal/shared"
)

type sqlExecutor interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

type postgresRepository struct {
	executor  sqlExecutor
	lockReads bool
}

type postgresUnitOfWork struct {
	db *sql.DB
}

type recipePayload struct {
	Name        string              `json:"name"`
	Description string              `json:"description"`
	Author      string              `json:"author"`
	Version     int                 `json:"version"`
	Text        string              `json:"text"`
	Tags        []string            `json:"tags"`
	Ingredients []ingredientPayload `json:"ingredients"`
	Images      []string            `json:"images"`
}

type ingredientPayload struct {
	Name   string  `json:"name"`
	Amount float64 `json:"amount"`
	Unit   string  `json:"unit"`
}

type recipePersistenceModel struct {
	ID        string        `db:"id"`
	CreatedAt time.Time     `db:"created_at"`
	UpdatedAt time.Time     `db:"updated_at"`
	Payload   recipePayload `db:"payload"`
}

func newPostgresRepository(db *sql.DB) *postgresRepository {
	return &postgresRepository{executor: db}
}

func newPostgresUnitOfWork(db *sql.DB) *postgresUnitOfWork {
	return &postgresUnitOfWork{db: db}
}

func (u *postgresUnitOfWork) Do(
	ctx context.Context,
	work func(context.Context, recipeRepositories) error,
) (err error) {
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin recipe transaction: %w", err)
	}
	committed := false
	defer func() {
		if committed {
			return
		}
		if rollbackErr := tx.Rollback(); rollbackErr != nil && !errors.Is(rollbackErr, sql.ErrTxDone) {
			err = errors.Join(err, fmt.Errorf("rollback recipe transaction: %w", rollbackErr))
		}
	}()

	repositories := &postgresRepository{executor: tx, lockReads: true}
	if err = work(ctx, repositories); err != nil {
		return err
	}
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("commit recipe transaction: %w", err)
	}
	committed = true
	return nil
}

func (r *postgresRepository) create(ctx context.Context, recipe Recipe) error {
	model := persistenceFromDomain(recipe)
	payload, err := json.Marshal(model.Payload)
	if err != nil {
		return fmt.Errorf("encode recipe payload: %w", err)
	}
	_, err = r.executor.ExecContext(
		ctx,
		`INSERT INTO recipes.recipes (id, created_at, updated_at, payload) VALUES ($1, $2, $3, $4::jsonb)`,
		model.ID,
		model.CreatedAt,
		model.UpdatedAt,
		string(payload),
	)
	if err != nil {
		return fmt.Errorf("insert recipe: %w", err)
	}
	return nil
}

func (r *postgresRepository) update(ctx context.Context, recipe Recipe) error {
	model := persistenceFromDomain(recipe)
	payload, err := json.Marshal(model.Payload)
	if err != nil {
		return fmt.Errorf("encode recipe payload: %w", err)
	}
	result, err := r.executor.ExecContext(
		ctx,
		`UPDATE recipes.recipes SET payload = $2::jsonb, updated_at = $3 WHERE id = $1`,
		model.ID,
		string(payload),
		model.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("update recipe: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("check updated recipe: %w", err)
	}
	if affected == 0 {
		return shared.ErrNotFound
	}
	return nil
}

func (r *postgresRepository) getByID(ctx context.Context, id string) (Recipe, error) {
	var model recipePersistenceModel
	var payload []byte
	query := `SELECT id, created_at, updated_at, payload FROM recipes.recipes WHERE id = $1`
	if r.lockReads {
		query += ` FOR UPDATE`
	}
	err := r.executor.QueryRowContext(
		ctx,
		query,
		id,
	).Scan(&model.ID, &model.CreatedAt, &model.UpdatedAt, &payload)
	if errors.Is(err, sql.ErrNoRows) {
		return Recipe{}, shared.ErrNotFound
	}
	if err != nil {
		return Recipe{}, fmt.Errorf("get recipe: %w", err)
	}
	if err := json.Unmarshal(payload, &model.Payload); err != nil {
		return Recipe{}, fmt.Errorf("decode recipe payload: %w", err)
	}
	return domainFromPersistence(model), nil
}

func (r *postgresRepository) getAll(ctx context.Context) ([]Recipe, error) {
	rows, err := r.executor.QueryContext(
		ctx,
		`SELECT id, created_at, updated_at, payload FROM recipes.recipes ORDER BY created_at DESC, id`,
	)
	if err != nil {
		return nil, fmt.Errorf("list recipes: %w", err)
	}
	defer rows.Close()

	recipes := make([]Recipe, 0)
	for rows.Next() {
		var model recipePersistenceModel
		var payload []byte
		if err := rows.Scan(&model.ID, &model.CreatedAt, &model.UpdatedAt, &payload); err != nil {
			return nil, fmt.Errorf("scan recipe: %w", err)
		}
		if err := json.Unmarshal(payload, &model.Payload); err != nil {
			return nil, fmt.Errorf("decode recipe payload: %w", err)
		}
		recipes = append(recipes, domainFromPersistence(model))
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recipes: %w", err)
	}
	return recipes, nil
}

func persistenceFromDomain(recipe Recipe) recipePersistenceModel {
	ingredients := make([]ingredientPayload, 0, len(recipe.Ingredients))
	for _, ingredient := range recipe.Ingredients {
		ingredients = append(ingredients, ingredientPayload{
			Name:   ingredient.Name,
			Amount: ingredient.Amount,
			Unit:   ingredient.Unit,
		})
	}
	return recipePersistenceModel{
		ID:        recipe.ID,
		CreatedAt: recipe.CreatedAt,
		UpdatedAt: recipe.UpdatedAt,
		Payload: recipePayload{
			Name:        recipe.Name,
			Description: recipe.Description,
			Author:      recipe.Author,
			Version:     recipe.Version,
			Text:        recipe.Text,
			Tags:        append([]string{}, recipe.Tags...),
			Ingredients: ingredients,
			Images:      append([]string{}, recipe.Images...),
		},
	}
}

func domainFromPersistence(model recipePersistenceModel) Recipe {
	ingredients := make([]Ingredient, 0, len(model.Payload.Ingredients))
	for _, ingredient := range model.Payload.Ingredients {
		ingredients = append(ingredients, Ingredient{
			Name:   ingredient.Name,
			Amount: ingredient.Amount,
			Unit:   ingredient.Unit,
		})
	}
	return Recipe{
		DomainModel: shared.DomainModel{
			ID:        model.ID,
			CreatedAt: model.CreatedAt,
			UpdatedAt: model.UpdatedAt,
		},
		Name:        model.Payload.Name,
		Description: model.Payload.Description,
		Author:      model.Payload.Author,
		Version:     model.Payload.Version,
		Text:        model.Payload.Text,
		Tags:        append([]string{}, model.Payload.Tags...),
		Ingredients: ingredients,
		Images:      append([]string{}, model.Payload.Images...),
	}
}
