package main

import (
	"code-runner/internal/api"
	"code-runner/internal/config"
	"code-runner/internal/database"
	"code-runner/internal/file"
	"code-runner/internal/queue"
	"code-runner/internal/sandbox"
	"code-runner/internal/sandbox/docker"
	"code-runner/internal/spec"
	"code-runner/internal/worker"
	"github.com/joho/godotenv"
	"github.com/zekrotja/rogu/log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	godotenv.Load()

	// 1. Config
	cfg := config.NewEnvProvider("RUNNER_")
	if err := cfg.Load(); err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	// 2. Database
	db, err := database.NewPostgresDB(cfg.Config().Database.DSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Database")
	}
	log.Info().Msg("Connected to Postgres")

	// 3. Queue
	q := queue.NewRedisQueue(cfg.Config().Redis.Addr, cfg.Config().Redis.Pwd)
	log.Info().Msg("Connected to Redis")

	// 4. Specs
	specProvider := spec.NewFileProvider("spec/spec.yaml")

	// 5. Docker Sandbox
	sandboxProvider, err := docker.NewProvider(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to init docker")
	}

	// 6. Sandbox Manager
	fileProvider := file.NewLocalFileProvider()
	mgr, err := sandbox.NewManager(sandboxProvider, specProvider, fileProvider, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create manager")
	}
	defer mgr.Cleanup()

	// 7. Start Auto-Scaling Worker Pool
	pool := worker.NewPool(cfg, q, db, mgr)
	pool.Start()

	// 8. Start API Server (Producer)
	webApi, err := api.NewRestAPI(cfg, specProvider, q, db)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create API")
	}

	go func() {
		if err := webApi.ListenAndServeBlocking(); err != nil {
			log.Fatal().Err(err).Msg("API Error")
		}
	}()

	log.Info().Msgf("Code Runner Started on port %s. Workers: Min=%d, Max=%d", 
		cfg.Config().API.BindAddress, cfg.Config().Worker.Min, cfg.Config().Worker.Max)

	// Graceful Shutdown
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}