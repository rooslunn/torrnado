package torrnado

import (
	"database/sql"
	_ "github.com/mattn/go-sqlite3"
)

type Topic struct {
	Title string
	Description string
}

type Storage interface {
	SaveTopic(topic Topic) error
}

type LiteStorage struct {
	db *sql.DB
	Status string
}

const DB_OPTIONS = "?cache=shared"

func MustSaveToLite(path string) (*LiteStorage, error) {
	const op = "storage.sqlite.connect"

	db, err := sql.Open("sqlite3", path + DB_OPTIONS)
	if err != nil {
		return nil, Operror(op, err)
	}

	// stmt, err := db.Prepare("select title from topics limit 1")
	stmt, err := db.Prepare("select 1")
	if err != nil {
		return nil, Operror(op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, Operror(op, err)
	}

	return &LiteStorage{db, "ready"}, nil
}

func (s *LiteStorage) SaveTopic(topic Topic) error {
	return nil
}
