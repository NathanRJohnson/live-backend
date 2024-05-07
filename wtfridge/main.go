package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	application "github.com/NathanRJohnson/live-backend/wtfridge/application"
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	cfg := application.LoadConfig()

	// TODO: add max idle connections via T
	app, err := application.New(ctx, cfg)
	if err != nil {
		fmt.Println("failed to initialize app:", err)
		return
	}

	err = app.Start(ctx)
	if err != nil {
		fmt.Println("failed to start app:", err)
	}

	// TODO: add middleware logging
	// TODO: check connection to FB
}
