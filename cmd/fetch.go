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

type fetchCmd struct{}

const (
	OUT_PATH            = ""
	DEFAULT_BATCH_SIZE = 100
)

// [ ] todo: devide and conquer
func (c fetchCmd) execute(log *slog.Logger, args []string) error {
	log.Info("executing fetch command", "args", args)

	if len(args) == 0 {
		c.expoTrick()
		return ErrDefectiveArgs
	}

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
		log.Error(err.Error())
		return err
	}
	log.Info("config loaded", "url", config.Env[torrnado.TORR_URL], "db", config.Env[torrnado.TORR_DB])

	// db
	db, err := torrnado.MustSaveToLite(config.Env[torrnado.TORR_DB])
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("DB connected", "status", db.Status)

	// topic source
	rt, err := torrnado.NeedSource(config)
	if err != nil {
		log.Error(err.Error())
		return err
	}
	log.Info("loggged to tracker", "status", rt.Status)

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
		log.Info("Received signal: %v. Initiating graceful shutdown...\n", "signal", sig)
		cancel()
	}()

	// 5448425
	topics := torrnado.ConjureTopicPlan(from_topic_id, topic_batch)
	err = db.SaveFetchPlan(topics)
	if err != nil {
		log.Error("error saving fetching plan", "err", err)
		return err
	}
	log.Info("fetching plan is ready")

	g, gCtx := errgroup.WithContext(ctx)
	results := make(chan result, len(topics))
	sem := make(chan struct{}, maxConcurrency)

	url_fmt := config.Env[torrnado.TORR_URL]

	for topic_id := range topics {

		topic_id := topic_id

		g.Go(func() error {
			select {
			case <-gCtx.Done():
				log.Info("cancellation detected, skipping...", "topic", topic_id)
				return gCtx.Err()
			case sem <- struct{}{}:
				defer func() { <-sem }()

				log.Info("fetching topic", "id", topic_id)
				startEra := time.Now()

				topicHtml, err := rt.FetchTopic(url_fmt, topic_id)

				if errors.Is(err, torrnado.ErrClaudfareWarden) {
					results <- result{topic_id, err}
					log.Warn("warden watching. time to sleep")
					time.Sleep(torrnado.AccidentalPeriodSec(13, 23))
					return nil
				}

				if err != nil {
					results <- result{topic_id, err}
					log.Error("fetching error", "err", err)
					return err
				}

				err = db.SaveEffort(topic_id, topicHtml, startEra)
				if err != nil {
					results <- result{topic_id, err}
					log.Error("fetching error", "err", err)
					return err
				}

				results <- result{topic_id, nil}
				log.Info("fetched successfully", "topic_id", topic_id)

				sleepFor := torrnado.AccidentalPeriodSec(3, 11)
				log.Info("sleeping for", "sec", sleepFor)
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
			log.Error("error", "topic_id", result.topic_id, "err", result.Error)
		}
	}

	if err := g.Wait(); err != nil {
		if err == context.Canceled {
			log.Info("graceful shutdown completed due to OS signal.")
		} else {
			log.Error("operation finished with a non-cancellation error:", "err", err)
		}
	} else {
		log.Info("All URLs fetched successfully!")
	}

	return nil
}

func (c fetchCmd) expoTrick() {
	fmt.Println("Usage: torrnado fetch <from topic_id> [count]")
}