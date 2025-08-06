package torrnado

import (
	"database/sql"
	"fmt"
	"log/slog"
	"strings"
	"time"

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

func MustHaveStorage(path string) (*LiteStorage, error) {
	const op = "storage.sqlite.connect"

	db, err := sql.Open("sqlite3", path+DB_OPTIONS)
	if err != nil {
		return nil, Operror(op, err)
	}

	return &LiteStorage{db, "ready"}, nil
}

func (s *LiteStorage) Migrate(log *slog.Logger) error {

	const op = "storage.migrate"

	var ddl []string
	
	ddl = append(ddl, "drop table if exists topics")

	ddl = append(ddl, `
		create table if not exists topics (
			id integer primary key,
			topic_id integer not null,
			html_source text,
			html_len int,
			time_taken_ms int,
			created_at text,
			fetched_at text 
		);
	`)

	ddl = append(ddl, "create index idx_topic_id on topics(topic_id);")
	ddl = append(ddl, "create index idx_html_len on topics(html_len);")

	var stmt *sql.Stmt
	var err error

	for _, ddl_sql := range ddl {
		stmt, err = s.db.Prepare(ddl_sql)
		if err != nil {
			return Operror(op, err)
		}
		_, err = stmt.Exec()
		if err != nil {
			return Operror(op, err)
		}
	}
	defer stmt.Close()

	return nil

}

func (s *LiteStorage) Hygienic(log *slog.Logger) (int, error) {

	log.Info("started cleaning")

	rows_affected, err := s.execSQL(`
		delete from topics where fetched_at < (select max(fetched_at) as last_fetched_at from topics group by topic_id order by last_fetched_at limit 1);
	`)

	return int(rows_affected), err 
}

func (s *LiteStorage) SaveEffort(topic_id int, html string, timeLog time.Time) error {
	stmt, err := s.db.Prepare(`
		update 
			topics set html_source = $1, fetched_at = $2, time_taken_ms = $3
		where 
			topic_id = $4 and fetched_at is null;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	time_taken := time.Since(timeLog).Milliseconds()

	_, err = stmt.Exec(html, ISaidNow(), time_taken, topic_id)

	return err
}

func (s *LiteStorage) execSQL(sql string) (changes int64, err error) {
	stmt, err := s.db.Prepare(sql)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	r, err := stmt.Exec()

	rows_affected, err := r.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rows_affected, err
}

func (s *LiteStorage) retrieveField(sql string) (string, error) {
	var value string
	err := s.db.QueryRow(sql).Scan(&value)
	if err != nil {
		return "", err
	}
	return value, nil
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
