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
	"github.com/dennypenta/go-api-walkthrough/pkg/metrics"
	"github.com/golang-migrate/migrate/v4"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

	if err := app.Migrate.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalln("failed to run migrations:", err)
	}

	reg := prometheus.NewRegistry()
	metricsObj := metrics.NewMetrics(reg)
	metricsHandler := promhttp.HandlerFor(reg, promhttp.HandlerOpts{})
	metricsMux := http.NewServeMux()
	metricsMux.Handle("/metrics", metricsHandler)
	metircsServer := &http.Server{Addr: ":8081", Handler: metricsMux}
	metricsMiddleware := metrics.NewMetricsMiddleware(metricsObj)

	server := &http.Server{Addr: ":" + conf.HttpPort, Handler: metricsMiddleware(app.Mux)}

	g, ctx := errgroup.WithContext(ctx)
	g.Go(func() error {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		<-c
		cancel()

		app.Log.InfoContext(ctx, "shutting down server")
		if err := server.Shutdown(ctx); err != nil {
			app.Log.ErrorContext(ctx, "failed to shutdown server", "err", err)
		}
		return app.Close(ctx)
	})
	g.Go(func() error {
		app.Log.InfoContext(ctx, "server has been started", "port", conf.HttpPort)
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
	g.Go(func() error {
		app.Log.InfoContext(ctx, "metrics server has been started", "port", "8081")
		if err := metircsServer.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				app.Log.InfoContext(ctx, "metrics server closed")
			} else {
				app.Log.ErrorContext(ctx, "metrics server error", "err", err)
			}

			return err
		}

		return nil
	})
	if err := g.Wait(); err != nil {
		app.Log.ErrorContext(ctx, "server stopped", "err", err)
		os.Exit(1)
	}

	app.Log.InfoContext(ctx, "server stopped")
}
