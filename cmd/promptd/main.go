package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/redhotvengeance/promptd/internal/ipc"
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

	server := ipc.NewServer(filepath.Join(appDir, socketName))

	if err := server.Start(); err != nil {
		log.Fatalf("Server crashed: %v", err)
	}

	log.Println("Daemon shutting down...")
}
