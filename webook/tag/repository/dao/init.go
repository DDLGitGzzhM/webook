package dao

import "gorm.io/gorm"

// InitTables 初始化标签相关表。
func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(
		&Tag{},
		&TagBiz{},
	)
}
