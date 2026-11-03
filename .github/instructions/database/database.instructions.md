---
description: "PostgreSQL schema and migration guidance for the repository's Go-based migration service."
applyTo: "src/database/**, tests/database/**"
---

# Database and Migration Instructions

The repository currently targets PostgreSQL and runs migrations with `golang-migrate`. These are technology-specific instructions; general ownership and bounded-context principles are defined in `../architecture/architecture.instructions.md`.

## 1. Schema Ownership

- Keep each bounded context's tables in its own PostgreSQL schema.
- Keep migration files directly under `src/database/migrations/` in one ordered, flat sequence.
- Use sequential migration identifiers followed by a concise context and change description, for example `000002_recipes_add_ingredients.up.sql`.
- Use a matching `.up.sql` and `.down.sql` file for each reversible migration.
- Use JSONB, relational columns, and PostgreSQL extensions only when justified by the context's query and consistency needs. Do not add fuzzy-search extensions or indexes speculatively.

## 2. Migration Behavior

- Treat each migration as a versioned change applied by the migration runner, not as a script that must be safely rerun by hand.
- Do not add `IF NOT EXISTS` or `IF EXISTS` everywhere to simulate idempotency; these can hide schema drift or unexpected starting states. Use guards only when the migration's intended behavior requires them.
- Follow the actual migration tool's transaction behavior. Do not add explicit `BEGIN`/`COMMIT` wrappers unless they are supported and appropriate for that migration and runner.
- Make each down migration reverse only the objects and changes owned by its paired up migration.
- Do not drop a whole context schema when later migrations or independently owned objects may exist in it. Prefer dropping the specific objects introduced by the migration.
- If a change cannot be safely reversed, document that limitation and its operational impact rather than providing a misleading rollback.
- Preserve migration ordering and do not edit an already-applied migration to change deployed behavior; add a new migration instead.

## 3. Validation and Tests

- Validate new migrations against the supported PostgreSQL version.
- Test both applying migrations from an empty database and rolling back/reapplying where the migration and tool support it.
- Put migration integration tests under `tests/database/`; do not rely solely on SQL inspection or unit tests for database behavior.
- Keep database credentials and environment-specific connection settings outside committed migration files.

## 4. Portability

SQLite is a future consideration, not a requirement for current PostgreSQL migrations. Avoid claiming portability unless the schema and every relevant query have been tested against the additional database.
