package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Address         string
	DatabaseDSN     string
	WebRoot         string
	ShutdownTimeout time.Duration
	WorkerInterval  time.Duration
	WorkerBatchSize int
	DefaultTenant   string
}

func Load() (Config, error) {
	cfg := Config{Address: env("WINDOPS_ADDR", ":8080"), DatabaseDSN: env("WINDOPS_DB", "./data/windops.db"), WebRoot: env("WINDOPS_WEB", "./web/dist"), DefaultTenant: env("WINDOPS_TENANT", "demo"), ShutdownTimeout: 10 * time.Second, WorkerInterval: 2 * time.Second, WorkerBatchSize: 25}
	var err error
	if cfg.ShutdownTimeout, err = duration("WINDOPS_SHUTDOWN_TIMEOUT", cfg.ShutdownTimeout); err != nil {
		return Config{}, err
	}
	if cfg.WorkerInterval, err = duration("WINDOPS_WORKER_INTERVAL", cfg.WorkerInterval); err != nil {
		return Config{}, err
	}
	if cfg.WorkerBatchSize, err = integer("WINDOPS_WORKER_BATCH", cfg.WorkerBatchSize); err != nil {
		return Config{}, err
	}
	if strings.TrimSpace(cfg.Address) == "" || strings.TrimSpace(cfg.DatabaseDSN) == "" || strings.TrimSpace(cfg.DefaultTenant) == "" {
		return Config{}, fmt.Errorf("address, database and tenant are required")
	}
	if cfg.WorkerBatchSize < 1 || cfg.WorkerBatchSize > 500 {
		return Config{}, fmt.Errorf("worker batch must be between 1 and 500")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value := strings.TrimSpace(os.Getenv(key)); value != "" {
		return value
	}
	return fallback
}
func duration(key string, fallback time.Duration) (time.Duration, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := time.ParseDuration(raw)
	if err != nil || value <= 0 {
		return 0, fmt.Errorf("%s must be a positive duration", key)
	}
	return value, nil
}
func integer(key string, fallback int) (int, error) {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, fmt.Errorf("%s must be an integer", key)
	}
	return value, nil
}
