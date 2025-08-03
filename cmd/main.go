package main

import (
	"fmt"
	"io"
	"log/slog"
	"os"
)

func main () {
	log := setupLogger(os.Stdout)
	exitOnError(rootCmd{}, log)
}

func setupLogger(out io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func exitOnError(command Command, log *slog.Logger) {
	err := command.Execute(log)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}