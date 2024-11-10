package application

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type App struct {
	router http.Handler
	gss    *sheets.Service
	config Config
}

func New(ctx context.Context, cfg Config) (*App, error) {
	// Initialize the Sheets API client
	service, err := sheets.NewService(ctx, option.WithCredentialsJSON(cfg.ServiceKey))
	log.Println("Service initialized: ", service.Spreadsheets.Sheets)
	if err != nil {
		log.Fatalf("Unable to create Sheets service: %v", err)
	}

	app := &App{
		gss:    service,
		config: cfg,
	}
	app.loadRoutes()

	return app, nil
}

func (a *App) Start(ctx context.Context) error {
	server := &http.Server{
		Addr:        fmt.Sprintf(":%d", a.config.ServerPort),
		Handler:     a.router,
		IdleTimeout: 30 * time.Second,
	}

	fmt.Println("Starting server")

	ch := make(chan error, 1)

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			ch <- fmt.Errorf("failed to start server: %w", err)
		}
		close(ch)
	}()

	select {
	case err := <-ch:
		return err

	case <-ctx.Done():
		timeout, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		return server.Shutdown(timeout)
	}

	return nil
}
