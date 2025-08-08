package main

import (
	// "log/slog"
	"errors"
	"fmt"
	"log/slog"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/rooslunn/torrnado"
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

const (
	SUBCMD_PARSE = "parse"
)

type parserCmd struct {
	log *slog.Logger
}

func (c parserCmd) execute(args []string) error {
	c.log.Info("executing parser command")

	if len(args) == 0 {
		c.expoTrick()
		return ErrDefectiveArgs
	}

	if isSubcommand(args[0], SUBCMD_PARSE) {
		return c.parse(args[1:])
	}

	return nil
}

func (c parserCmd) parse(args []string) (error) {
	const op = "parse.main"

	dbpath, err := getDbPath()
	if err != nil {
		c.log.Error("get db path", "op", op, "err", err)
		return err
	}

	stash, err := torrnado.MustHaveStorage(dbpath)
	if err != nil {
		c.log.Error("connect to db", "op", op, "err", err)
		return  err
	}

	// topics range
	var topics torrnado.TopicPlan	

	if len(args) == 1 {
		from_topic_id, err := translateToTopicId(args[0])
		if err != nil {
			return err
		}
		topics = torrnado.ConjureTopicPlan(from_topic_id, 1)
	} else {
		topics, err = stash.AllFetchedRange()
		if err != nil {
			return err
		}
	}

	topicIndexes := torrnado.MapKeys(topics)
	c.log.Info("topics range to parse", "from", slices.Min(topicIndexes), "to", slices.Max(topicIndexes))

	for topic_id := range topics {

		c.log.Info("fetch html_source", "op", op, "topic_id", topic_id)
		html_source, err := stash.GetHTML(topic_id)
		if err != nil {
			c.log.Error("fetch html_source", "op", op, "topic_id", topic_id, "err", err)
			return  err
		}

		c.log.Info("anatomize html_source", "op", op, "topic_id", topic_id)
		parsed, err := c.anatomize(html_source)
		if err != nil {
			c.log.Error("anatomize html_source", "op", op, "topic_id", topic_id, "err", err)
			return err
		}

		c.log.Info("parsing done", "topic_id", topic_id, "title", parsed["title"])
	}

	return nil
}

func translateToTopicId(value string) (int, error) {
	topic_id, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return topic_id, nil
}

func (c parserCmd) expoTrick() {
	fmt.Println("Usage: torrnado parser:<subcommand>")
	fmt.Println("Usage: torrnado parser:parse [topic_id]")
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
	topicFieldsIndex = torrnado.MapKeys(topicFieldsTranslation)

)

func (c parserCmd) anatomize(html_source string) (parsed_html, error) {
	const op = "parser.anatomize"

	topic := make(parsed_html, 14)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html_source))
	if err != nil {
		c.log.Error("creating goquery from html_source", "op", op, "err", err)
		return topic, err
	}

	post_body := doc.Find("div.post_wrap > div.post_body span.post-align")
	if len(post_body.Nodes) == 0 {
		c.log.Error("can't find post_body in html_source", "op", op)
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