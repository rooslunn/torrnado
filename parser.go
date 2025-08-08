package torrnado

import (
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type parsed struct {
	topic_id int
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

type parsed_html map[string]string

// [ ] https://rutracker.org/forum/viewtopic.php?t=5104420 голгофа
// [ ] https://rutracker.org/forum/viewtopic.php?t=4624488 семь психопатов
// [ ] https://rutracker.org/forum/viewtopic.php?t=5352911 война против всех
// [ ] https://rutracker.org/forum/viewtopic.php?t=6238349 пройщенный
// [ ] https://rutracker.org/forum/viewtopic.php?t=5527258 шестизарядник
// [ ] https://rutracker.org/forum/viewtopic.php?t=3201064 альмодовар (сборник фильмов)

type parser struct {
	log *slog.Logger
}

func NewParser(log *slog.Logger) (*parser, error) {
	return &parser{log}, nil
}

func (p *parser) Parse(html_source string) (parsed_html, error) {
	return p.anatomize(html_source)
}

var (
	ErrIndexNotFound = errors.New("index not found")
	ErrNoBodyInSource = errors.New("can't point out post_body in source")
)

var (
	topicFieldsTranslation = map[string]string{
		"Формат": "format", "Качество":"quality", "Видео:":"video", "Перевод:":"",
		"Аудио":"audio", "Аудио #1":"audio.1", "Аудио #2":"audio.2", "Аудио #3":"audio.3",
	}
	topicFieldsIndex = MapKeys(topicFieldsTranslation)
)

func (p *parser) anatomize(html_source string) (parsed_html, error) {
	const op = "parser.anatomize"

	topic := make(parsed_html, 14)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html_source))
	if err != nil {
		p.log.Error("creating goquery from html_source", "op", op, "err", err)
		return topic, err
	}

	post_body := doc.Find("div.post_wrap > div.post_body span.post-align")
	if len(post_body.Nodes) == 0 {
		p.log.Error("can't find post_body in html_source", "op", op)
		return topic, ErrNoBodyInSource
	}

	topic["title"] = freeFromDebris(post_body.Nodes[0].FirstChild.Data)

	details := doc.Find("span.post-b")
	details.Each(func(i int, e *goquery.Selection) {
		field := freeFromDebris(e.Text())
		if slices.Contains(topicFieldsIndex, field) {
			topic[topicFieldsTranslation[field]] = strings.Trim(e.Nodes[0].NextSibling.Data, ": ")
			return
		}
	})

	return topic, nil
}

const FIELD_DEBRIS = " :"

func freeFromDebris(s string) string {
	return strings.Trim(s, FIELD_DEBRIS)
}