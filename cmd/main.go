package main

import (
	"code-runner/internal/api"
	"code-runner/internal/config"
	"code-runner/internal/database"
	"code-runner/internal/file"
	"code-runner/internal/question"
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

	cfg := config.NewEnvProvider("RUNNER_")
	if err := cfg.Load(); err != nil {
		log.Fatal().Err(err).Msg("Failed to load config")
	}

	db, err := database.NewPostgresDB(cfg.Config().Database.DSN)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to connect to Database")
	}
	log.Info().Msg("Connected to Postgres")

	q := queue.NewRedisQueue(cfg.Config().Redis.Addr, cfg.Config().Redis.Pwd)
	log.Info().Msg("Connected to Redis")

	specProvider := spec.NewFileProvider("spec/spec.yaml")
	
	// Init Question Provider
	// Assumes "Questions" dir is in the current working directory
	questionProvider := question.NewProvider("Questions")

	sandboxProvider, err := docker.NewProvider(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to init docker")
	}

	fileProvider := file.NewLocalFileProvider()
	mgr, err := sandbox.NewManager(sandboxProvider, specProvider, fileProvider, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create manager")
	}
	defer mgr.Cleanup()

	for i := 0; i < cfg.Config().WorkerCount; i++ {
		// Pass question provider to worker
		w := worker.NewWorker(i+1, q, db, mgr, questionProvider)
		go w.Start()
	}

	// Pass question provider to API
	webApi, err := api.NewRestAPI(cfg, specProvider, q, db, questionProvider)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create API")
	}

	go func() {
		if err := webApi.ListenAndServeBlocking(); err != nil {
			log.Fatal().Err(err).Msg("API Error")
		}
	}()

	log.Info().Msgf("Code Runner Started on port %s with %d workers...", cfg.Config().API.BindAddress, cfg.Config().WorkerCount)

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}