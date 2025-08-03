package main

import (
	"log/slog"

	"github.com/rooslunn/torrnado"
)

type Command interface {
	Execute(log *slog.Logger) error
}

type rootCmd struct {}

func (c rootCmd) Execute(log *slog.Logger) error {

	log.Info("Starting roor command")

	// config
	config, err := torrnado.MustConfig()
	if err != nil {
		return  err
	}
	log.Info("config loaded", "url", config.TopicUrl, "db", config.StoragePath)

	// db
	db, err := torrnado.MustSaveToLite(config.StoragePath)
	if err != nil {
		return err
	}
	log.Info("DB connected", "status", db.Status)

	// topic parser

	// finalization

	return nil
}