package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
)

type Command interface {
	execute(log *slog.Logger, args []string) error
	expoTrick()
}

var ErrDefectiveArgs = errors.New("not enough args")

func main() {
	log := setupSnitch(os.Stdout)

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "you forgot command nick (fetch, export, clean)")
		os.Exit(2)
	}

	var command Command

	commandBrand := os.Args[1]

	switch commandBrand {
	case "fetch":
		command = new(fetchCmd)
	case "export":
		command = new(exportCmd)
	case "clean":
		command = new(cleanCmd)
	default:
	}

	goodbyeIfFuckedUp(
		command,
		log,
		os.Args[2:],
	)
}

func setupSnitch(out io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func goodbyeIfFuckedUp(command Command, log *slog.Logger, args []string) {
	err := command.execute(log, args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
