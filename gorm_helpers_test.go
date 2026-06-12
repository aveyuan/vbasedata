package vbasedata

import (
	"errors"
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"testing"

	mysqldriver "github.com/go-sql-driver/mysql"
	pgconn "github.com/jackc/pgx/v5/pgconn"
	glogger "gorm.io/gorm/logger"
)

func TestQuoteMySQLIdent(t *testing.T) {
	cases := map[string]string{
		"db":      "`db`",
		"my`db":   "`my``db`",
		"a`b`c":   "`a``b``c`",
		"":        "``",
	}
	for in, want := range cases {
		if got := quoteMySQLIdent(in); got != want {
			t.Errorf("quoteMySQLIdent(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestQuotePGIdent(t *testing.T) {
	cases := map[string]string{
		"db":      `"db"`,
		`my"db`:   `"my""db"`,
		`a"b"c`:   `"a""b""c"`,
		"":        `""`,
	}
	for in, want := range cases {
		if got := quotePGIdent(in); got != want {
			t.Errorf("quotePGIdent(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestBuildPGDsn(t *testing.T) {
	const base = "host=h user=u password=p dbname=d port=5432 sslmode=disable"

	t.Run("empty tz omits TimeZone", func(t *testing.T) {
		got := buildPGDsn("h", "5432", "u", "p", "d", "disable", "")
		if got != base {
			t.Errorf("got %q, want %q", got, base)
		}
	})

	t.Run("local tz omits TimeZone", func(t *testing.T) {
		got := buildPGDsn("h", "5432", "u", "p", "d", "disable", "local")
		if got != base {
			t.Errorf("got %q, want %q", got, base)
		}
		// case-insensitive
		if got := buildPGDsn("h", "5432", "u", "p", "d", "disable", "Local"); got != base {
			t.Errorf("Local: got %q, want %q", got, base)
		}
	})

	t.Run("explicit tz appended", func(t *testing.T) {
		got := buildPGDsn("h", "5432", "u", "p", "d", "disable", "Asia/Shanghai")
		want := base + " TimeZone=Asia/Shanghai"
		if got != want {
			t.Errorf("got %q, want %q", got, want)
		}
	})
}

func TestParseLogLevel(t *testing.T) {
	cases := map[string]glogger.LogLevel{
		"silent": glogger.Silent,
		"error":  glogger.Error,
		"warn":   glogger.Warn,
		"info":   glogger.Info,
		"INFO":   glogger.Info, // case-insensitive
		" warn ": glogger.Warn, // trimmed
		"":       glogger.Warn, // default
		"bogus":  glogger.Warn, // default
	}
	for in, want := range cases {
		if got := parseLogLevel(in); got != want {
			t.Errorf("parseLogLevel(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestIsMySQLUnknownDatabaseErr(t *testing.T) {
	if !isMySQLUnknownDatabaseErr(&mysqldriver.MySQLError{Number: 1049}) {
		t.Error("expected true for MySQL error 1049")
	}
	if isMySQLUnknownDatabaseErr(&mysqldriver.MySQLError{Number: 1234}) {
		t.Error("expected false for unrelated MySQL error")
	}
	if isMySQLUnknownDatabaseErr(errors.New("boom")) {
		t.Error("expected false for non-MySQL error")
	}
	if isMySQLUnknownDatabaseErr(nil) {
		t.Error("expected false for nil error")
	}
}

func TestIsPGDatabaseDoesNotExistErr(t *testing.T) {
	if !isPGDatabaseDoesNotExistErr(&pgconn.PgError{Code: "3D000"}) {
		t.Error("expected true for PG SQLSTATE 3D000")
	}
	if !isPGDatabaseDoesNotExistErr(errors.New(`database "x" does not exist`)) {
		t.Error("expected true for message containing 'does not exist'")
	}
	if isPGDatabaseDoesNotExistErr(&pgconn.PgError{Code: "42P01"}) {
		t.Error("expected false for unrelated PG code")
	}
	if isPGDatabaseDoesNotExistErr(errors.New("connection refused")) {
		t.Error("expected false for unrelated error")
	}
}

func TestSystemTimeZoneName(t *testing.T) {
	t.Run("TZ env wins", func(t *testing.T) {
		t.Setenv("TZ", "Asia/Tokyo")
		if got := systemTimeZoneName(); got != "Asia/Tokyo" {
			t.Errorf("got %q, want Asia/Tokyo", got)
		}
	})
	t.Run("never empty", func(t *testing.T) {
		t.Setenv("TZ", "")
		if got := systemTimeZoneName(); strings.TrimSpace(got) == "" {
			t.Error("systemTimeZoneName returned empty")
		}
	})
}

// TestNewGorm_Defaults verifies that NewGorm fills in default Conns/Logconfig
// values in-place when the caller leaves them unset.
func TestNewGorm_Defaults(t *testing.T) {
	log := slog.New(slog.NewTextHandler(io.Discard, nil))
	c := &GormConfig{
		Type:   "sqlite",
		DBPath: filepath.Join(t.TempDir(), "defaults.db"),
	}
	_, closeFn, err := NewGorm(c, log)
	if err != nil {
		t.Fatalf("NewGorm: %v", err)
	}
	defer closeFn()

	if c.Conns == nil {
		t.Fatal("expected Conns to be populated")
	}
	if c.Conns.Maxidle != 5 || c.Conns.Maxopen != 10 || c.Conns.Maxlifetime != 1800 {
		t.Errorf("unexpected Conns defaults: %+v", c.Conns)
	}
	if c.Logconfig == nil {
		t.Fatal("expected Logconfig to be populated")
	}
	if c.Logconfig.SlowThreshold != 3000 {
		t.Errorf("SlowThreshold = %d, want 3000", c.Logconfig.SlowThreshold)
	}
}

func TestNewGorm_NilConfig(t *testing.T) {
	if _, _, err := NewGorm(nil, slog.Default()); err == nil {
		t.Error("expected error for nil config")
	}
}
