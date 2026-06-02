//go:build k8s

package config

var Config = config{
	DB: DBConfig{
		DSN: "root:root@tcp(byline-mysql:11309)/byline",
	},
	Redis: RedisConfig{
		Addr: "byline-redis:11479",
	},
}
