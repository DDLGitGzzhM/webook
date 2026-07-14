//go:build k8s

package config

var Config = config{
	Db: DBConfig{
		DSN: "root:root@tcp(webook-mysql:3309)/webook",
	},
	Redis: RedisConfig{
		Addr: "webook-redis:11479",
	},
}
