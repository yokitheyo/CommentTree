package app

import (
	"fmt"
	"time"

	"github.com/wb-go/wbf/dbpg"
	"github.com/wb-go/wbf/ginext"
	"github.com/wb-go/wbf/retry"
	"github.com/wb-go/wbf/zlog"
	"github.com/yokitheyo/CommentTree/internal/config"
	"github.com/yokitheyo/CommentTree/internal/handler/http"
	"github.com/yokitheyo/CommentTree/internal/handler/middleware"
	infradatabase "github.com/yokitheyo/CommentTree/internal/infrastructure/database"
	"github.com/yokitheyo/CommentTree/internal/infrastructure/search"
	"github.com/yokitheyo/CommentTree/internal/repository/postgres"
	retrypkg "github.com/yokitheyo/CommentTree/internal/retry"
	"github.com/yokitheyo/CommentTree/internal/usecase"
)

type dependencies struct {
	database *dbpg.DB
	engine   *ginext.Engine
	usecase  *usecase.CommentUsecase
}

type dependencyBuilder struct {
	cfg  *config.Config
	lg   *zlog.Zerolog
	rm   *resourceManager
	deps *dependencies
}

func newDependencyBuilder(cfg *config.Config, lg *zlog.Zerolog) *dependencyBuilder {
	return &dependencyBuilder{
		cfg:  cfg,
		lg:   lg,
		rm:   &resourceManager{},
		deps: &dependencies{},
	}
}

func (b *dependencyBuilder) initDatabase() error {
	b.lg.Info().Msg("initializing database")

	masterDSN := b.cfg.Database.DSN
	slaves := b.cfg.Database.Slaves

	dbOpts := &dbpg.Options{
		MaxOpenConns:    b.cfg.Database.MaxOpenConns,
		MaxIdleConns:    b.cfg.Database.MaxIdleConns,
		ConnMaxLifetime: time.Duration(b.cfg.Database.ConnMaxLifetimeSec) * time.Second,
	}

	var slavesList []string
	if slaves != "" {
		slavesList = []string{slaves}
	}

	fn := func() error {
		database, err := dbpg.New(masterDSN, slavesList, dbOpts)
		if err != nil {
			b.lg.Warn().Err(err).Msg("failed to connect to database")
			return err
		}

		if database.Master == nil || database.Master.Ping() != nil {
			b.lg.Warn().Msg("database ping failed")
			return fmt.Errorf("database ping failed")
		}

		b.deps.database = database
		return nil
	}

	if err := retry.Do(fn, retry.Strategy(retrypkg.DefaultStrategy)); err != nil {
		return fmt.Errorf("initializing database: %w", err)
	}

	b.lg.Info().Msg("database initialized successfully")
	b.rm.addResource(resource{
		name:      "database",
		closeFunc: func() error { return closeDB(b.deps.database) },
	})

	return nil
}

func (b *dependencyBuilder) runMigrations() error {
	b.lg.Info().Msg("running migrations")

	if err := infradatabase.RunMigrations(b.deps.database, b.cfg.Migrations.Path); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	b.lg.Info().Msg("migrations completed")
	return nil
}

func (b *dependencyBuilder) initRepository() error {
	b.lg.Info().Msg("initializing repository")

	repo := postgres.NewCommentRepository(b.deps.database, retrypkg.DefaultStrategy)
	fts := search.NewPostgresFullText(repo)

	b.deps.usecase = usecase.NewCommentUsecase(repo, fts)

	b.lg.Info().Msg("repository and usecase initialized")
	return nil
}

func (b *dependencyBuilder) initEngine() error {
	b.lg.Info().Msg("initializing Gin engine")

	engine := ginext.New("")
	engine.Use(middleware.LoggerMiddleware(), middleware.CORSMiddleware())

	engine.GET("/", func(c *ginext.Context) {
		c.File("./static/index.html")
	})
	engine.Static("/static", "./static")

	handler := http.NewCommentHandler(b.deps.usecase)
	handler.RegisterRoutes(engine)

	b.deps.engine = engine

	b.lg.Info().Msg("Gin engine initialized")
	return nil
}

func (b *dependencyBuilder) build() (*dependencies, *resourceManager, error) {
	if err := b.initDatabase(); err != nil {
		return nil, b.rm, err
	}

	if err := b.runMigrations(); err != nil {
		return nil, b.rm, err
	}

	if err := b.initRepository(); err != nil {
		return nil, b.rm, err
	}

	if err := b.initEngine(); err != nil {
		return nil, b.rm, err
	}

	return b.deps, b.rm, nil
}

func closeDB(db *dbpg.DB) error {
	if db == nil {
		return nil
	}

	if db.Master != nil {
		if err := db.Master.Close(); err != nil {
			return fmt.Errorf("closing db master: %w", err)
		}
	}

	for i, slave := range db.Slaves {
		if slave != nil {
			if err := slave.Close(); err != nil {
				return fmt.Errorf("closing db slave %d: %w", i, err)
			}
		}
	}

	return nil
}
