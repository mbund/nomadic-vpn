package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

var DB *sql.DB

func CreateDb(dbPath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create parent directories for the database: %s", err)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the database: %s", err)
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS keys (
			id INTEGER PRIMARY KEY CHECK (id = 0),
			vultr VARCHAR(64)
		);`,
		`INSERT OR IGNORE INTO keys (id, vultr) VALUES (0, NULL);`,
	}

	for _, stmt := range stmts {
		_, err := db.Exec(stmt)
		if err != nil {
			return nil, err
		}
	}

	return db, nil
}

// type Store struct {
// 	Db *sql.DB
// }

// func NewStore(dbName string) (Store, error) {
// 	Db, err := getConnection(dbName)
// 	if err != nil {
// 		return Store{}, err
// 	}

// 	if err := createMigrations(dbName, Db); err != nil {
// 		return Store{}, err
// 	}

// 	return Store{
// 		Db,
// 	}, nil
// }

// func getConnection(dbName string) (*sql.DB, error) {
// 	var (
// 		err error
// 		db  *sql.DB
// 	)

// 	if db != nil {
// 		return db, nil
// 	}

// 	db, err = sql.Open("sqlite", dbName)
// 	if err != nil {
// 		// log.Fatalf("🔥 failed to connect to the database: %s", err.Error())
// 		return nil, fmt.Errorf("🔥 failed to connect to the database: %s", err)
// 	}

// 	log.Println("🚀 Connected Successfully to the Database")

// 	return db, nil
// }

// func createMigrations(dbName string, db *sql.DB) error {
// 	stmts := []string{
// 		`CREATE TABLE IF NOT EXISTS keys (
// 			id INTEGER PRIMARY KEY CHECK (id = 0),
// 			vultr VARCHAR(64)
// 		);`,
// 		`INSERT OR IGNORE INTO keys (id, vultr) VALUES (0, NULL);`,
// 	}

// 	for _, stmt := range stmts {
// 		_, err := db.Exec(stmt)
// 		if err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
