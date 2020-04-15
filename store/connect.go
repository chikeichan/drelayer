package store

import (
	"database/sql"
	"ddrp-relayer/config"
	"ddrp-relayer/log"
	"fmt"
)

func Connect(cfg *config.Database) (*sql.DB, error) {
	connStr := fmt.Sprintf(
		"user=%s password='%s' dbname=%s sslmode=%s host=%s port=%d",
		cfg.Username,
		cfg.Password,
		cfg.Name,
		cfg.SSLMode,
		cfg.Host,
		cfg.Port,
	)
	if cfg.SSLRootCert != "" {
		connStr += fmt.Sprintf(" sslrootcert=%s", cfg.SSLRootCert)
	}

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}
	log.WithModule("store").Info(
		"connected to database",
		"name", cfg.Name,
		"host", cfg.Host,
		"port", cfg.Port,
	)
	return db, nil
}
