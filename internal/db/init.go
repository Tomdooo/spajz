package db

import (
	"database/sql"
	_ "embed"
	"fmt"
)

//go:embed schema.sql
var dbSchema string

// InitDatabase otevře SQLite spojení a rovnou se ujistí, že tabulky existují
func InitDatabase(dbPath string) (*sql.DB, error) {
	// 1. Otevřeme spojení s databází
	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	// 2. Nastavíme zlatou SQLite konfiguraci (WAL režim atd.)
	if _, err := database.Exec("PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;"); err != nil {
		database.Close()
		return nil, fmt.Errorf("executing PRAGMA configurations: %w", err)
	}

	// 3. Spustíme vestavěnou migraci schématu
	// Exec bez problému schroustá i soubor s více příkazy oddělenými středníkem
	if _, err := database.Exec(dbSchema); err != nil {
		database.Close()
		return nil, fmt.Errorf("loading database schema: %w", err)
	}

	return database, nil
}
