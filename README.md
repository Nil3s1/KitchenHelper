# KitchenHelper

KitchenHelper is a Go, PostgreSQL, and React monorepo. The first bounded context is `recipes`; its recipe document is stored as JSONB in the `recipes` schema.

## Local development

Start the application stack with Docker Compose:

```sh
docker compose up --build
```

- Web app: <http://localhost:3000>
- Recipes API: <http://localhost:8080/api/v1/recipes>
- MinIO console: <http://localhost:9001> (`kitchenhelper` / `kitchenhelper-local`)

The database migrator runs after PostgreSQL becomes healthy and before the backend starts. Nginx serves the frontend and proxies `/api/` requests to the backend. The backend exposes a Recipes application API independently of HTTP. Its command methods delegate to use-case handlers; simple reads return DTOs through an injected query port without pass-through query handlers. Shared application contracts live in `app_command.go` and `app_query.go`; the context remains one flat Go package.

Recipe image references are stored in the `images` array; MinIO is available as the local S3-compatible storage service, but application uploads and asset retrieval are not yet wired to it.

## Recipes application API

`recipes.NewModule(db)` returns a `*recipes.Module`, not an HTTP handler. It wires PostgreSQL adapters to the application operations without starting a server. The caller opens and closes the database connection.

| Method | Input | Result |
| --- | --- | --- |
| `AddRecipe` | `context.Context`, `AddRecipeCommand` | `Recipe`, `error` |
| `UpdateRecipe` | `context.Context`, `UpdateRecipeCommand` | `Recipe`, `error` |
| `GetRecipe` | `context.Context`, recipe ID | `RecipeReadModel`, `error` |
| `ListRecipes` | `context.Context` | `[]RecipeReadModel`, `error` |
| `HTTPHandler` | none | `http.Handler` |

For example, code within the backend module can call the application directly:

```go
func addSoup(ctx context.Context, db *sql.DB) (recipes.Recipe, error) {
    module := recipes.NewModule(db)
    return module.AddRecipe(ctx, recipes.AddRecipeCommand{
        Name: "Suppe",
        Ingredients: []recipes.IngredientInput{
            {Name: "Salz", Amount: 1, Unit: "g"},
        },
    })
}
```

The snippet uses `context`, `database/sql`, and `kitchenhelper/backend/internal/contexts/recipes`. Direct calls return application errors such as `recipes.ErrValidation` and `recipes.ErrNotFound`, which callers can inspect with `errors.Is`. Successful commands return domain values; reads return DTOs without loading domain aggregates.

HTTP is attached separately:

```go
module := recipes.NewModule(db)
server := &http.Server{
    Addr:    ":8080",
    Handler: module.HTTPHandler(),
}
```

Obtaining the handler does not start the server; the API entrypoint handles startup and shutdown. HTTP decoding and error/status mapping remain in the HTTP adapter. A future MCP adapter can call the same application methods directly without making HTTP requests; no MCP adapter is currently implemented.

The public API remains subject to Go's `internal/` import restrictions and the architecture's prohibition on direct bounded-context imports. It is not a standalone library for unrelated Go modules. PostgreSQL repository implementations and transaction details remain private; the current constructor deliberately selects PostgreSQL, not an arbitrary database.

## Development checks

Run the relevant checks from the repository root using the existing dependencies:

```powershell
Push-Location src\backend
go test ./...
go build ./...
Pop-Location

Push-Location src\database
go test ./...
go build ./...
Pop-Location

Push-Location src\frontend
npm run build
Pop-Location
```

The frontend build checks TypeScript and creates the Vite production bundle; no frontend test script is configured yet. If dependencies are missing, restore the declared dependencies first; new third-party dependencies require explicit approval.

For cross-component changes, start the Compose stack and verify the affected browser/API flow. A successful build alone does not verify end-to-end behavior. Migration integration checks additionally require PostgreSQL and `TEST_DATABASE_URL`, as described below.

## Recipe API

- `POST /api/v1/recipes` — create a recipe (initial version is `1`)
- `GET /api/v1/recipes/{id}` — retrieve one recipe
- `GET /api/v1/recipes` — list recipes
- `PUT /api/v1/recipes/{id}` — replace a recipe document (version increments)

The request document includes `name`, `description`, `author`, `text`, `tags`, `ingredients` (`name`, `amount`, `unit`), and `images`. Recipe and ingredient names and ingredient units must be non-empty; ingredient names must be unique within a recipe.

## Database migration integration test

The migration integration test creates and removes a uniquely named temporary database. Set `TEST_DATABASE_URL` to a PostgreSQL connection URL whose user can create databases, then run:

```sh
cd tests/database
go test ./...
```

The test applies the migrations, rolls them back, verifies the owned table and GIN index, then reapplies them. It skips when `TEST_DATABASE_URL` is unset.
