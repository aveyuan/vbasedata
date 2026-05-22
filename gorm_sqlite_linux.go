//go:build linux
// +build linux

package vbasedata

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openSQLite(dbPath string, glog *gorm.Config) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dbPath), glog)
}
