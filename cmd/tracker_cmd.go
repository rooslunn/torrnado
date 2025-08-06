package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/rooslunn/torrnado"
	"golang.org/x/sync/errgroup"
)

type result struct {
	topic_id int
	Error    error
}

type trackerCmd struct{
	log *slog.Logger
}

const (
	DEFAULT_BATCH_SIZE = 100
)

const (
	SUBCMD_FETCH = "fetch"
)

func (c trackerCmd) execute(args []string) error {

	c.log.Info("executing fetch command", "args", args)

	if len(args) == 0 {
		c.expoTrick()
		return ErrDefectiveArgs
	}

	if isSubcommand(args[0], SUBCMD_FETCH) {
		return c.fetch(args[1:])
	}

	return ErrUnknownSubCommand
}

// [ ] todo: devide and conquer
// [ ] todo: return count of empty sources
// info: if length(html_source) < 14000, it is Тема не найдена
func (c trackerCmd) fetch(args []string) error {

	from_topic_id, err := strconv.Atoi(args[0])
	if err != nil {
		return ErrDefectiveArgs
	}

	topic_batch := DEFAULT_BATCH_SIZE
	if len(args) == 2 {
		topic_batch, err = strconv.Atoi(args[1])
		if err != nil {
			return ErrDefectiveArgs
		}
	}

	// config
	config, err := torrnado.MustConfig()
	if err != nil {
		c.log.Error(err.Error())
		return err
	}
	c.log.Info("config loaded", "url", config.Env[torrnado.TORR_URL], "db", config.Env[torrnado.TORR_DB])

	// db

	db, err := joinDb()
	if err != nil {
		c.log.Error(err.Error())
		return err
	}
	c.log.Info("DB joined", "status", db.Status)

	// topic source
	rt, err := torrnado.NeedSource(config)
	if err != nil {
		c.log.Error(err.Error())
		return err
	}
	c.log.Info("loggged to tracker", "status", rt.Status)

	// concurrent fetching

	maxConcurrency, err := strconv.Atoi(config.Env[torrnado.MAX_CONCURRENCY])
	if err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// handle ctrc+c/z (SIGINT, SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		sig := <-sigChan
		c.log.Info("Received signal: %v. Initiating graceful shutdown...\n", "signal", sig)
		cancel()
	}()

	// 5448425
	topics := torrnado.ConjureTopicPlan(from_topic_id, topic_batch)
	err = db.SaveFetchPlan(topics)
	if err != nil {
		c.log.Error("error saving fetching plan", "err", err)
		return err
	}
	c.log.Info("fetching plan is ready")

	g, gCtx := errgroup.WithContext(ctx)
	results := make(chan result, len(topics))
	sem := make(chan struct{}, maxConcurrency)

	url_fmt := config.Env[torrnado.TORR_URL]

	for topic_id := range topics {

		topic_id := topic_id

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				c.log.Info("cancellation detected, skipping...", "topic", topic_id)
				return gCtx.Err()
			case sem <- struct{}{}:
				defer func() { <-sem }()

				c.log.Info("fetching topic", "id", topic_id)
				startEra := time.Now()

				topicHtml, err := rt.FetchTopic(url_fmt, topic_id)

				if errors.Is(err, torrnado.ErrClaudfareWarden) {
					results <- result{topic_id, err}
					c.log.Warn("warden watching. time to sleep")
					time.Sleep(torrnado.AccidentalPeriodSec(13, 23))
					return nil
				}

				if err != nil {
					results <- result{topic_id, err}
					c.log.Error("fetching error", "err", err)
					return err
				}

				err = db.SaveEffort(topic_id, topicHtml, startEra)
				if err != nil {
					results <- result{topic_id, err}
					c.log.Error("fetching error", "err", err)
					return err
				}

				results <- result{topic_id, nil}
				c.log.Info("fetched successfully", "topic_id", topic_id)

				sleepFor := torrnado.AccidentalPeriodSec(3, 11)
				c.log.Info("sleeping for", "sec", sleepFor)
				time.Sleep(sleepFor)

				return nil
			}
		})
	}

	go func() {
		g.Wait()
		close(results)
	}()

	for result := range results {
		if result.Error != nil {
			c.log.Error("error", "topic_id", result.topic_id, "err", result.Error)
		}
	}

	if err := g.Wait(); err != nil {
		if err == context.Canceled {
			c.log.Info("graceful shutdown completed due to OS signal.")
		} else {
			c.log.Error("operation finished with a non-cancellation error:", "err", err)
		}
	} else {
		c.log.Info("All URLs fetched successfully!")
	}

	return nil
}

func (c trackerCmd) expoTrick() {
	fmt.Println("Usage: torrnado fetch <from topic_id> [count]")
}
