package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/dennypenta/go-api-walkthrough/assembly"
	_ "go.uber.org/automaxprocs"
	"golang.org/x/sync/errgroup"
)

func main() {
	conf, err := assembly.NewConfig()
	if err != nil {
		log.Fatalln("failed to load config:", err)
	}

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	app, err := assembly.NewApp(conf)
	if err != nil {
		log.Fatalln("failed to create app:", err)
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	server := &http.Server{Addr: ":" + conf.HttpPort, Handler: app.Mux}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancel()

		app.Log.InfoContext(ctx, "shutting down server")
		app.Close(ctx)
		return server.Shutdown(context.Background())
	})
	g.Go(func() error {
		if err := server.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				app.Log.InfoContext(ctx, "server closed")
			} else {
				app.Log.ErrorContext(ctx, "server error", "err", err)
			}

			return err
		}

		return nil
	})
	g.Wait()
}
