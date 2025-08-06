package main

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/rooslunn/torrnado"
)

type dbCmd struct {
	log *slog.Logger
}

const (
	SUBCMD_EXPORT  = "export"
	SUBCMD_CLEAN   = "organize"
	SUBCMD_MIGRATE = "invent"
)

func (c dbCmd) execute(args []string) error {
	c.log.Info("executing db command", "args", args)

	if len(args) == 0 {
		c.expoTrick()
		return ErrDefectiveArgs
	}

	if isSubcommand(args[0], SUBCMD_CLEAN) {
		return c.clean(args[1:])
	} else if isSubcommand(args[0], SUBCMD_EXPORT) {
		return c.export(args[1:])
	} else if isSubcommand(args[0], SUBCMD_MIGRATE) {
		return c.migrate()
	}

	return ErrUnknownSubCommand
}

func exportTricks() {
	fmt.Println("Usage: export <topic_id>")
}

func (c dbCmd) migrate() error {
	c.log.Info("executing migrate sub command")

	// remove prev file
	dbpath, err := getDbPath()
	if err != nil {
		return err
	}

	c.log.Info("expel prev db file")
	err = os.Remove(dbpath)
	if err != nil {
		if os.IsNotExist(err) {
			c.log.Info("file doen't live", "path", dbpath)
		} else {
			return err
		}
	}


	db, err := joinDb()
	if err != nil {
		return err
	}
	c.log.Info("DB joined")

	err = db.Migrate(c.log)
	if err != nil {
		return err
	}
	c.log.Info("DB migrated")

	return nil
}

func getDbPath() (string, error) {
	dbpath := os.Getenv(torrnado.TORR_DB)
	if dbpath == "" {
		return "", fmt.Errorf(torrnado.ErrEnvNotSet, torrnado.TORR_DB)
	}
	return dbpath, nil
}

func joinDb() (*torrnado.LiteStorage, error) {
	dbpath, err := getDbPath()
	if err != nil {
		return nil, fmt.Errorf(torrnado.ErrEnvNotSet, err)
	}

	db, err := torrnado.MustHaveStorage(dbpath)
	if err != nil {
		return nil, err
	}

	return db, nil
}

func (c *dbCmd) export(args []string) error {
	c.log.Info("executing export sub command", "args", args)

	if len(args) == 0 {
		exportTricks()
		return ErrDefectiveArgs
	}

	topic_id, err := strconv.Atoi(args[0])
	if err != nil {
		exportTricks()
		return ErrDefectiveArgs
	}

	db, err := joinDb()
	if err != nil {
		return err
	}

	html, err := db.GetHTML(topic_id)
	if err != nil {
		return err
	}
	c.log.Info("obtained html for", "topic_id", topic_id)

	filename := fmt.Sprintf("%d.html", topic_id)
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error creating file: %v", err)
	}
	defer file.Close()

	_, err = file.WriteString(html)
	if err != nil {
		return fmt.Errorf("error writing to file: %v", err)
	}
	c.log.Info("deliverd html to file", "file", filename)

	return nil
}

func (c dbCmd) clean(args []string) error {
	c.log.Info("executing clean sub command", "args", args)

	db, err := joinDb()
	if err != nil {
		return err
	}

	records_removed, err := db.Hygienic(c.log)
	if err != nil {
		return err
	}
	c.log.Info("cleared", "records evacuated", records_removed)

	return nil
}

func (c dbCmd) expoTrick() {
	fmt.Println("Usage: torrnado export <topic_id> <filepath>")
}
