package torrnado

import (
	"errors"
	"log/slog"
	"slices"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"golang.org/x/net/html"
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

type fieldCareFunc = func(node *html.Node) (key string, value string)

const SEP_DOG = " @ "

var translationCareFunc = func(node *html.Node) (key string, value string) {
	i := 1
	crashStop := 6

	translations := make([]string, 0)

	currNode := node.NextSibling

	for (i < crashStop) && (currNode.Data != "span") {
		if currNode.Type == html.TextNode {
			trans := freeFromDebris(currNode.Data)
			translations = append(translations, trans)
		}
		currNode = currNode.NextSibling
		i++
	}

	return "translation", strings.Join(translations, SEP_DOG) 
}

var (
	topicFieldsTranslation = map[string]string{
		"Формат": "format", "Качество":"quality", "Видео":"video", "Перевод:":"",
		"Аудио":"audio", "Аудио #1":"audio.1", "Аудио #2":"audio.2", "Аудио #3":"audio.3",
		"Субтитры": "subtitles", 
	}
	topicFieldsIndex = MapKeys(topicFieldsTranslation)

	topicSpecialCareFields = map[string]fieldCareFunc{
		"Перевод": translationCareFunc,
	}
	topicSpecialCareFieldsIndex = MapKeys(topicSpecialCareFields)
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
			topic[topicFieldsTranslation[field]] = freeFromDebris(e.Nodes[0].NextSibling.Data)
			return
		}
		if slices.Contains(topicSpecialCareFieldsIndex, field) {
			k, v := topicSpecialCareFields[field](e.Nodes[0])
			topic[k] = v
		}
	})

	details = doc.Find("table#t-tor-stats tbody tr > td b")
	if len(details.Nodes) > 2 {
		topic["likes"] = details.Nodes[2].FirstChild.Data
	}

	details = doc.Find("p.nick.nick-author")
	topic["author"] = details.First().Text()

	// [ ] todo: check validity
	details = doc.Find("a.magnet-link")
	topic["magnet_link"] = details.First().AttrOr("href", "")

	return topic, nil
}

const FIELD_DEBRIS = " :\n+"

func freeFromDebris(s string) string {
	return strings.Trim(s, FIELD_DEBRIS)
}