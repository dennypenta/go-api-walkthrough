package assembly

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/dennypenta/go-api-walkthrough/domain"
	"github.com/dennypenta/go-api-walkthrough/handlers"
	"github.com/dennypenta/go-api-walkthrough/pkg/log"
	"github.com/dennypenta/go-api-walkthrough/repository"
	_ "github.com/jackc/pgx/stdlib"
	"github.com/jmoiron/sqlx"
)

type App struct {
	Mux http.Handler
	Log *slog.Logger

	db *sqlx.DB
}

func (a *App) Close(ctx context.Context) {
	if err := a.db.Close(); err != nil {
		a.Log.ErrorContext(ctx, "failed to close database connection", "err", err)
	}
}

func NewApp(conf Config) (*App, error) {
	db, err := sqlx.Connect("pgx", conf.PostresDsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to postgres: %w", err)
	}
	db.DB.SetMaxOpenConns(conf.DbMaxOpenConns)
	db.DB.SetMaxIdleConns(conf.DbMaxIdleConns)
	db.DB.SetConnMaxLifetime(conf.DbConnMaxLifetime)
	db.DB.SetConnMaxIdleTime(conf.DbConnMaxIdleTime)

	userRepo := repository.NewUserRepository(db)
	userService := domain.NewUserService(userRepo)
	userHandlers := handlers.NewHandler(userService)

	mux := http.NewServeMux()

	mux.HandleFunc("GET /v1/users", userHandlers.ListUsers)
	mux.HandleFunc("GET /v1/users/{id}", userHandlers.GetUserByID)
	mux.HandleFunc("POST /v1/users", userHandlers.CreateUser)
	mux.HandleFunc("PUT /v1/users/{id}", userHandlers.UpdateUser)
	mux.HandleFunc("DELETE /v1/users/{id}", userHandlers.DeleteUser)

	// https://www.gnu.org/software/libc/manual/html_node/Standard-Streams.html
	// errors and diagnostic messages should go to stderr
	l := log.NewLogger(os.Stderr)
	loggingMiddleware := log.NewLoggingMiddleware(l)
	return &App{
		Mux: loggingMiddleware(mux),
		Log: l,

		db: db,
	}, nil
}
