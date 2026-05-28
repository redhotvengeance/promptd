package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/redhotvengeance/promptd/internal/config"
	"github.com/redhotvengeance/promptd/internal/generation"
	"github.com/redhotvengeance/promptd/internal/ipc/router"
	"github.com/redhotvengeance/promptd/internal/ipc/server"
	"github.com/redhotvengeance/promptd/internal/libsql"
	"github.com/redhotvengeance/promptd/internal/llm"
	"github.com/redhotvengeance/promptd/internal/workspace"
)

const (
	socketDir = "lmp"
	socketName = "promptd.sock"
)

func main() {
	log.Println("Daemon booting...")

	var baseDir string
	if xdg := os.Getenv("XDG_RUNTIME_DIR"); xdg != "" {
		baseDir = xdg
	} else {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			baseDir = os.TempDir()
		} else {
			baseDir = filepath.Join(homeDir, ".local", "state")
		}
	}

	appDir := filepath.Join(baseDir, socketDir)

	if err := os.MkdirAll(appDir, 0700); err != nil {
		log.Fatalf("Could not create socket directory at %s: %v", appDir, err)
	}

	config, err := config.LoadConfig()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	datastore := libsql.NewDatastore()
	if err := datastore.Open(); err != nil {
		fmt.Fprintf(os.Stderr, "failed to open db: %s", err)
		os.Exit(1)
	}

	llmManager := llm.NewManager(&config)

	workspaceService := workspace.NewService(llmManager, datastore)

	genService := generation.NewService(llmManager, datastore, workspaceService)

	router := router.NewRouter(genService, datastore, workspaceService)

	server := ipc.NewServer(filepath.Join(appDir, socketName), router.Handle)
	if err := server.Start(); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}

	log.Println("Daemon shutting down...")
}
