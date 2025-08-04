package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Command interface {
	Execute(log *slog.Logger) error
}

func main () {
	log := setupSnitch(os.Stdout)
	goodbyeIfFuckedUp(
		RootCmd(), 
		log,
	)
}

func setupSnitch(out io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func goodbyeIfFuckedUp(command Command, log *slog.Logger) {
	err := command.Execute(log)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}