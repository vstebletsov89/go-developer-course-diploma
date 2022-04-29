package configs

import (
	"flag"
	"github.com/caarlos0/env/v6"
)

type Config struct {
	RunAddress           string `env:"RUN_ADDRESS" envDefault:"localhost:8080"`
	DatabaseURI          string `env:"DATABASE_URI" envDefault:""`
	AccrualSystemAddress string `env:"ACCRUAL_SYSTEM_ADDRESS" envDefault:""`
	LogLevel             string `env:"LOG_LEVEL" envDefault:"debug"`
}

func (c *Config) readCommandLineArgs() {
	flag.StringVar(&c.RunAddress, "a", c.RunAddress, "server and port to listen on")
	flag.StringVar(&c.DatabaseURI, "d", c.DatabaseURI, "database URI")
	flag.StringVar(&c.AccrualSystemAddress, "r", c.AccrualSystemAddress, "address of external accrual system")
	flag.StringVar(&c.LogLevel, "l", c.LogLevel, "log level")
	flag.Parse()
}

func ReadConfig() (*Config, error) {
	var cfg Config
	err := env.Parse(&cfg)

	if err != nil {
		return nil, err
	}
	cfg.readCommandLineArgs()

	return &cfg, nil
}
