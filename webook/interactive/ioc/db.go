package ioc

import (
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"

	"webook/webook/internal/pkg/logger"
	"webook/webook/internal/repository/dao"
	"webook/webook/pkg/gormx/connpool"
)

func InitSRC(l logger.Logger) SrcDB {
	return InitDB(l, "src")
}

func InitDST(l logger.Logger) DstDB {
	return InitDB(l, "dst")
}

func InitDoubleWritePool(src SrcDB, dst DstDB) *connpool.DoubleWritePool {
	pattern := viper.GetString("migrator.pattern")
	if pattern == "" {
		pattern = connpool.PatternSrcOnly
	}
	return connpool.NewDoubleWritePool(src.ConnPool, dst.ConnPool, pattern)
}

// InitBizDB 这个是业务用的，支持双写的 DB
func InitBizDB(pool *connpool.DoubleWritePool) *gorm.DB {
	db, err := gorm.Open(mysql.New(mysql.Config{
		Conn: pool,
	}))
	if err != nil {
		panic(err)
	}
	return db
}

type SrcDB *gorm.DB
type DstDB *gorm.DB

func InitDB(l logger.Logger, key string) *gorm.DB {
	_ = l
	type Config struct {
		DSN string `yaml:"dsn"`
	}
	var cfg = Config{
		DSN: "root:root@tcp(localhost:13316)/webook_default",
	}
	err := viper.UnmarshalKey("db."+key, &cfg)
	if err != nil {
		panic(err)
	}
	db, err := gorm.Open(mysql.Open(cfg.DSN), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	err = dao.InitTable(db)
	if err != nil {
		panic(err)
	}
	return db
}
