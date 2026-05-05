package main

import (
	"log"

	"github.com/anomalyco/vagai-cli/cmd/root"
	"github.com/anomalyco/vagai-cli/internal/db"
)

func main() {
	if err := db.Init(); err != nil {
		log.Printf("Aviso: banco não conectado: %v", err)
	}

	root.Execute()
}
