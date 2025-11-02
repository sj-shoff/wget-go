package main

import (
	"log"
	"os"
	"wget-go/internal/app"
)

func main() {
	application := app.New()
	if err := application.Run(); err != nil {
		log.Fatalf(
			"Application failed: %s\n",
			err,
		)
		os.Exit(1)
	}
}
