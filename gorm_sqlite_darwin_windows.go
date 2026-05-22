//go:build darwin || windows
// +build darwin windows

package vbasedata

import (
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func openSQLite(dbPath string, glog *gorm.Config) (*gorm.DB, error) {
	return gorm.Open(sqlite.Open(dbPath), glog)
}
