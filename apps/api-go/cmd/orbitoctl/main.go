package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/orbito/api-go/internal/config"
	"github.com/orbito/api-go/internal/platform/database"
)

const version = "1.0.0"

func loadDotEnv() {
	paths := []string{".env", "../../backend/.env", "../backend/.env", "backend/.env"}
	for _, p := range paths {
		cleaned := filepath.Clean(p)
		data, err := os.ReadFile(cleaned)
		if err != nil {
			continue
		}
		lines := strings.Split(string(data), "\n")
		for _, line := range lines {
			line = strings.TrimSpace(line)
			if line == "" || strings.HasPrefix(line, "#") {
				continue
			}
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}
			key := strings.TrimSpace(parts[0])
			val := strings.Trim(strings.TrimSpace(parts[1]), `"'`)
			if os.Getenv(key) == "" {
				_ = os.Setenv(key, val)
			}
		}
		break
	}
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(0)
	}

	command := os.Args[1]
	switch command {
	case "version":
		fmt.Printf("orbitoctl version %s\n", version)
	case "db":
		handleDBCommand(os.Args[2:])
	default:
		fmt.Fprintf(os.Stderr, "Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Orbito Platform CLI Tooling (orbitoctl)")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  orbitoctl <command> [arguments]")
	fmt.Println()
	fmt.Println("Available Commands:")
	fmt.Println("  version          Print orbitoctl version")
	fmt.Println("  db ping          Verify PostgreSQL database connectivity")
	fmt.Println("  db status        Display database connection status and table statistics")
	fmt.Println()
}

func handleDBCommand(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: orbitoctl db [ping|status]")
		os.Exit(1)
	}

	loadDotEnv()
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	if cfg.DatabaseURL == "" {
		fmt.Fprintln(os.Stderr, "DATABASE_URL is not set")
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.New(ctx, cfg.DatabaseURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Database connection error: %v\n", err)
		os.Exit(1)
	}
	defer db.Close()

	sub := args[0]
	switch sub {
	case "ping":
		if err := db.Ping(ctx); err != nil {
			fmt.Fprintf(os.Stderr, "Database ping failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Database connection OK (ping succeeded)")
	case "status":
		stat := db.Pool().Stat()
		fmt.Println("--- Database Status ---")
		fmt.Printf("Total Connections: %d\n", stat.TotalConns())
		fmt.Printf("Idle Connections:  %d\n", stat.IdleConns())
		fmt.Printf("Acquired Conns:    %d\n", stat.AcquiredConns())
		fmt.Printf("Max Connections:   %d\n", stat.MaxConns())

		var userCount, orgCount, wsCount, projCount, wfCount, taskCount int64
		_ = db.Pool().QueryRow(ctx, `SELECT count(*) FROM "users"`).Scan(&userCount)
		_ = db.Pool().QueryRow(ctx, `SELECT count(*) FROM "organizations"`).Scan(&orgCount)
		_ = db.Pool().QueryRow(ctx, `SELECT count(*) FROM "workspaces"`).Scan(&wsCount)
		_ = db.Pool().QueryRow(ctx, `SELECT count(*) FROM "projects"`).Scan(&projCount)
		_ = db.Pool().QueryRow(ctx, `SELECT count(*) FROM "workflows"`).Scan(&wfCount)
		_ = db.Pool().QueryRow(ctx, `SELECT count(*) FROM "tasks"`).Scan(&taskCount)

		fmt.Println("--- Schema Table Counts ---")
		fmt.Printf("Users:         %d\n", userCount)
		fmt.Printf("Organizations: %d\n", orgCount)
		fmt.Printf("Workspaces:    %d\n", wsCount)
		fmt.Printf("Projects:      %d\n", projCount)
		fmt.Printf("Workflows:     %d\n", wfCount)
		fmt.Printf("Tasks:         %d\n", taskCount)
	default:
		fmt.Fprintf(os.Stderr, "Unknown db subcommand: %s\n", sub)
		os.Exit(1)
	}
}
