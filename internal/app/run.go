package app

import (
	"context"
	"fmt"
	"net/http"
	"time"
)

func (a *App) Run(ctx context.Context) error {
	srv := &http.Server{
		Addr:    a.cfg.Server.Addr,
		Handler: a.deps.engine,
	}

	go func() {
		a.lg.Info().Str("addr", srv.Addr).Msg("starting HTTP server")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.lg.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	<-ctx.Done()
	return a.Shutdown(ctx, srv)
}

func (a *App) Shutdown(parentCtx context.Context, srv *http.Server) error {
	a.lg.Info().Msg("shutdown signal received")

	shutdownCtx, cancel := context.WithTimeout(
		context.WithoutCancel(parentCtx),
		time.Duration(a.cfg.Server.ShutdownTimeoutSec)*time.Second,
	)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		a.lg.Error().Err(err).Msg("HTTP server shutdown failed")
		return fmt.Errorf("shutdown http server: %w", err)
	}

	a.lg.Info().Msg("HTTP server stopped gracefully")

	if err := a.Close(); err != nil {
		a.lg.Error().Err(err).Msg("failed to close resources")
		return fmt.Errorf("close app resources: %w", err)
	}

	a.lg.Info().Msg("shutdown complete")
	return nil
}
