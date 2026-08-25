package main

import (
	"errors"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: go run ./cmd/migrate <up|down|version> [steps]")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}

	m, err := migrate.New("file://migrations", databaseURL)
	if err != nil {
		log.Fatalf("create migrator: %v", err)
	}
	defer func() {
		sourceErr, databaseErr := m.Close()
		if sourceErr != nil || databaseErr != nil {
			log.Printf("close migrator: source=%v database=%v", sourceErr, databaseErr)
		}
	}()

	switch os.Args[1] {
	case "up":
		runMigration(m.Up())
	case "down":
		steps := 1
		if len(os.Args) >= 3 {
			steps, err = strconv.Atoi(os.Args[2])
			if err != nil || steps < 1 {
				log.Fatal("down steps must be a positive integer")
			}
		}
		runMigration(m.Steps(-steps))
	case "version":
		version, dirty, err := m.Version()
		if errors.Is(err, migrate.ErrNilVersion) {
			fmt.Println("version: none")
			return
		}
		if err != nil {
			log.Fatalf("read migration version: %v", err)
		}
		fmt.Printf("version: %d (dirty: %t)\n", version, dirty)
	default:
		log.Fatalf("unknown command %q; use up, down, or version", os.Args[1])
	}
}

func runMigration(err error) {
	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("no migration changes")
		return
	}
	if err != nil {
		log.Fatalf("migration failed: %v", err)
	}
	fmt.Println("migration completed")
}
