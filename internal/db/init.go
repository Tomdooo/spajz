package db

import (
	"database/sql"
	_ "embed"
)

//go:embed schema.sql
var dbSchema string

// InitDatabase otevře SQLite spojení a rovnou se ujistí, že tabulky existují
func InitDatabase(dbPath string) (*sql.DB, error) {
	// 1. Otevřeme spojení s databází
	database, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	// 2. Nastavíme zlatou SQLite konfiguraci (WAL režim atd.)
	if _, err := database.Exec("PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;"); err != nil {
		database.Close()
		return nil, err
	}

	// 3. Spustíme vestavěnou migraci schématu
	// Exec bez problému schroustá i soubor s více příkazy oddělenými středníkem
	if _, err := database.Exec(dbSchema); err != nil {
		database.Close()
		return nil, err
	}

	return database, nil
}
