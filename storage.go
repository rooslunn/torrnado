package torrnado

import (
	"database/sql"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

type Topic struct {
	Title       string
	Description string
}

type TopicPlan = map[int]string

type Storage interface {
	SaveHTML(topic_id int, url, html_source string) error
	GetHTML(topic_id int) (string, error)
}

type LiteStorage struct {
	db     *sql.DB
	Status string
}

const DB_OPTIONS = "?cache=shared"

func MustSaveToLite(path string) (*LiteStorage, error) {
	const op = "storage.sqlite.connect"

	db, err := sql.Open("sqlite3", path+DB_OPTIONS)
	if err != nil {
		return nil, Operror(op, err)
	}

	stmt, err := db.Prepare(`
		create table if not exists topics (
			id integer primary key,
			topic_id integer not null,
			url text,
			html_source text,
			created_at text,
			fetched_at text 
		);
		create index idx_topic_id on topics(topic_id);
	`)
	if err != nil {
		return nil, Operror(op, err)
	}

	_, err = stmt.Exec()
	if err != nil {
		return nil, Operror(op, err)
	}

	return &LiteStorage{db, "ready"}, nil
}

func (s *LiteStorage) SaveHTML(topic_id int, html string) error {
	stmt, err := s.db.Prepare("update topics set html_source = ?, fetched_at = ? where topic_id = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(html, ISaidNow(), topic_id)
	return err
}

func (s *LiteStorage) GetHTML(topic_id int) (string, error) {
	var htmlSource string
	err := s.db.QueryRow(`
		SELECT html_source FROM topics WHERE topic_id = ? ORDER BY DATE(created_at) desc limit 1
	`, topic_id).Scan(&htmlSource)
	if err != nil {
		return "", err
	}
	return htmlSource, nil
}

func (s *LiteStorage) SaveFetchPlan(plan TopicPlan) error {
	sql := "insert into topics(topic_id, created_at) values "

	values := make([]string, 0, len(plan))
	for k, v := range plan {
		new_value := fmt.Sprintf("(%d, '%s')", k, v)
		values = append(values, new_value)
	}
	values_sql := strings.Join(values, ",")

	stmt, err := s.db.Prepare(sql + values_sql)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec()
	return err
}
