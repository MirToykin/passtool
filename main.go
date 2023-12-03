package main

import (
	"github.com/MirToykin/passtool/cmd"
	"github.com/MirToykin/passtool/internal/storage/sqlite"
	"log"
)

const storagePath = "storage.db"

func main() {
	db, err := sqlite.New(storagePath)
	if err != nil {
		log.Fatalf("unable to initialize DB: %v", err)
	}
	cmd.Execute(db)
}
