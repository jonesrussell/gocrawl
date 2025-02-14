package main

import (
	"log"

	"github.com/jonesrussell/gocrawl/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatal("Error executing command: ", err)
	}
}
