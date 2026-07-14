//go:build !k8s

package config

var Config = config{
	Db: DBConfig{
		DSN: "localhost:3306",
	},
	Redis: RedisConfig{
		Addr: "localhost:6379",
	},
}
