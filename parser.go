package torrnado

import (
	"errors"
	"log/slog"
	"regexp"
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

type parsed_html = map[string]any

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

func (p *parser) anatomize(html_source string) (parsed_html, error) {
	const op = "parser.anatomize"

	topic := make(parsed_html, 14)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html_source))
	if err != nil {
		p.log.Error("creating goquery from html_source", "op", op, "err", err)
		return topic, err
	}

	description := doc.Find("head > meta[name=\"description\"]").AttrOr("content", "")
	topic["description"] = description
	topic["title"] = extractTitle(description)
	topic["audio"] = extractAudio(description)
	topic["category"] = extractCategory(description)
	topic["status"] = "status"

	details := doc.Find("span.post-b")
	details.Each(func(i int, e *goquery.Selection) {
		field := freeFromDebris(e.Text())
		if strings.HasPrefix(field, "Жанр"){
			topic["genre"] = freeFromDebris(e.Nodes[0].NextSibling.Data)
			return
		}
	})

	details = doc.Find("table#t-tor-stats tbody tr > td b")
	if len(details.Nodes) > 2 {
		topic["likes"] = removeNonNumeric(details.Nodes[2].FirstChild.Data)
	}

	details = doc.Find("p.nick.nick-author")
	topic["author"] = details.First().Text()

	// [ ] todo: check validity
	details = doc.Find("a.magnet-link")
	topic["magnet_link"] = details.First().AttrOr("href", "")

	return topic, nil
}

func extractCategory(description string) any {
	category := ""

	start_bit := strings.Index(description, "»")
	stop_bit := strings.Index(description, "::")
	if start_bit < 0 || stop_bit < 0 {
		return "extractCategory: can't pos on start or stop bits"
	}

	category = description[start_bit+2 : stop_bit] 

	return freeFromDebris(category)
}

func extractAudio(description string) any {
	audio := make([]string, 0)

	start_bit := strings.IndexRune(description, ']')
	stop_bit := strings.Index(description, " »")
	if start_bit < 0 || stop_bit < 0 {
		return []string{"extractAudio: can't pos on start or stop bits"}
	}

	audio_secion := description[start_bit+1:stop_bit]

	parts := strings.Split(audio_secion, " + ")
	
	for len(parts) != 0 {
		audio = append(audio, freeFromDebris(parts[0]))
		parts = slices.Delete(parts, 0, 1)
	}

	return audio
}

func extractTitle(description string) []string {
	lang := make([]string, 0)

	stop_bit := strings.IndexRune(description, '(')
	title_secion := description[:stop_bit]

	parts := strings.Split(title_secion, " / ")
	
	for len(parts) != 0 {
		lang = append(lang, freeFromDebris(parts[0]))
		parts = slices.Delete(parts, 0, 1)
	}

	return lang
}

const FIELD_DEBRIS = " :\n+"

func freeFromDebris(s string) string {
	return strings.Trim(s, FIELD_DEBRIS)
}

func removeNonNumeric(s string) string {
	reg := regexp.MustCompile("[^0-9]+")
	return reg.ReplaceAllString(s, "")
}