package main

import (
	"os"

	"github.com/cmj0121/pwnpi"
)

func main() {
	agent := pwnpi.New()
	os.Exit(agent.ParseAndRun())
}
