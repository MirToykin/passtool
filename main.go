package main

import (
	"github.com/MirToykin/passtool/cmd"
	"github.com/MirToykin/passtool/internal/config"
	out "github.com/MirToykin/passtool/internal/output"
	"github.com/MirToykin/passtool/internal/storage"
	"os"
)

func main() {
	prt := out.New()
	cfg := config.Load()
	if !cfg.IsValid() {
		cmd.PrintServiceRequirements(cfg, prt)
		os.Exit(0)
	}

	db, err := storage.New(cfg.StoragePath)
	if err != nil {
		prt.ErrorWithExit("unable to initialize DB: %v", err)
	}
	cmd.Execute(db, cfg, prt)
}
