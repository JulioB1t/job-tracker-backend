package config

import "os"

type Config struct {
	HTTPAddr string
}

func Load() Config {
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		port := os.Getenv("PORT")
		if port == "" {
			port = "8080"
		}
		addr = ":" + port
	}

	return Config{
		HTTPAddr: addr,
	}
}
