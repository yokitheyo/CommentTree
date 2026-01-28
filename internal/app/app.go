package app

import (
	"fmt"

	"github.com/wb-go/wbf/zlog"
	"github.com/yokitheyo/CommentTree/internal/config"
)

type App struct {
	cfg  *config.Config
	lg   *zlog.Zerolog
	deps *dependencies
	rm   *resourceManager
}

func New(configPath string) (*App, error) {
	cfg, err := config.Load(configPath)
	if err != nil {
		return nil, fmt.Errorf("loading configuration: %w", err)
	}

	lg := zlog.Logger.With().Str("component", "app").Logger()
	lg.Info().Msg("starting dependency initialization")

	deps, rm, err := newDependencyBuilder(cfg, &lg).build()
	if err != nil {
		return nil, fmt.Errorf("initializing dependencies: %w", err)
	}

	lg.Info().Msg("dependencies initialized successfully")

	return &App{
		cfg:  cfg,
		lg:   &lg,
		deps: deps,
		rm:   rm,
	}, nil
}

func (a *App) Config() *config.Config {
	return a.cfg
}

func (a *App) Logger() *zlog.Zerolog {
	return a.lg
}

func (a *App) Deps() *dependencies {
	return a.deps
}

func (a *App) ResourceManager() *resourceManager {
	return a.rm
}

func (a *App) Close() error {
	return a.rm.closeAll()
}
