package main

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/rooslunn/torrnado"
)

type Command interface {
	execute(args []string) error
	expoTrick()
}

var (
	ErrDefectiveArgs = errors.New("not enough args")
	ErrUnknownCommand = errors.New("exotic commmand")
	ErrUnknownSubCommand = errors.New("unplanned sub commmand")
)

const (
	CMD_PING = "ping"
	CMD_DB = "db:"
	CMD_TRACKER = "tracker:"
	CMD_MOVIEDB = "moviedb:"
)

func main() {
	log := setupSnitch(os.Stdout)

	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "you forgot command nick (fetch, export, clean)")
		os.Exit(2)
	}

	var command Command

	commandBrand := os.Args[1]

	if commandBrand == "ping" {
		os.Exit(pingPong(log))
	}

	if strings.HasPrefix(commandBrand, CMD_TRACKER) {
		command = trackerCmd{log}
	} else if strings.HasPrefix(commandBrand, CMD_DB) {
		command = dbCmd{log}
	} else {
		fmt.Fprintln(os.Stderr, ErrUnknownCommand)
		os.Exit(1)
	}

	goodbyeIfFuckedUp(
		command,
		os.Args[1:],
	)
}

func pingPong(log *slog.Logger) int {
	log.Info("checking config...")
	cfg, err := torrnado.MustConfig()
	if err != nil {
		log.Error(err.Error())	
		return 1
	}
	for k, v := range cfg.Env {
		log.Info("config", k, v)
	}

	log.Info("checking db contact...")
	_, err = joinDb()
	if err != nil {
		log.Error(err.Error())
		return 1
	}
	log.Info("db is available")

	return 0
}

func setupSnitch(out io.Writer) *slog.Logger {
	return slog.New(slog.NewTextHandler(out, &slog.HandlerOptions{Level: slog.LevelInfo}))
}

func goodbyeIfFuckedUp(command Command, args []string) {
	err := command.execute(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func isSubcommand(command, sub_command string) bool {
	parts := strings.Split(command, ":")
	if len(parts) < 2 {
		return false
	}
	return parts[1] == sub_command
}
