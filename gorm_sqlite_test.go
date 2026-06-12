package vbasedata

import (
	"io"
	"log/slog"
	"path/filepath"
	"testing"
	"time"
)

type gormSQLiteTestUser struct {
	ID        uint `gorm:"primaryKey"`
	Name      string
	Email     string `gorm:"uniqueIndex"`
	CreatedAt time.Time
}

func TestNewGorm_SQLite(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "gorm_sqlite_test.db")
	log := slog.New(slog.NewTextHandler(io.Discard, nil))

	db, closeFn, err := NewGorm(&GormConfig{
		Type:   "sqlite",
		DBPath: dbPath,
	}, log)
	if err != nil {
		t.Fatalf("NewGorm sqlite: %v", err)
	}
	defer closeFn()

	if err := db.AutoMigrate(&gormSQLiteTestUser{}); err != nil {
		t.Fatalf("auto migrate sqlite test user: %v", err)
	}

	user := gormSQLiteTestUser{
		Name:  "SQLite User",
		Email: "sqlite-user@example.com",
	}
	if err := db.Create(&user).Error; err != nil {
		t.Fatalf("create sqlite test user: %v", err)
	}
	if user.ID == 0 {
		t.Fatal("expected sqlite test user ID to be set")
	}

	var got gormSQLiteTestUser
	if err := db.Where("email = ?", user.Email).First(&got).Error; err != nil {
		t.Fatalf("query sqlite test user: %v", err)
	}
	if got.Name != user.Name {
		t.Fatalf("sqlite test user name = %q, want %q", got.Name, user.Name)
	}
}
