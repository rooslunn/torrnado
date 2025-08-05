package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/rooslunn/torrnado"
)

type exportCmd struct{}

func (c exportCmd) execute(log *slog.Logger, args []string) error {
	if len(args) ==  0 {
		c.expoTrick()
		return ErrDefectiveArgs
	}

	log.Info("executing export topic command", "args", args)
	
	topic_id, err := strconv.Atoi(args[0])
	if err != nil {
		c.expoTrick()
		return ErrDefectiveArgs
	}

	dbpath := os.Getenv("TORR_DB")
	db, err := torrnado.MustSaveToLite(dbpath)
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

func (c exportCmd) expoTrick() {
	fmt.Println("Usage: torrnado export <topic_id> <filepath>")
}
