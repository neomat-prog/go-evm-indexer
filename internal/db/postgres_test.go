package db

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
)

func testPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("DB_URL")
	if dsn == "" {
		t.Skip("DB_URL not set, skipping Postgres test")
	}

	pool, err := NewPostgres(context.Background(), dsn)
	if err != nil {
		t.Fatalf("NewPostgres: %v", err)
	}
	t.Cleanup(pool.Close)

	return pool
}

func TestNewPostgresConnects(t *testing.T) {
	testPool(t)
}

func TestNewPostgresFailsOnBadDSN(t *testing.T) {
	_, err := NewPostgres(context.Background(),
		"postgres://nobody@127.0.0.1:1/nope")
	if err == nil {
		t.Fatal("NewPostgres succeeded against an unreachable server error")
	}
}

func TestMigrateCreatesTable(t *testing.T) {
	pool := testPool(t)
	ctx := context.Background()

	if err := Migrate(ctx, pool); err != nil {
		t.Fatalf("Migrate: %v", err)
	}

	var name *string
	if err := pool.QueryRow(ctx, `SELECT 
  to_regclass('public.transfers')::text`).Scan(&name); err != nil {
		t.Fatalf("to_regclass: %v", err)
	}
	if name == nil {
		t.Fatal("transfers table does not exist after Migrate")
	}
}
