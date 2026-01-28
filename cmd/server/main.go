package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/wb-go/wbf/zlog"
	"github.com/yokitheyo/CommentTree/internal/app"
)

func main() {
	zlog.Init()
	zlog.Logger.Info().Msg("application starting")

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	application, err := app.New("config.yaml")
	if err != nil {
		zlog.Logger.Fatal().Err(err).Msg("failed to create app")
	}
	defer application.Close()

	if err := application.Run(ctx); err != nil {
		zlog.Logger.Fatal().Err(err).Msg("application failed")
	}
}
