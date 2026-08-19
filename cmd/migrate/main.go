package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"

	"devclub.com/identity/internal/api/config"
	"devclub.com/identity/pkg/utils"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	infoLogger, _, errorLogger := utils.CreateLoggers()

	cfg, err := config.NewConfig()
	if err != nil {
		errorLogger.Error("failed to load configuration", "error", err)
		os.Exit(1)
	}

	migrationsPath := flag.String("path", "file://migrations", "Path to migration files (must prefix with file://)")
	flag.Parse()

	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: go run cmd/migrate/main.go [options] <command> [args]")
		fmt.Println("\nCommands:")
		fmt.Println("  up             Apply all pending migrations")
		fmt.Println("  down           Roll back all migrations")
		fmt.Println("  step <n>       Apply or rollback n migrations (e.g. step 1, step -1)")
		fmt.Println("  force <v>      Force clean dirty database migration version")
		fmt.Println("  version        Print current migration version")
		os.Exit(1)
	}

	cmd := args[0]

	m, err := migrate.New(*migrationsPath, cfg.DBConn)
	if err != nil {
		errorLogger.Error("failed to initialize migrate driver", "error", err)
		os.Exit(1)
	}
	defer func() {
		srcErr, dbErr := m.Close()
		if srcErr != nil {
			errorLogger.Error("migration source close error", "error", srcErr)
		}
		if dbErr != nil {
			errorLogger.Error("migration database close error", "error", dbErr)
		}
	}()

	switch cmd {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				infoLogger.Info("no new migrations to apply")
				return
			}
			errorLogger.Error("migration up failed", "error", err)
			os.Exit(1)
		}
		infoLogger.Info("migrations applied successfully (up)")

	case "down":
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				infoLogger.Info("database is already at initial state")
				return
			}
			errorLogger.Error("migration down failed", "error", err)
			os.Exit(1)
		}
		infoLogger.Info("migrations rolled back successfully (down)")

	case "step":
		if len(args) < 2 {
			errorLogger.Error("step command requires a step count (e.g., 'step 1' or 'step -1')")
			os.Exit(1)
		}
		steps, err := strconv.Atoi(args[1])
		if err != nil {
			errorLogger.Error("invalid step number", "error", err)
			os.Exit(1)
		}
		if err := m.Steps(steps); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				infoLogger.Info("no changes applied")
				return
			}
			errorLogger.Error("migration step failed", "error", err)
			os.Exit(1)
		}
		infoLogger.Info("migration steps executed successfully", "steps", steps)

	case "force":
		if len(args) < 2 {
			errorLogger.Error("force command requires a target version number")
			os.Exit(1)
		}
		v, err := strconv.Atoi(args[1])
		if err != nil {
			errorLogger.Error("invalid version number", "error", err)
			os.Exit(1)
		}
		if err := m.Force(v); err != nil {
			errorLogger.Error("failed to force migration version", "error", err)
			os.Exit(1)
		}
		infoLogger.Info("migration version forced successfully", "version", v)

	case "version":
		version, dirty, err := m.Version()
		if err != nil {
			if errors.Is(err, migrate.ErrNilVersion) {
				infoLogger.Info("no migrations applied yet")
				return
			}
			errorLogger.Error("failed to get migration version", "error", err)
			os.Exit(1)
		}
		infoLogger.Info("current database schema version", "version", version, "dirty", dirty)

	default:
		errorLogger.Error("unknown command", "command", cmd)
		os.Exit(1)
	}
}
