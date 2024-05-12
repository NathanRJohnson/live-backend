package application

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"cloud.google.com/go/firestore"
)

type App struct {
	router http.Handler
	fdb    *firestore.Client
	config Config
}

func New(ctx context.Context, cfg Config) (*App, error) {
	client, err := firestore.NewClient(ctx, cfg.Secrets.ProjectID)
	if err != nil {
		fmt.Println("failed to connect to firebase client", err)
		return nil, err
	}

	app := &App{
		fdb:    client,
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

	defer func() {
		if err := a.fdb.Close(); err != nil {
			fmt.Println("failed to close firebase", err)
		}
	}()

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
