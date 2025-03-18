package app

import (
	"fmt"
	"github.com/labstack/echo/v4"
	log "github.com/sirupsen/logrus"
	"github.com/spanwalla/song-library/config"
	_ "github.com/spanwalla/song-library/docs"
	v1 "github.com/spanwalla/song-library/internal/controller/http/v1"
	"github.com/spanwalla/song-library/internal/repository"
	"github.com/spanwalla/song-library/internal/service"
	"github.com/spanwalla/song-library/internal/webapi"
	"github.com/spanwalla/song-library/pkg/httpserver"
	"github.com/spanwalla/song-library/pkg/postgres"
	"os"
	"os/signal"
	"syscall"
)

// @title Song Library
// @version 1.0

// @host localhost:8080
// @BasePath /api/v1

// Run creates objects via constructors
func Run() {
	// Config
	configPath, ok := os.LookupEnv("CONFIG_PATH")
	if !ok || len(configPath) == 0 {
		log.Fatal("app - os.LookupEnv: CONFIG_PATH is empty")
	}

	cfg, err := config.New(configPath)
	if err != nil {
		log.Fatal(fmt.Errorf("app - config.New: %w", err))
	}

	// Logger
	initLogger(cfg.Log.Level)
	log.Info("Config read")

	// Postgres
	log.Info("Connecting to postgres...")
	pg, err := postgres.New(cfg.PG.URL, postgres.MaxPoolSize(cfg.PG.PoolMax))
	if err != nil {
		log.Fatal(fmt.Errorf("app - Run - postgres.New: %w", err))
	}
	defer pg.Close()

	// Services and repos
	log.Info("Initializing services and repos...")
	services := service.NewServices(service.Dependencies{
		Repos:    repository.NewRepositories(pg),
		SongInfo: webapi.NewSongInfoWebAPI(cfg.SongAPI.URL),
	})

	// Echo handler
	log.Info("Initializing handlers and routes...")
	handler := echo.New()
	v1.ConfigureRouter(handler, services)

	// HTTP Server
	log.Info("Starting HTTP server...")
	log.Debugf("Server port: %s", cfg.HTTP.Port)
	httpServer := httpserver.New(handler, httpserver.Port(cfg.HTTP.Port))

	log.Info("Configuring graceful shutdown...")
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt, syscall.SIGTERM)

	select {
	case s := <-interrupt:
		log.Info("app - Run - signal: " + s.String())
	case err = <-httpServer.Notify():
		log.Errorf("app - Run - httpServer.Notify: %v", err)
	}

	// Graceful shutdown
	log.Info("Shutting down...")

	err = httpServer.Shutdown()
	if err != nil {
		log.Errorf("app - Run - httpServer.Shutdown: %v", err)
	}
}
