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

The database migrator runs after PostgreSQL becomes healthy and before the backend starts. Nginx serves the frontend and proxies `/api/` requests to the backend. The backend applies Recipes use cases and reads/writes PostgreSQL through context-owned adapters.

Recipe image references are stored in the `images` array; MinIO is available as the local S3-compatible storage service, but application uploads and asset retrieval are not yet wired to it.

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
