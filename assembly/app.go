package assembly

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"

	"github.com/dennypenta/go-api-walkthrough/domain"
	"github.com/dennypenta/go-api-walkthrough/handlers"
	"github.com/dennypenta/go-api-walkthrough/pkg/log"
	"github.com/dennypenta/go-api-walkthrough/repository"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

type App struct {
	Mux     http.Handler
	Log     *slog.Logger
	Migrate *migrate.Migrate

	db *sqlx.DB
}

func (a *App) Close(ctx context.Context) error {
	if err := a.db.Close(); err != nil {
		a.Log.ErrorContext(ctx, "failed to close database connection", "err", err)
		return err
	}

	return nil
}

func NewApp(conf Config) (*App, error) {
	wd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}

	db, err := sqlx.Connect("pgx", conf.PostresDsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	db.DB.SetMaxOpenConns(conf.DbMaxOpenConns)
	db.DB.SetMaxIdleConns(conf.DbMaxIdleConns)
	db.DB.SetConnMaxLifetime(conf.DbConnMaxLifetime)
	db.DB.SetConnMaxIdleTime(conf.DbConnMaxIdleTime)

	migrationsDir := filepath.Join(filepath.Join("file:///", wd), conf.MigrationsDir)
	m, err := migrate.New(
		migrationsDir,
		conf.PostresDsn)
	if err != nil {
		return nil, fmt.Errorf("failed to create migration instance: %w", err)
	}

	userRepo := repository.NewUserRepository(db)
	userService := domain.NewUserService(userRepo)
	userHandlers := handlers.NewHandler(userService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/users", userHandlers.ListUsers)
	mux.HandleFunc("GET /v1/users/{id}", userHandlers.GetUserByID)
	mux.HandleFunc("POST /v1/users", userHandlers.CreateUser)
	mux.HandleFunc("PUT /v1/users/{id}", userHandlers.UpdateUser)
	mux.HandleFunc("DELETE /v1/users/{id}", userHandlers.DeleteUser)

	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
	})

	// https://www.gnu.org/software/libc/manual/html_node/Standard-Streams.html
	// errors and diagnostic messages should go to stderr
	l := log.NewLogger(os.Stderr)
	loggingMiddleware := log.NewLoggingMiddleware(l)
	return &App{
		Mux:     loggingMiddleware(mux),
		Log:     l,
		Migrate: m,

		db: db,
	}, nil
}
