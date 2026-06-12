package vbasedata

import (
	"errors"
	"fmt"
	"log/slog"
	"net"
	"os"
	"strings"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
	pgconn "github.com/jackc/pgx/v5/pgconn"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

func systemTimeZoneName() string {
	if tz := strings.TrimSpace(os.Getenv("TZ")); tz != "" {
		return tz
	}
	if b, err := os.ReadFile("/etc/timezone"); err == nil {
		if tz := strings.TrimSpace(string(b)); tz != "" {
			return tz
		}
	}
	if link, err := os.Readlink("/etc/localtime"); err == nil {
		const prefix = "/usr/share/zoneinfo/"
		if idx := strings.Index(link, prefix); idx >= 0 {
			if tz := strings.TrimSpace(link[idx+len(prefix):]); tz != "" {
				return tz
			}
		}
	}
	return "UTC"
}

// parseLogLevel 将配置中的字符串日志级别转换为 gorm logger.LogLevel，默认 warn。
func parseLogLevel(level string) logger.LogLevel {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "silent":
		return logger.Silent
	case "warn":
		return logger.Warn
	case "error":
		return logger.Error
	case "info", "debug":
		return logger.Info
	default:
		return logger.Warn
	}
}

func isMySQLUnknownDatabaseErr(err error) bool {
	var mysqlErr *mysqldriver.MySQLError
	if errors.As(err, &mysqlErr) {
		return mysqlErr.Number == 1049
	}
	return false
}

func isPGDatabaseDoesNotExistErr(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "3D000"
	}
	return strings.Contains(strings.ToLower(err.Error()), "does not exist")
}

func quoteMySQLIdent(s string) string {
	return "`" + strings.ReplaceAll(s, "`", "``") + "`"
}

func quotePGIdent(s string) string {
	return "\"" + strings.ReplaceAll(s, "\"", "\"\"") + "\""
}

func buildMySQLDSN(c *GormConfig, dbName string) string {
	return (&mysqldriver.Config{
		User:      c.Username,
		Passwd:    c.Password,
		Net:       "tcp",
		Addr:      c.Address,
		DBName:    dbName,
		ParseTime: true,
		Loc:       time.Local,
		Params: map[string]string{
			"charset": "utf8mb4",
		},
	}).FormatDSN()
}

func newMySQLDialector(c *GormConfig, dbName string) gorm.Dialector {
	return mysql.New(mysql.Config{
		DSN:                       buildMySQLDSN(c, dbName),
		DefaultStringSize:         256,
		DisableDatetimePrecision:  true,
		DontSupportRenameIndex:    true,
		DontSupportRenameColumn:   true,
		SkipInitializeWithVersion: false,
	})
}

func ensureMySQLDatabase(c *GormConfig, glog *gorm.Config) error {
	adminDB, err := gorm.Open(newMySQLDialector(c, ""), glog)
	if err != nil {
		return err
	}
	sqlDB, err := adminDB.DB()
	if err != nil {
		return err
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	createSQL := fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s DEFAULT CHARACTER SET utf8mb4", quoteMySQLIdent(c.DBName))
	return adminDB.Exec(createSQL).Error
}

func buildPGDsn(host, port, username, password, dbname, sslmode, tz string) string {
	base := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		host, username, password, dbname, port, sslmode,
	)
	if tz == "" || strings.EqualFold(tz, "local") {
		return base
	}
	return base + fmt.Sprintf(" TimeZone=%s", tz)
}

func normalizePostgresOptions(c *GormConfig) (host, port, sslmode, tz string) {
	host = c.Address
	port = "5432"
	if h, p, err := net.SplitHostPort(c.Address); err == nil {
		host = h
		if p != "" {
			port = p
		}
	} else if h, p, ok := strings.Cut(c.Address, ":"); ok {
		host = h
		if p != "" {
			port = p
		}
	}

	sslmode = strings.TrimSpace(c.SSLMode)
	if sslmode == "" {
		sslmode = "disable"
	}

	tz = strings.TrimSpace(c.TimeZone)
	if tz == "" || strings.EqualFold(tz, "local") {
		tz = systemTimeZoneName()
	}
	return host, port, sslmode, tz
}

func ensurePGDatabase(c *GormConfig, glog *gorm.Config, host, port, sslmode, tz string) error {
	adminDsn := buildPGDsn(host, port, c.Username, c.Password, "postgres", sslmode, tz)
	adminDB, err := gorm.Open(postgres.Open(adminDsn), glog)
	if err != nil {
		return err
	}
	sqlDB, err := adminDB.DB()
	if err != nil {
		return err
	}
	defer func() {
		_ = sqlDB.Close()
	}()

	var exists int
	if err := adminDB.Raw("SELECT 1 FROM pg_database WHERE datname = ?", c.DBName).Scan(&exists).Error; err != nil {
		return err
	}
	if exists == 1 {
		return nil
	}

	createSQL := fmt.Sprintf("CREATE DATABASE %s", quotePGIdent(c.DBName))
	return adminDB.Exec(createSQL).Error
}

type GormConfig struct {
	Type      string     `yaml:"type" json:"type"`         //类型 mysql/sqlite
	DBPath    string     `yaml:"db_path" json:"db_path"`   //数据库路径
	Name      string     `yaml:"name" json:"name"`         //别名，用来区分多个gorm客户端
	Username  string     `yaml:"username" json:"username"` // 数据库用户名
	Password  string     `yaml:"password" json:"password"` // 数据库密码
	Address   string     `yaml:"address" json:"address"`   // 数据库地址
	DBName    string     `yaml:"dbname" json:"dbname"`     // 数据库名称
	SSLMode   string     `yaml:"sslmode" json:"sslmode"`
	TimeZone  string     `yaml:"timezone" json:"timezone"`
	Logconfig *Logconfig `yaml:"logconfig" json:"logconfig"` // 日志配置
	Conns     *Conns     `yaml:"conns" json:"conns"`         // 连接池配置
}

// Logconfig 日志配置
type Logconfig struct {
	SlowThreshold             int    `yaml:"slow_threshold" json:"slow_threshold"`                               // 慢 SQL 阈值 单位：毫秒
	IgnoreRecordNotFoundError bool   `yaml:"ignore_record_not_found_error" json:"ignore_record_not_found_error"` // 忽略ErrRecordNotFound（记录未找到）错误
	Colorful                  bool   `yaml:"colorful" json:"colorful"`                                           // 是否彩色打印
	ParameterizedQueries      bool   `yaml:"parameterized_queries" json:"parameterized_queries"`
	Level                     string `yaml:"level" json:"level"`
}

// Conns 连接池配置
type Conns struct {
	Maxidle     int `yaml:"maxidle" json:"maxidle"`         // 最大空闲连接数
	Maxopen     int `yaml:"maxopen" json:"maxopen"`         // 最大连接数
	Maxlifetime int `yaml:"maxlifetime" json:"maxlifetime"` // 连接最大存活时间 单位：秒
}

func defaultLogConfig(conf *Logconfig) Logconfig {
	defaults := Logconfig{
		SlowThreshold:             3000,
		IgnoreRecordNotFoundError: true,
	}
	if conf == nil {
		return defaults
	}
	defaults.Colorful = conf.Colorful
	defaults.ParameterizedQueries = conf.ParameterizedQueries
	defaults.Level = conf.Level
	defaults.IgnoreRecordNotFoundError = conf.IgnoreRecordNotFoundError
	if conf.SlowThreshold > 0 {
		defaults.SlowThreshold = conf.SlowThreshold
	}
	return defaults
}

func defaultConns(conf *Conns) Conns {
	defaults := Conns{
		Maxidle:     5,
		Maxopen:     10,
		Maxlifetime: 1800,
	}
	if conf == nil {
		return defaults
	}
	if conf.Maxidle > 0 {
		defaults.Maxidle = conf.Maxidle
	}
	if conf.Maxopen > 0 {
		defaults.Maxopen = conf.Maxopen
	}
	if conf.Maxlifetime > 0 {
		defaults.Maxlifetime = conf.Maxlifetime
	}
	return defaults
}

func normalizeGormConfig(c *GormConfig) GormConfig {
	cfg := *c
	logConfig := defaultLogConfig(c.Logconfig)
	conns := defaultConns(c.Conns)
	cfg.Type = strings.ToLower(strings.TrimSpace(c.Type))
	cfg.DBPath = strings.TrimSpace(c.DBPath)
	if cfg.DBPath == "" {
		cfg.DBPath = "data.db"
	}
	cfg.Logconfig = &logConfig
	cfg.Conns = &conns
	return cfg
}

func applyNormalizedGormConfig(dst *GormConfig, src *GormConfig) {
	dst.Type = src.Type
	dst.DBPath = src.DBPath
	dst.Logconfig = src.Logconfig
	dst.Conns = src.Conns
}

func newGormConfig(c *GormConfig, log *slog.Logger) *gorm.Config {
	glog := &gorm.Config{
		Logger: logger.NewSlogLogger(log, logger.Config{
			SlowThreshold:             time.Duration(c.Logconfig.SlowThreshold) * time.Millisecond,
			Colorful:                  c.Logconfig.Colorful,
			IgnoreRecordNotFoundError: c.Logconfig.IgnoreRecordNotFoundError,
			ParameterizedQueries:      c.Logconfig.ParameterizedQueries,
			LogLevel:                  parseLogLevel(c.Logconfig.Level),
		}),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
	}
	return glog
}

func openGormDB(c *GormConfig, glog *gorm.Config) (*gorm.DB, error) {
	switch c.Type {
	case "mysql":
		db, err := gorm.Open(newMySQLDialector(c, c.DBName), glog)
		if err != nil {
			if isMySQLUnknownDatabaseErr(err) {
				if err2 := ensureMySQLDatabase(c, glog); err2 != nil {
					return nil, err2
				}
				db, err = gorm.Open(newMySQLDialector(c, c.DBName), glog)
			}
			if err != nil {
				return nil, err
			}
		}
		return db, nil
	case "pg", "postgres":
		host, port, sslmode, tz := normalizePostgresOptions(c)
		dsn := buildPGDsn(host, port, c.Username, c.Password, c.DBName, sslmode, tz)
		db, err := gorm.Open(postgres.Open(dsn), glog)
		if err != nil {
			if isPGDatabaseDoesNotExistErr(err) {
				if err2 := ensurePGDatabase(c, glog, host, port, sslmode, tz); err2 != nil {
					return nil, err2
				}
				db, err = gorm.Open(postgres.Open(dsn), glog)
			}
			if err != nil {
				return nil, err
			}
		}
		return db, nil
	default:
		return gorm.Open(sqlite.Open(c.DBPath), glog)
	}
}

func logDBConnected(log *slog.Logger, c *GormConfig) {
	switch c.Type {
	case "mysql":
		log.Info(fmt.Sprintf("数据库配置:%s:******@tcp(%s)/%s?charset=utf8mb4&parseTime=true&loc=Local 连接成功", c.Username, c.Address, c.DBName))
	case "pg", "postgres":
		log.Info(fmt.Sprintf("数据库配置:%s:******@%s/%s 连接成功", c.Username, c.Address, c.DBName))
	default:
		log.Info(fmt.Sprintf("数据库配置:%s:连接成功", c.DBPath))
	}
}

func applyConnPool(sqlDB interface {
	SetMaxIdleConns(int)
	SetMaxOpenConns(int)
	SetConnMaxLifetime(time.Duration)
}, conns *Conns) {
	sqlDB.SetMaxIdleConns(conns.Maxidle)
	sqlDB.SetMaxOpenConns(conns.Maxopen)
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(conns.Maxlifetime))
}

// NewGorm 初始化一个 gorm 客户端。
func NewGorm(c *GormConfig, log *slog.Logger) (*gorm.DB, func(), error) {
	if c == nil {
		return nil, nil, errors.New("GORM配置参数不能为空")
	}
	if log == nil {
		log = slog.Default()
	}

	cfg := normalizeGormConfig(c)
	applyNormalizedGormConfig(c, &cfg)
	glog := newGormConfig(&cfg, log)

	db, err := openGormDB(&cfg, glog)
	if err != nil {
		return nil, nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		log.Error(fmt.Sprintf("DB:%v PING错误,%v", cfg.DBName, err))
		return nil, nil, err
	}
	logDBConnected(log, &cfg)

	applyConnPool(sqlDB, cfg.Conns)
	theF := func() {
		log.Info(fmt.Sprintf("DB 连接池关闭-%v", cfg.DBName))
		if err := sqlDB.Close(); err != nil {
			log.Error(fmt.Sprintf("DB 连接池关闭失败-%v", cfg.DBName))
		}
	}
	return db, theF, nil
}
