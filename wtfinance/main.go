package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/NathanRJohnson/live-backend/wtfinance/application"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := application.LoadConfig()

	// TODO: add max idle connections via T
	app, err := application.New(ctx, cfg)
	if err != nil {
		log.Println("failed to initialize app:", err)
		return
	}

	err = app.Start(ctx)
	if err != nil {
		log.Println("failed to start app:", err)
	}
}
