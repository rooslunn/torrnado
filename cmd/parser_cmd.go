package main

import (
	"fmt"
	"log/slog"
	"slices"
	"strconv"

	"github.com/rooslunn/torrnado"
)


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

	parser, err := torrnado.NewParser(c.log)
	if err != nil {
		return err
	}

	for topic_id := range topics {

		c.log.Info("fetch html_source", "op", op, "topic_id", topic_id)
		html_source, err := stash.GetHTML(topic_id)
		if err != nil {
			c.log.Error("fetch html_source", "op", op, "topic_id", topic_id, "err", err)
			return  err
		}

		c.log.Info("anatomize html_source", "op", op, "topic_id", topic_id)
		parsed, err := parser.Parse(html_source)
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
