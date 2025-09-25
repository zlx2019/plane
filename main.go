package main

import (
	"log/slog"
	"os"
	"plane/cmd"
)

// TCP local forward
func main() {
	if err := cmd.Run(); err != nil {
		slog.Error("Startup error: " + err.Error())
		os.Exit(1)
	}
}
