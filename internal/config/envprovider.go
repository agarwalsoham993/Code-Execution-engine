package config

import (
	"os"
	"strconv"
)

type Config struct {
	Debug    bool
	API      struct { BindAddress string }
	Sandbox  struct { TimeoutSeconds int; Memory string }
	HostRootDir string
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
	return nil
}

func (ep *EnvProvider) Config() Config { return ep.c }

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" { return v }
	return def
}
