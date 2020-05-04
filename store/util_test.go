package store

import (
	"database/sql"
	"fmt"
	"os"
	"sync/atomic"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/stretchr/testify/require"

	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

var dbID int32

func DBTest(t *testing.T, cb func(t *testing.T, db *sql.DB)) {
	nextID := atomic.AddInt32(&dbID, 1)
	dbName := fmt.Sprintf("ddrp_relayer_test_%d", nextID)
	migrationsDir := os.Getenv("MIGRATIONS_DIR")
	require.NotEmpty(t, migrationsDir, "migrations directory must be set")
	mgrDB, err := sql.Open("postgres", "postgres://localhost:5432/?sslmode=disable")
	require.NoError(t, err)
	_, err = mgrDB.Exec(fmt.Sprintf("DROP DATABASE IF EXISTS %s", dbName))
	require.NoError(t, err)
	_, err = mgrDB.Exec(fmt.Sprintf("CREATE DATABASE %s", dbName))
	require.NoError(t, err)
	db, err := sql.Open("postgres", fmt.Sprintf("postgres://localhost:5432/%s?sslmode=disable", dbName))
	require.NoError(t, err)
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	require.NoError(t, err)
	m, err := migrate.NewWithDatabaseInstance(fmt.Sprintf("file://%s", migrationsDir), "postgres", driver)
	require.NoError(t, err)
	require.NoError(t, m.Up())
	cb(t, db)
	require.NoError(t, db.Close())
	_, err = mgrDB.Exec(
		"SELECT pg_terminate_backend(pg_stat_activity.pid) FROM pg_stat_activity WHERE datname=$1 AND pid<>pg_backend_pid()",
		dbName,
	)
	require.NoError(t, err)
	_, err = mgrDB.Exec(fmt.Sprintf("DROP DATABASE %s", dbName))
	require.NoError(t, err)
	require.NoError(t, mgrDB.Close())
}
