package vbasedata

import (
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// TestNewGorm_MySQL_AutoCreateDatabase 验证目标库不存在时 NewGorm 会自动建库。
// 优先使用 VB_TEST_MYSQL_* 环境变量；未设置时回退到默认测试凭据，
// 若默认凭据连不上则跳过（视为本地无 MySQL），显式配置了凭据才判定失败。
func TestNewGorm_MySQL_AutoCreateDatabase(t *testing.T) {
	addr, d1 := testEnvDefault("VB_TEST_MYSQL_ADDR", "127.0.0.1:3306")
	user, d2 := testEnvDefault("VB_TEST_MYSQL_USER", "root")
	pass, d3 := testEnvDefault("VB_TEST_MYSQL_PASS", "123456")
	usingDefault := d1 || d2 || d3

	dbName := fmt.Sprintf("vbasedata_test_%d", time.Now().UnixNano())

	adminDsn := fmt.Sprintf("%s:%s@tcp(%s)/?charset=utf8&parseTime=True&loc=Local", user, pass, addr)
	adminDB, err := gorm.Open(mysql.Open(adminDsn), &gorm.Config{})
	if err != nil {
		if usingDefault {
			t.Skipf("mysql is not available at %s with default test credentials: %v", addr, err)
		}
		t.Fatalf("admin mysql open: %v", err)
	}
	dropDB := func() {
		_ = adminDB.Exec("DROP DATABASE IF EXISTS `" + dbName + "`").Error
	}
	dropDB()          // 清理可能残留的同名库
	t.Cleanup(dropDB) // 测试结束后删除

	gdb, closeFn, err := NewGorm(&GormConfig{
		Type:     "mysql",
		Username: user,
		Password: pass,
		Address:  addr,
		DBName:   dbName,
	}, discardLogger())
	if err != nil {
		t.Fatalf("NewGorm mysql: %v", err)
	}
	// Cleanup 为 LIFO：先关连接池再删库
	t.Cleanup(closeFn)

	// 自动创建的库应可正常查询
	if err := gdb.Exec("SELECT 1").Error; err != nil {
		t.Fatalf("query auto-created mysql db: %v", err)
	}
}

// TestNewGorm_PG_AutoCreateDatabase 验证 PostgreSQL 下的自动建库逻辑。
func TestNewGorm_PG_AutoCreateDatabase(t *testing.T) {
	addr, d1 := testEnvDefault("VB_TEST_PG_ADDR", "127.0.0.1:5432")
	user, d2 := testEnvDefault("VB_TEST_PG_USER", "postgres")
	pass, d3 := testEnvDefault("VB_TEST_PG_PASS", "123456")
	sslmode := os.Getenv("VB_TEST_PG_SSLMODE")
	tz := os.Getenv("VB_TEST_PG_TIMEZONE")
	usingDefault := d1 || d2 || d3
	if sslmode == "" {
		sslmode = "disable"
	}

	// 与 NewGorm 一致的 host:port 解析
	host, port := addr, "5432"
	if h, p, ok := strings.Cut(addr, ":"); ok {
		host = h
		if p != "" {
			port = p
		}
	}

	dbName := fmt.Sprintf("vbasedata_test_%d", time.Now().UnixNano())

	adminDsn := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=postgres port=%s sslmode=%s",
		host, user, pass, port, sslmode,
	)
	if tz != "" {
		adminDsn += " TimeZone=" + tz
	}
	adminDB, err := gorm.Open(postgres.Open(adminDsn), &gorm.Config{})
	if err != nil {
		if usingDefault {
			t.Skipf("postgres is not available at %s with default test credentials: %v", addr, err)
		}
		t.Fatalf("admin pg open: %v", err)
	}
	dropDB := func() {
		// PG 删库前需断开占用该库的连接
		_ = adminDB.Exec("SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = ?", dbName).Error
		_ = adminDB.Exec(`DROP DATABASE IF EXISTS "` + dbName + `"`).Error
	}
	dropDB()
	t.Cleanup(dropDB)

	gdb, closeFn, err := NewGorm(&GormConfig{
		Type:     "pg",
		Username: user,
		Password: pass,
		Address:  addr,
		DBName:   dbName,
		SSLMode:  sslmode,
		TimeZone: tz,
	}, discardLogger())
	if err != nil {
		t.Fatalf("NewGorm pg: %v", err)
	}
	t.Cleanup(closeFn)

	if err := gdb.Exec("SELECT 1").Error; err != nil {
		t.Fatalf("query auto-created pg db: %v", err)
	}
}

// testEnvDefault 返回环境变量值；为空时返回 fallback，并以第二个返回值标记是否用了默认值。
func testEnvDefault(key, fallback string) (string, bool) {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value, false
	}
	return fallback, true
}
