package db

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"gorm.io/gorm"
)

var versionedMigrationPattern = regexp.MustCompile(`^[0-9]{4}_.+\.sql$`)

// RunSQLMigrations executes versioned SQL migration files in lexical order.
// Only files matching NNNN_*.sql are applied.
func RunSQLMigrations(db *gorm.DB, migrationsDir string) error {
	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !versionedMigrationPattern.MatchString(name) {
			continue
		}
		path := filepath.Join(migrationsDir, name)
		sqlBytes, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}
		if err := db.Exec(string(sqlBytes)).Error; err != nil {
			return fmt.Errorf("apply migration %s: %w", name, err)
		}
	}

	return nil
}
