package main

import (
	"code-runner/internal/api"
	"code-runner/internal/config"
	"code-runner/internal/file"
	"code-runner/internal/sandbox"
	"code-runner/internal/sandbox/docker"
	"code-runner/internal/spec"
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

	specProvider := spec.NewFileProvider("spec/spec.yaml")
	//if err := specProvider.Load(); err != nil {
	//	log.Fatal().Err(err).Msg("Failed to load specs")
	//}

	sandboxProvider, err := docker.NewProvider(cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to init docker")
	}

	// 4. Initialize Managers
	fileProvider := file.NewLocalFileProvider()
	mgr, err := sandbox.NewManager(sandboxProvider, specProvider, fileProvider, cfg)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create manager")
	}

	defer mgr.Cleanup()

	webApi, err := api.NewRestAPI(cfg, specProvider, mgr)
	if err != nil {
		log.Fatal().Err(err).Msg("Failed to create API")
	}

	go func() {
		if err := webApi.ListenAndServeBlocking(); err != nil {
			log.Fatal().Err(err).Msg("API Error")
		}
	}()

	log.Info().Msg("Code Runner Started on port 8080...")

	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc
}
