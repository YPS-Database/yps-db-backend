package yps

import (
	"context"

	"github.com/joho/godotenv"
	"github.com/sethvargo/go-envconfig"
)

type Config struct {
	Address                string `env:"ADDR,required"`
	DatabaseUrl            string `env:"DATABASE_URL,required"`
	DatabaseMigrationsPath string `env:"DATABASE_MIGRATIONS_PATH,default=migrations"`
}

func LoadConfig() (config Config, err error) {
	if err := godotenv.Load(); err != nil {
		return config, err
	}

	ctx := context.Background()

	if err := envconfig.Process(ctx, &config); err != nil {
		return config, err
	}

	return config, nil
}
