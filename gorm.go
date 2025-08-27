package vbasedata

import (
	"errors"
	"fmt"
	"time"

	"github.com/go-kratos/kratos/v2/log"

	"github.com/aveyuan/vlogger"

	"github.com/glebarez/sqlite"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	glogger "gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type GormConfig struct {
	Type      string     `yaml:"type" json:"type"`           //类型 mysql/sqlite
	DBPath    string     `yaml:"db_path" json:"db_path"`     //数据库路径
	Name      string     `yaml:"name" json:"name"`           //别名，用来区分多个gorm客户端
	Username  string     `yaml:"username" json:"username"`   // 数据库用户名
	Password  string     `yaml:"password" json:"password"`   // 数据库密码
	Address   string     `yaml:"address" json:"address"`     // 数据库地址
	DBName    string     `yaml:"db_name" json:"db_name"`     // 数据库名称
	Logconfig *Logconfig `yaml:"logconfig" json:"logconfig"` // 日志配置
	Conns     *Conns     `yaml:"conns" json:"conns"`         // 连接池配置
}

// Logconfig 日志配置
type Logconfig struct {
	SlowThreshold             int    `yaml:"slow_threshold" json:"slow_threshold"`                               // 慢 SQL 阈值 单位：毫秒
	IgnoreRecordNotFoundError bool   `yaml:"ignore_record_not_found_error" json:"ignore_record_not_found_error"` // 忽略ErrRecordNotFound（记录未找到）错误
	Colorful                  bool   `yaml:"colorful" json:"colorful"`                                           // 是否彩色打印
	Level                     string `yaml:"level" json:"level"`
}

// Conns 连接池配置
type Conns struct {
	Maxidle     int `yaml:"maxidle" json:"maxidle"`         // 最大空闲连接数
	Maxopen     int `yaml:"maxopen" json:"maxopen"`         // 最大连接数
	Maxlifetime int `yaml:"maxlifetime" json:"maxlifetime"` // 连接最大存活时间 单位：秒
}

// NewGorm 初始化一个gorm的客户端
func NewGorm(c *GormConfig, logger *log.Helper) (*gorm.DB, func(), error) {
	if c == nil {
		return nil, nil, errors.New("GORM配置参数不能为空")
	}
	//默认配置
	if c.Logconfig == nil {
		c.Logconfig = &Logconfig{
			SlowThreshold:             3000,
			IgnoreRecordNotFoundError: true,
		}
	}
	if c.Logconfig.SlowThreshold == 0 {
		c.Logconfig.SlowThreshold = 3000
	}
	//默认配置
	if c.Conns == nil {
		c.Conns = &Conns{
			Maxidle:     5,
			Maxopen:     10,
			Maxlifetime: 1800,
		}
	}
	if c.Conns.Maxidle == 0 {
		c.Conns.Maxidle = 5
	}
	if c.Conns.Maxopen == 0 {
		c.Conns.Maxopen = 10
	}
	if c.Conns.Maxlifetime == 0 {
		c.Conns.Maxlifetime = 1800
	}

	if c.DBPath == "" {
		c.DBPath = "data.db"
	}

	// 设置日志级别
	l, ok := vlogger.LogStr2Level[c.Logconfig.Level]
	if !ok {
		l = -1
	}

	glog := &gorm.Config{
		// 改写日志
		Logger: vlogger.NewGormLog(logger, vlogger.Config{
			SlowThreshold:             time.Duration(c.Logconfig.SlowThreshold) * time.Millisecond, // 慢 SQL 阈值
			LogLevel:                  glogger.LogLevel(l),                                         // 日志级别
			IgnoreRecordNotFoundError: c.Logconfig.IgnoreRecordNotFoundError,                       // 忽略ErrRecordNotFound（记录未找到）错误
			Colorful:                  c.Logconfig.Colorful,                                        // 彩色打印，zap下警用
		}),
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, //单数表
		},
	}

	var db *gorm.DB
	var err error

	if c.Type == "mysql" {
		dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local", c.Username, c.Password, c.Address, c.DBName)
		db, err = gorm.Open(mysql.New(mysql.Config{
			DSN:                       dsn,
			DefaultStringSize:         256,
			DisableDatetimePrecision:  true,
			DontSupportRenameIndex:    true,
			DontSupportRenameColumn:   true,
			SkipInitializeWithVersion: false,
		}), glog)
		if err != nil {
			return nil, nil, err
		}
	} else {
		db, err = gorm.Open(sqlite.Open(c.DBPath), glog)
		if err != nil {
			return nil, nil, err
		}
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}

	if err := sqlDB.Ping(); err != nil {
		logger.Errorf("DB:%v PING错误,%v", c.DBName, err)
		return nil, nil, err
	} else {
		if c.Type == "mysql" {
			logger.Infof("数据库配置:%v", fmt.Sprintf("%s:******@tcp(%s)/%s?charset=utf8&parseTime=True&loc=Local 连接成功", c.Username, c.Address, c.DBName))
		} else {
			logger.Infof("数据库配置:%v", fmt.Sprintf("%s:连接成功", c.DBPath))
		}
	}

	// SetMaxIdleConns 用于设置连接池中空闲连接的最大数量。
	sqlDB.SetMaxIdleConns(c.Conns.Maxidle)
	// SetMaxOpenConns 设置打开数据库连接的最大数量。
	sqlDB.SetMaxOpenConns(c.Conns.Maxopen)
	// SetConnMaxLifetime 设置了连接可复用的最大时间。
	sqlDB.SetConnMaxLifetime(time.Second * time.Duration(c.Conns.Maxlifetime))
	theF := func() {
		logger.Infof("MySQL 连接池关闭-%v", c.DBName)
		if err := sqlDB.Close(); err != nil {
			logger.Errorf("MySQL 连接池关闭失败-%v", c.DBName)
		}
	}
	return db, theF, nil
}
