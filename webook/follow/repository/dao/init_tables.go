package dao

import "gorm.io/gorm"

// InitTables 初始化关注关系相关表
func InitTables(db *gorm.DB) error {
	return db.AutoMigrate(&FollowRelation{})
}
