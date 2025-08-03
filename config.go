package torrnado

import (
	"fmt"
	"os"
)

type (
	EnvStore map[string]string
	Config   struct {
		Env         EnvStore
		TopicUrlFmt string
		StoragePath string
	}
)

var EnvVarConfig = []string{TORR_URL, TORR_DB, TORR_USER, TORR_PSWD}

const (
	ErrorEnvNotSet = "can't resolve %s env var"
)

const (
	TORR_URL  = "TORR_URL"
	TORR_DB   = "TORR_DB"
	TORR_USER = "TORR_USER"
	TORR_PSWD = "TORR_PSWD"
)

func MustConfig() (*Config, error) {
	envStore := make(EnvStore, len(EnvVarConfig))

	for _, envVar := range EnvVarConfig {
		value := os.Getenv(envVar)
		if value == "" {
			return nil, fmt.Errorf(ErrorEnvNotSet, envVar)
		}
		envStore[envVar] = value
	}

	return &Config{Env: envStore}, nil
}
