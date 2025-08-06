package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/rooslunn/torrnado"
)

type dbCmd struct{}

func (c dbCmd) execute(log *slog.Logger, args []string) error {
	log.Info("executing db command", "args", args)

	if len(args) == 0 {
		c.expoTrick()
		return ErrDefectiveArgs
	}

	if isSubcommand(args[0], "clean") {
		return clean(log, args[1:])
	} else if isSubcommand(args[0], "export") {
		return export(log, args[1:])
	}

	return ErrUnknownSubCommand
}

func exportTricks() {
	fmt.Println("Usage: export <topic_id>")
}

func export(log *slog.Logger, args []string) error {
	log.Info("executing export sub command", "args", args)

	if len(args) == 0 {
		exportTricks()
		return ErrDefectiveArgs
	}

	topic_id, err := strconv.Atoi(args[0])
	if err != nil {
		exportTricks()
		return ErrDefectiveArgs
	}

	dbpath := os.Getenv("TORR_DB")
	db, err := torrnado.MustHaveStorage(dbpath)
	if err != nil {
		return err
	}
	log.Info("connected to db", "path", dbpath)

	html, err := db.GetHTML(topic_id)
	if err != nil {
		return err
	}
	log.Info("obtained html for", "topic_id", topic_id)

	filename := fmt.Sprintf("%d.html", topic_id)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("Error creating file: %v\n", err)
	}
	defer file.Close()

	_, err = file.WriteString(html)
	if err != nil {
		return fmt.Errorf("error writing to file: %v\n", err)
	}
	log.Info("deliverd html to file", "file", filename)

	return nil
}

func clean(log *slog.Logger, args []string) error {
	log.Info("executing clean sub command", "args", args)

	dbpath := os.Getenv("TORR_DB")
	db, err := torrnado.MustHaveStorage(dbpath)
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

func (c dbCmd) expoTrick() {
	fmt.Println("Usage: torrnado export <topic_id> <filepath>")
}
