package main

import (
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/bazuker/browserbro/pkg/fs"
	localFS "github.com/bazuker/browserbro/pkg/fs/local"
	"github.com/bazuker/browserbro/pkg/manager"
	"github.com/bazuker/browserbro/pkg/plugins"
	"github.com/bazuker/browserbro/pkg/plugins/googlesearch"
	"github.com/bazuker/browserbro/pkg/plugins/screenshot"
	"github.com/gin-gonic/gin"
	"github.com/go-rod/rod"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type config struct {
	ServerAddress         string
	BrowserServerID       int
	BrowserServiceURL     string
	BrowserMonitorEnabled bool
	UserDataDir           string
	FileStoreBasePath     string
}

func main() {
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	gin.SetMode(gin.ReleaseMode)

	cfg := config{
		ServerAddress:         ":10001",
		UserDataDir:           "/tmp/rod/user-data/browserBro_userData",
		FileStoreBasePath:     "/tmp/browserBro_files",
		BrowserServerID:       1,
		BrowserServiceURL:     "ws://localhost:7317",
		BrowserMonitorEnabled: true,
	}
	readConfigFromEnvironment(&cfg)

	fileStore, err := localFS.New(localFS.Config{
		BasePath: cfg.FileStoreBasePath,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize file store")
		return
	}

	browser := rod.New()
	allPlugins := initPlugins(browser, fileStore)
	m, err := manager.New(manager.Config{
		ServerAddress:         cfg.ServerAddress,
		FileStore:             fileStore,
		Browser:               browser,
		BrowserUserDataDir:    cfg.UserDataDir,
		BrowserServerID:       cfg.BrowserServerID,
		BrowserServiceURL:     cfg.BrowserServiceURL,
		BrowserMonitorEnabled: cfg.BrowserMonitorEnabled,
		Plugins:               allPlugins,
	})
	if err != nil {
		log.Fatal().Err(err).Msg("failed to initialize manager")
		return
	}

	log.Info().Msg("running API server on " + cfg.ServerAddress)

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

func readConfigFromEnvironment(cfg *config) {
	serverAddress := os.Getenv("BROWSERBRO_SERVER_ADDRESS")
	if serverAddress != "" {
		cfg.ServerAddress = serverAddress
	}
	browserServerID := os.Getenv("BROWSERBRO_BROWSER_SERVER_ID")
	if browserServerID != "" {
		i, err := strconv.Atoi(browserServerID)
		if err != nil {
			log.Fatal().Err(err).
				Msg("failed to parse 'BROWSERBRO_BROWSER_SERVER_ID' environment variable")
			return
		}
		cfg.BrowserServerID = i
	}
	browserServiceURL := os.Getenv("BROWSERBRO_BROWSER_SERVICE_URL")
	if browserServiceURL != "" {
		cfg.BrowserServiceURL = browserServiceURL
	}
	browserMonitorEnabled := os.Getenv("BROWSERBRO_BROWSER_MONITOR_ENABLED")
	if browserMonitorEnabled != "" {
		b, err := strconv.ParseBool(browserMonitorEnabled)
		if err != nil {
			log.Fatal().Err(err).
				Msg("failed to parse 'BROWSERBRO_BROWSER_MONITOR_ENABLED' environment variable")
			return
		}
		cfg.BrowserMonitorEnabled = b
	}
	userDataDir := os.Getenv("BROWSERBRO_BROWSER_USER_DATA_DIR")
	if userDataDir != "" {
		cfg.UserDataDir = userDataDir
	}
	fileStoreBasePath := os.Getenv("BROWSERBRO_FILE_STORE_BASE_PATH")
	if fileStoreBasePath != "" {
		cfg.FileStoreBasePath = fileStoreBasePath
	}
}

func initPlugins(browser *rod.Browser, fileStore fs.FileStore) []plugins.Plugin {
	return []plugins.Plugin{
		googlesearch.New(browser),
		screenshot.New(browser, fileStore),
	}
}
