package config

type config struct {
	Db    DBConfig
	Redis RedisConfig
}

type DBConfig struct {
	DSN string
}

type RedisConfig struct {
	Addr string
}
