package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"nlw-journey/internal/api"
	"nlw-journey/internal/api/spec"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mbndr/figlet4go"
	"github.com/phenpessoa/gutils/netutils/httputils"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func main() {
	ctx := context.Background()
	ctx, cancel := signal.NotifyContext(ctx, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGKILL)

	defer cancel()
	welcome()

	if err := exec(ctx); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}

func welcome() {
	ascii := figlet4go.NewAsciiRender()
	renderStr, _ := ascii.Render("NLW Journey v1.0.0")

	fmt.Println(renderStr)
}

func exec(ctx context.Context) error {
	logger, err := createLogger()
	if err != nil {
		return err
	}

	pool, err := createDatabasePool(ctx)
	if err != nil {
		return err
	}

	si := api.NewAPI(pool, logger)
	router := chi.NewMux()

	router.Use(middleware.RequestID, middleware.Recoverer, httputils.ChiLogger(logger))
	router.Mount("/", spec.Handler(&si))

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		IdleTimeout:  time.Minute,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	}

	defer func() {
		const timeout = 30 * time.Second
		ctx, cancel := context.WithTimeout(context.Background(), timeout)

		defer cancel()

		err := srv.Shutdown(ctx)
		if err != nil {
			logger.Error("Failed to stop process gracefully!", zap.Error(err))
		}
	}()

	errChan := make(chan error, 1)

	go func() {
		err := srv.ListenAndServe()
		if err != nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		return nil
	case err := <-errChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	}

	return nil
}

func createLogger() (*zap.Logger, error) {
	cfg := zap.NewDevelopmentConfig()
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder

	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	logger = logger.Named(">>> JOURNEY:")
	defer logger.Sync()

	return logger, nil
}

func createDatabasePool(ctx context.Context) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, fmt.Sprintf(
		"user=%s password=%s host=%s port=%s dbname=%s",
		"j587893",
		"josnei123",
		"localhost",
		"5432",
		"nlw_journey",
	))
	if err != nil {
		return nil, err
	}

	err = pool.Ping(ctx)
	if err != nil {
		return nil, err
	}

	defer pool.Close()

	return pool, nil
}
