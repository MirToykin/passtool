package main

import (
	"github.com/MirToykin/passtool/cmd"
	"github.com/MirToykin/passtool/internal/config"
	"github.com/MirToykin/passtool/internal/storage"
	"log"
	"os"
)

func main() {
	cfg := config.Load()
	if !cfg.IsValid() {
		cmd.PrintServiceRequirements(cfg)
		os.Exit(0)
	}

	db, err := storage.New(cfg.StoragePath)
	if err != nil {
		log.Fatalf("unable to initialize DB: %v", err)
	}
	cmd.Execute(db, cfg)
}
