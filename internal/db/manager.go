package db

import (
	"database/sql"
	"errors"
	"path/filepath"
	"sync"

	"github.com/Tomdooo/spajz/internal/config"
	_ "modernc.org/sqlite"
)

var (
	ErrConnectionAlreadyExists = errors.New("Connection already exists.")
	ErrBucketNotExist          = errors.New("Bucket does not exist.")
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
		return err
	}
	for _, bucket := range buckets {
		err := m.InitDatabase(bucket)
		if err != nil {
			return err
		}
	}
	return nil
}

func (m *DatabaseManager) InitDatabase(bucket string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	bucketConfig := m.bucketsMap[bucket] // ?: cannot understand now, check
	if bucketConfig != nil {
		return ErrBucketNotExist
	}

	dbPath := filepath.Join(config.GetBucketDir(bucket), "bucket.db")

	// 1. Otevřeme spojení s databází
	database, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return err
	}

	// 2. Nastavíme zlatou SQLite konfiguraci (WAL režim atd.)
	if _, err := database.Exec("PRAGMA journal_mode=WAL; PRAGMA synchronous=NORMAL;"); err != nil {
		database.Close()
		return err
	}

	// 3. Spustíme vestavěnou migraci schématu
	// Exec bez problému schroustá i soubor s více příkazy oddělenými středníkem
	if _, err := database.Exec(dbSchema); err != nil {
		database.Close()
		return err
	}

	queries := New(database)

	bucketDb := new(BucketDatabase{
		Database: database,
		Queries:  queries,
	})

	m.bucketsMap[bucket] = bucketDb
	return nil
}

func (m *DatabaseManager) getDatabase(bucket string) (*BucketDatabase, error) {
	connection := m.bucketsMap[bucket]
	if connection == nil {
		return nil, ErrBucketNotExist
	}
	return connection, nil
}

func (m *DatabaseManager) GetDatabase(bucket string) (*BucketDatabase, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.getDatabase(bucket)
}
