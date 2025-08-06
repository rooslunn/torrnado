package main

import (
	"log/slog"
	"time"
)


type parsed struct {
	title string
	quality string
	format string
	created_at time.Time
	liked int
	status string // проверено
	size string // 
	magnet string
	torrent_file string
	audio []string
	video []string
}

// https://rutracker.org/forum/viewtopic.php?t=4624488 семь психопатов
// https://rutracker.org/forum/viewtopic.php?t=5104420 голгофа
// https://rutracker.org/forum/viewtopic.php?t=5352911 война против всех
// https://rutracker.org/forum/viewtopic.php?t=6238349 пройщенный
// https://rutracker.org/forum/viewtopic.php?t=5527258 шестизарядник

type parserCmd struct {}

func (c parserCmd) execute(log *slog.Logger, args []string) error {
	return nil
}