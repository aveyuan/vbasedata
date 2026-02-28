package vbasedata

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func newTestLogger() *log.Helper {
	l := log.NewStdLogger(os.Stdout)
	return log.NewHelper(l)
}

func TestNewGorm_MySQL_AutoCreateDatabase(t *testing.T) {
	addr := os.Getenv("VB_TEST_MYSQL_ADDR")
	user := os.Getenv("VB_TEST_MYSQL_USER")
	pass := os.Getenv("VB_TEST_MYSQL_PASS")
	if addr == "" || user == "" {
		t.Skip("set VB_TEST_MYSQL_ADDR, VB_TEST_MYSQL_USER, VB_TEST_MYSQL_PASS to run")
	}

	dbName := fmt.Sprintf("vbasedata_test_%d", time.Now().UnixNano())

	adminDsn := fmt.Sprintf("%s:%s@tcp(%s)/?charset=utf8&parseTime=True&loc=Local", user, pass, addr)
	adminDB, err := gorm.Open(mysql.Open(adminDsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("admin mysql open: %v", err)
	}
	_ = adminDB.Exec("DROP DATABASE IF EXISTS `" + dbName + "`").Error

	gdb, closeFn, err := NewGorm(&GormConfig{
		Type:     "mysql",
		Username: user,
		Password: pass,
		Address:  addr,
		DBName:   dbName,
	}, newTestLogger())
	if err != nil {
		t.Fatalf("NewGorm mysql: %v", err)
	}
	closeFn()

	sqlDB, err := gdb.DB()
	if err == nil {
		_ = sqlDB.Close()
	}

	_ = adminDB.Exec("DROP DATABASE IF EXISTS `" + dbName + "`").Error
}

func TestNewGorm_PG_AutoCreateDatabase(t *testing.T) {
	addr := os.Getenv("VB_TEST_PG_ADDR")
	user := os.Getenv("VB_TEST_PG_USER")
	pass := os.Getenv("VB_TEST_PG_PASS")
	sslmode := os.Getenv("VB_TEST_PG_SSLMODE")
	tz := os.Getenv("VB_TEST_PG_TIMEZONE")
	if addr == "" || user == "" {
		t.Skip("set VB_TEST_PG_ADDR, VB_TEST_PG_USER, VB_TEST_PG_PASS (optional: VB_TEST_PG_SSLMODE, VB_TEST_PG_TIMEZONE) to run")
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	dbName := fmt.Sprintf("vbasedata_test_%d", time.Now().UnixNano())

	host := addr
	port := "5432"
	if i := len(addr); i > 0 {
		// reuse the same parsing strategy as NewGorm
		if idx := indexByte(addr, ':'); idx >= 0 {
			host = addr[:idx]
			if idx+1 < len(addr) && addr[idx+1:] != "" {
				port = addr[idx+1:]
			}
		}
	}

	adminDsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s",
		host, user, pass, port, sslmode,
	)
	if tz != "" {
		adminDsn = adminDsn + " TimeZone=" + tz
	}
	adminDB, err := gorm.Open(postgres.Open(adminDsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("admin pg open: %v", err)
	}
	_ = adminDB.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = ?", dbName).Error
	_ = adminDB.Exec("DROP DATABASE IF EXISTS \"" + dbName + "\"").Error

	gdb, closeFn, err := NewGorm(&GormConfig{
		Type:     "pg",
		Username: user,
		Password: pass,
		Address:  addr,
		DBName:   dbName,
		SSLMode:  sslmode,
		TimeZone: tz,
	}, newTestLogger())
	if err != nil {
		t.Fatalf("NewGorm pg: %v", err)
	}
	closeFn()

	sqlDB, err := gdb.DB()
	if err == nil {
		_ = sqlDB.Close()
	}

	_ = adminDB.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = ?", dbName).Error
	_ = adminDB.Exec("DROP DATABASE IF EXISTS \"" + dbName + "\"").Error
}

func indexByte(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}
