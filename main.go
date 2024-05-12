package main

import (
	"os"
	"os/signal"
	"syscall"

	localFS "github.com/bazuker/browserbro/pkg/fs/local"
	"github.com/bazuker/browserbro/pkg/manager"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	gin.SetMode(gin.ReleaseMode)

	fileStore, err := localFS.New(localFS.Config{
		BasePath: "/tmp/browserBro_files",
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize file store")
		return
	}

	serverAddress := ":10001"
	m := manager.New(manager.Config{
		ServerAddress:         serverAddress,
		FileStore:             fileStore,
		BrowserServiceURL:     "ws://localhost:7317",
		UserDataDir:           "/tmp/rod/user-data/browserBro_userData",
		BrowserMonitorEnabled: true,
	})

	log.Info().Msg("running API server on " + serverAddress)

	if err := m.Run(); err != nil {
		log.Fatal().Err(err).Msg("error running the API server")
	}

	quit := make(chan os.Signal, 1)
	// Signal notification for Interrupt (Ctrl+C) and SIGTERM (Termination signal from the OS)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	if err := m.Stop(); err != nil {
		log.Fatal().Err(err).Msg("error stopping the API server")
	}
}
