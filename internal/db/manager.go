package db

import (
	"database/sql"
	"fmt"
	"path/filepath"
	"sync"

	"github.com/Tomdooo/spajz/internal/config"
	"github.com/Tomdooo/spajz/internal/models"
	_ "modernc.org/sqlite"
)

var databaseManager *DatabaseManager
var once sync.Once

func GetDatabaseManager() *DatabaseManager {
	once.Do(func() {
		databaseManager = &DatabaseManager{
			bucketsMap:          make(DBConnectionMap),
			bucketConfigManager: *config.GetBucketConfigManager(),
		}
	})

	return databaseManager
}

type BucketDatabase struct {
	Database *sql.DB
	Queries  *Queries
}

type DBConnectionMap map[string]*BucketDatabase

type DatabaseManager struct {
	mu                  sync.RWMutex
	bucketsMap          DBConnectionMap
	bucketConfigManager config.BucketConfigManager
}

func (m *DatabaseManager) InitBucketDatabases() error {
	buckets, err := config.GetBucketList()
	if err != nil {
		return fmt.Errorf("getting bucket list: %w", err)
	}
	for _, bucket := range buckets {
		err := m.InitDatabase(bucket)
		if err != nil {
			return fmt.Errorf("initializing bucket database: %w", err)
		}
	}
	return nil
}

func (m *DatabaseManager) InitDatabase(bucket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	bucketConfig := m.bucketsMap[bucket]
	if bucketConfig != nil {
		return models.ErrBucketAlreadyExists
	}

	dbPath := filepath.Join(config.GetBucketDir(bucket), "bucket.db")

	// 1. Otevřeme spojení s databází
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}

	// 2. Nastavíme zlatou SQLite konfiguraci (WAL režim atd.)
	if _, err := database.Exec("PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;"); err != nil {
		database.Close()

		return fmt.Errorf("executing PRAGMA configurations: %w", err)
	}

	// 3. Spustíme vestavěnou migraci schématu
	// Exec bez problému schroustá i soubor s více příkazy oddělenými středníkem
	if _, err := database.Exec(dbSchema); err != nil {
		database.Close()
		return fmt.Errorf("loading database schema: %w", err)
	}

	queries := New(database)

	bucketDb := new(BucketDatabase{
		Database: database,
		Queries:  queries,
	})

	m.bucketsMap[bucket] = bucketDb
	return nil
}

func (m *DatabaseManager) DestroyDatabase(bucket string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.bucketsMap, bucket)
}

func (m *DatabaseManager) getDatabase(bucket string) (*BucketDatabase, error) {
	database := m.bucketsMap[bucket]
	if database == nil {
		return nil, models.ErrBucketNotFound // ?: maybe database not found?
	}
	return database, nil
}

func (m *DatabaseManager) GetDatabase(bucket string) (*BucketDatabase, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getDatabase(bucket)
}
