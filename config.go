package torrnado

import (
	"fmt"
	"os"
)

type Config struct {
	TopicUrl     string
	StoragePath string
}

const (
	ErrorEnvNotSet = "can't resolve %s env var"
)

func MustConfig() (*Config, error) {
	url := os.Getenv("TORR_URL")
	if url == "" {
		return nil, fmt.Errorf(ErrorEnvNotSet, "TORR_URL")
	}

	db := os.Getenv("TORR_DB")
	if db == "" {
		return nil, fmt.Errorf(ErrorEnvNotSet, "TORR_DB")
	}

	return &Config{
		url,
		db,
	}, nil
}