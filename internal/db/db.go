package db

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Connect membuka koneksi GORM ke Postgres.
// Driver postgres GORM berjalan di atas pgx v5, jadi DATABASE_URL
// format postgres://... bisa dipakai langsung tanpa diubah.
func Connect(url string) (*gorm.DB, error) {
	return gorm.Open(postgres.Open(url), &gorm.Config{})
}

// Migrate menerapkan file *.sql di dir sesuai urutan nama,
// dicatat di schema_migrations. Ini pendekatan hibrida yang disengaja:
// skema tetap SQL eksplisit (reviewable, versioned), GORM hanya
// dipakai untuk query aplikasi — bukan untuk membuat tabel.
func Migrate(db *gorm.DB, dir string) error {
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS schema_migrations (name TEXT PRIMARY KEY)`).Error; err != nil {
		return fmt.Errorf("migrations table: %w", err)
	}

	files, err := filepath.Glob(filepath.Join(dir, "*.sql"))
	if err != nil {
		return err
	}
	sort.Strings(files)

	for _, f := range files {
		name := filepath.Base(f)

		var count int64
		if err := db.Raw(`SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name).Scan(&count).Error; err != nil {
			return fmt.Errorf("check %s: %w", name, err)
		}
		if count > 0 {
			continue
		}

		raw, err := os.ReadFile(f)
		if err != nil {
			return err
		}
		// database/sql mengeksekusi satu statement per Exec, jadi file
		// dipecah per ';'. Aman selama migrasi tidak mengandung ';'
		// di dalam string literal / function body.
		for _, stmt := range splitStatements(string(raw)) {
			if err := db.Exec(stmt).Error; err != nil {
				return fmt.Errorf("apply %s: %w", name, err)
			}
		}

		if err := db.Exec(`INSERT INTO schema_migrations(name) VALUES(?)`, name).Error; err != nil {
			return fmt.Errorf("record %s: %w", name, err)
		}
	}
	return nil
}

func splitStatements(sql string) []string {
	var out []string
	for _, part := range strings.Split(sql, ";") {
		if s := strings.TrimSpace(part); s != "" {
			out = append(out, s)
		}
	}
	return out
}
