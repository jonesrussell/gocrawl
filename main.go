package main

import (
	"github.com/jonesrussell/gocrawl/cmd"
)

func main() {
	// Call the Execute function from the cmd package
	if err := cmd.Execute(); err != nil {
		// Handle any errors that occur during command execution
		panic(err)
	}
}
