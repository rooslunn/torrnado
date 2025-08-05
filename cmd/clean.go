package main

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/rooslunn/torrnado"
)

type cleanCmd struct{}

func (c cleanCmd) execute(log *slog.Logger, args []string) error {
	log.Info("executing clean command", "args", args)
	
	dbpath := os.Getenv("TORR_DB")
	db, err := torrnado.MustSaveToLite(dbpath)
	if err != nil {
		return err
	}
	log.Info("connected to db", "path", dbpath)

	records_removed, err := db.Hygienic(log)
	if err != nil {
		return err
	}
	log.Info("cleared", "records evacuated", records_removed)

	return nil
}

func (c cleanCmd) expoTrick() {
	fmt.Println("Usage: torrnado clean")
}
