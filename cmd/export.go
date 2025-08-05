package main

import (
	"fmt"
	"log/slog"
	"os"
)

type exportCmd struct{}

func (c exportCmd) execute(log *slog.Logger, args []string) error {
	if len(os.Args) < 2 {
		c.expoTrick()
		return ErrDefectiveArgs
	}

	log.Info("executing export topic command", "args", args)

	return nil
}

func (c exportCmd) expoTrick() {
	fmt.Println("Usage: torrnado export <topic_id> <filepath>")
}
