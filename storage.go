package torrnado

import (
	"database/sql"
	"errors"
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
	path   string
	Status string
}

func (s *LiteStorage) AllFetchedRange() (TopicPlan, error) {
	plan := make(TopicPlan)

	query := `
		select distinct topic_id 
		from topics 
		where 
			fetched_at is not null and html_source is not null;
	`;
	var (
		topic_id int
	)

	rows, err := s.db.Query(query)
	if err != nil {
		return plan, err
	}
	defer rows.Close()

	for rows.Next() {
		err = rows.Scan(&topic_id)
		if err != nil {
			return plan, err
		}
		plan[topic_id] = ISaidNow()
	}

	return plan, rows.Err()
}

const DB_OPTIONS = "?cache=shared"

func MustHaveStorage(path string) (*LiteStorage, error) {
	const op = "storage.sqlite.connect"

	db, err := sql.Open("sqlite3", path+DB_OPTIONS)
	if err != nil {
		return nil, Operror(op, err)
	}

	return &LiteStorage{db, path, "ready"}, nil
}

func (s *LiteStorage) Migrate(log *slog.Logger) error {

	const op = "storage.migrate"

	var ddl []string

	// drops
	ddl = append(ddl, `
		drop view if exists topics_debug;
		drop table if exists topics;
	`)

	// tables
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

	// indexes
	ddl = append(ddl, `
		create index if not exists idx_topic_id on topics(topic_id);
		create index if not exists idx_html_len on topics(html_len);
	`)

	// views
	ddl = append(ddl, `
		CREATE VIEW topics_debug AS 
		select 
			id, topic_id, substr(html_source, 1, 32) as html_spot, html_len, time_taken_ms, 
			created_at, fetched_at 
		from topics;
	`)

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
		log.Info("shooting", "ddl", ddl_sql[:32])
	}
	defer stmt.Close()

	return nil

}

func (s *LiteStorage) Hygienic(log *slog.Logger) (int, error) {

	log.Info("started cleaning")

	rows_affected, err := s.execSQL(`
		delete from topics 
		where fetched_at < (
			select max(fetched_at) as last_fetched_at 
			from topics 
			group by topic_id 
			having last_fetched_at is not null 
			order by last_fetched_at limit 1
		) 
			or fetched_at is null;
	`)

	return int(rows_affected), err
}

func (s *LiteStorage) SaveEffort(topic_id int, html string, timeLog time.Time) error {
	stmt, err := s.db.Prepare(`
		update topics 
			set html_source = $1, html_len = $2, fetched_at = $3, time_taken_ms = $4
		where 
			topic_id = $5 and fetched_at is null;
	`)
	if err != nil {
		return err
	}
	defer stmt.Close()

	time_taken := time.Since(timeLog).Milliseconds()

	_, err = stmt.Exec(html, len(html), ISaidNow(), time_taken, topic_id)

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

func (s *LiteStorage) retrieveField(query string, params ...any) (string, error) {
	var value string

	err := s.db.QueryRow(query, params...).Scan(&value)

	if errors.Is(sql.ErrNoRows, err) {
		return "", ErrSelectNoRows
	}

	return value, nil
}

var (
	ErrSelectNoRows = errors.New("select returned no rows")
)

func (s *LiteStorage) GetHTML(topic_id int) (string, error) {
	query := `
		SELECT html_source 
		FROM topics 
		WHERE topic_id = ? and fetched_at is not null 
		ORDER BY fetched_at desc limit 1
	`
	return s.retrieveField(query, fmt.Sprint(topic_id))
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
