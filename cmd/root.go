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

	log.Info("Starting root command")

	// config
	config, err := torrnado.MustConfig()
	if err != nil {
		log.Error(err.Error())
		return  err
	}
	log.Info("config loaded", "url", config.Env[torrnado.TORR_URL], "db", config.Env[torrnado.TORR_DB])

	// db
	db, err := torrnado.MustSaveToLite(config.StoragePath)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("DB connected", "status", db.Status)

	// get Topic
	rt, err := torrnado.NeedSource(config)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("loggged to source", "status", rt.Status)

	// parse Topic

	// finalization

	return nil
}