package config

import (
	"os"
	"strconv"
)

type Config struct {
	Debug       bool
	HostRootDir string
	API         struct {
		BindAddress string
	}
	Sandbox struct {
		TimeoutSeconds int
		Memory         string
	}
	Redis struct {
		Addr string
		Pwd  string
	}
	Database struct {
		DSN string
	}
	WorkerCount int
}

type EnvProvider struct {
	prefix string
	c      Config
}

func NewEnvProvider(prefix string) *EnvProvider {
	return &EnvProvider{prefix: prefix}
}

func (ep *EnvProvider) Load() error {
	ep.c.Debug = os.Getenv(ep.prefix+"DEBUG") == "true"
	ep.c.HostRootDir = getEnv(ep.prefix+"HOSTROOTDIR", "./data")
	ep.c.API.BindAddress = getEnv(ep.prefix+"API_BINDADDRESS", ":8080")
	ep.c.Sandbox.Memory = getEnv(ep.prefix+"SANDBOX_MEMORY", "100M")
	ep.c.Sandbox.TimeoutSeconds, _ = strconv.Atoi(getEnv(ep.prefix+"SANDBOX_TIMEOUTSECONDS", "20"))
	
	ep.c.Redis.Addr = getEnv(ep.prefix+"REDIS_ADDR", "localhost:6379")
	ep.c.Redis.Pwd = getEnv(ep.prefix+"REDIS_PWD", "")
	
	// Default to local postgres container
	ep.c.Database.DSN = getEnv(ep.prefix+"DB_DSN", "postgres://postgres:postgres@localhost:5432/runner?sslmode=disable")
	
	ep.c.WorkerCount, _ = strconv.Atoi(getEnv(ep.prefix+"WORKER_COUNT", "3"))

	return nil
}

func (ep *EnvProvider) Config() Config { return ep.c }

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
