package main

import (
	"fmt"
	"log/slog"

	"github.com/rooslunn/torrnado"
)

type Command interface {
	Execute(log *slog.Logger) error
}

type rootCmd struct{}

const (
	OUT_PATH = ""
)

func (c rootCmd) Execute(log *slog.Logger) error {
	log.Info("Starting root command")

	// config
	config, err := torrnado.MustConfig()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("config loaded", "url", config.Env[torrnado.TORR_URL], "db", config.Env[torrnado.TORR_DB])

	// db
	db, err := torrnado.MustSaveToLite(config.StoragePath)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("DB connected", "status", db.Status)

	// fetch topic to file
	rt, err := torrnado.NeedSource(config)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("loggged to tracker", "status", rt.Status)
	// 5448425, Fight Club
	url := fmt.Sprintf(config.Env[torrnado.TORR_URL], "5448425")
	filepath := "5448425.html"
	nBytes, err := rt.SaveTopicFile(url, filepath)
	if err != nil {
		return err 
	}
	log.Info("saved topic to file", "topic", url, "file", filepath, "size", nBytes)

	// parse Topic

	// finalization

	return nil
}
