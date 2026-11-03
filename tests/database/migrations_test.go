package database_test

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

func TestRecipesMigrationApplyRollbackAndReapply(t *testing.T) {
	databaseURL := os.Getenv("TEST_DATABASE_URL")
	if databaseURL == "" {
		t.Skip("set TEST_DATABASE_URL to a PostgreSQL URL whose credentials can create databases")
	}

	adminURL, err := url.Parse(databaseURL)
	if err != nil {
		t.Fatalf("parse TEST_DATABASE_URL: %v", err)
	}
	if adminURL.Scheme != "postgres" && adminURL.Scheme != "postgresql" {
		t.Fatalf("TEST_DATABASE_URL must use postgres or postgresql scheme, got %q", adminURL.Scheme)
	}
	adminURL.Path = "/postgres"

	adminDB, err := sql.Open("postgres", adminURL.String())
	if err != nil {
		t.Fatalf("open PostgreSQL admin connection: %v", err)
	}
	defer adminDB.Close()
	if err := adminDB.Ping(); err != nil {
		t.Fatalf("connect to PostgreSQL admin database: %v", err)
	}

	randomID := make([]byte, 8)
	if _, err := rand.Read(randomID); err != nil {
		t.Fatalf("generate temporary database name: %v", err)
	}
	databaseName := "kitchenhelper_migration_test_" + hex.EncodeToString(randomID)
	if _, err := adminDB.Exec(`CREATE DATABASE "` + databaseName + `"`); err != nil {
		t.Fatalf("create temporary migration database: %v", err)
	}
	t.Cleanup(func() {
		cleanupDB, err := sql.Open("postgres", adminURL.String())
		if err != nil {
			t.Errorf("open PostgreSQL cleanup connection: %v", err)
			return
		}
		defer cleanupDB.Close()
		if _, err := cleanupDB.Exec(`DROP DATABASE "` + databaseName + `"`); err != nil {
			t.Errorf("drop temporary migration database: %v", err)
		}
	})

	testURL := *adminURL
	testURL.Path = "/" + databaseName
	migrationsURL := migrationSourceURL(t)

	runMigrations(t, migrationsURL, testURL.String(), func(m *migrate.Migrate) error {
		return m.Up()
	})
	assertRecipesTableAndIndex(t, testURL.String(), true)

	runMigrations(t, migrationsURL, testURL.String(), func(m *migrate.Migrate) error {
		return m.Down()
	})
	assertRecipesTableAndIndex(t, testURL.String(), false)

	runMigrations(t, migrationsURL, testURL.String(), func(m *migrate.Migrate) error {
		return m.Up()
	})
	assertRecipesTableAndIndex(t, testURL.String(), true)
}

func migrationSourceURL(t *testing.T) string {
	t.Helper()
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve migration test location")
	}
	migrationsPath, err := filepath.Abs(filepath.Join(filepath.Dir(currentFile), "..", "..", "src", "database", "migrations"))
	if err != nil {
		t.Fatalf("resolve migrations directory: %v", err)
	}
	migrationsPath = filepath.ToSlash(migrationsPath)
	if !strings.HasPrefix(migrationsPath, "/") {
		migrationsPath = "/" + migrationsPath
	}
	return (&url.URL{Scheme: "file", Path: migrationsPath}).String()
}

func runMigrations(t *testing.T, sourceURL, databaseURL string, run func(*migrate.Migrate) error) {
	t.Helper()
	migrationRunner, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		t.Fatalf("initialize migration runner: %v", err)
	}
	runErr := run(migrationRunner)
	sourceErr, databaseErr := migrationRunner.Close()
	if runErr != nil && !errors.Is(runErr, migrate.ErrNoChange) {
		t.Fatalf("run migration: %v", runErr)
	}
	if sourceErr != nil {
		t.Errorf("close migration source: %v", sourceErr)
	}
	if databaseErr != nil {
		t.Errorf("close migration database: %v", databaseErr)
	}
}

func assertRecipesTableAndIndex(t *testing.T, databaseURL string, exists bool) {
	t.Helper()
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		t.Fatalf("open migration test database: %v", err)
	}
	defer db.Close()
	if err := db.Ping(); err != nil {
		t.Fatalf("connect to migration test database: %v", err)
	}

	var tableName sql.NullString
	if err := db.QueryRow(`SELECT to_regclass('recipes.recipes')`).Scan(&tableName); err != nil {
		t.Fatalf("query recipes table: %v", err)
	}
	if tableName.Valid != exists {
		t.Fatalf("recipes table exists = %t, want %t", tableName.Valid, exists)
	}
	if !exists {
		var schemaExists bool
		if err := db.QueryRow(`SELECT EXISTS (SELECT 1 FROM information_schema.schemata WHERE schema_name = 'recipes')`).Scan(&schemaExists); err != nil {
			t.Fatalf("query recipes schema: %v", err)
		}
		if !schemaExists {
			t.Fatal("down migration removed the context schema")
		}
		return
	}

	var indexDefinition string
	if err := db.QueryRow(
		`SELECT indexdef FROM pg_indexes WHERE schemaname = 'recipes' AND tablename = 'recipes' AND indexname = 'recipes_payload_gin_idx'`,
	).Scan(&indexDefinition); err != nil {
		t.Fatalf("query recipes GIN index: %v", err)
	}
	if !strings.Contains(strings.ToLower(indexDefinition), "using gin") {
		t.Fatalf("recipes payload index definition = %q, want GIN", indexDefinition)
	}

	expectedTypes := map[string]string{
		"id":         "uuid",
		"created_at": "timestamp with time zone",
		"updated_at": "timestamp with time zone",
		"payload":    "jsonb",
	}
	for column, expectedType := range expectedTypes {
		var actualType string
		if err := db.QueryRow(
			`SELECT data_type FROM information_schema.columns WHERE table_schema = 'recipes' AND table_name = 'recipes' AND column_name = $1`,
			column,
		).Scan(&actualType); err != nil {
			t.Fatalf("query recipes.%s type: %v", column, err)
		}
		if !strings.EqualFold(actualType, expectedType) {
			t.Errorf("recipes.%s type = %s, want %s", column, actualType, expectedType)
		}
	}
}
