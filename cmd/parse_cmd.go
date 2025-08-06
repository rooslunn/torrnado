package main

import "log/slog"

type parserCmd struct {}

func (c parserCmd) execute(log *slog.Logger, args []string) error {
	return nil
}