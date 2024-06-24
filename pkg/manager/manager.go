package manager

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/bazuker/browserbro/pkg/fs"
	"github.com/bazuker/browserbro/pkg/fs/local"
	fsEndpoints "github.com/bazuker/browserbro/pkg/manager/fs"
	"github.com/bazuker/browserbro/pkg/manager/healthcheck"
	"github.com/bazuker/browserbro/pkg/manager/helper"
	pluginsRegistry "github.com/bazuker/browserbro/pkg/plugins"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/go-rod/rod"
	"github.com/rs/zerolog/log"
)

// Manager is an HTTP server controller.
type Manager struct {
	server           *http.Server
	router           *gin.Engine
	fileStore        fs.FileStore
	cors             cors.Config
	plugins          []pluginsRegistry.Plugin
	browserConnector connector
}

type connector interface {
	Connect() error
}

type Config struct {
	// ServerAddress is server HTTP address (required).
	ServerAddress string
	// ServerCORS is cross-origin resource sharing configuration
	ServerCORS *cors.Config
	// FileStore is a file storage provider interface (required).
	FileStore fs.FileStore
	// Router is a Gin router instance.
	Router *gin.Engine
	// Browser is a Rod browser instance.
	Browser *rod.Browser
	// BrowserServerID is a unique identifier for the browser server.
	BrowserServerID int
	// BrowserServiceURL is the URL of the browser service.
	BrowserServiceURL string
	// BrowserUserDataDir is the directory where the browser stores its data in the browser service.
	BrowserUserDataDir string
	// BrowserMonitorEnabled enables the browser monitor.
	BrowserMonitorEnabled bool
	// Plugins is a list of plugins to load.
	Plugins []pluginsRegistry.Plugin
}

func DefaultManagerConfig() (Config, error) {
	localFS, err := local.New(local.Config{BasePath: "/tmp/browserBro_files"})
	if err != nil {
		return Config{}, err
	}
	return Config{
		ServerAddress:         ":10001",
		FileStore:             localFS,
		BrowserUserDataDir:    "/tmp/rod/user-data/browserBro_userData",
		BrowserServerID:       1,
		BrowserServiceURL:     "ws://localhost:7317",
		BrowserMonitorEnabled: true,
	}, nil
}

func New(cfg Config) (*Manager, error) {
	// Required configurations.
	if cfg.ServerAddress == "" {
		return nil, errors.New("server address is required")
	}
	if cfg.FileStore == nil {
		return nil, errors.New("file store is required")
	}

	// Optional configurations.
	if cfg.ServerCORS == nil {
		corsCfg := cors.DefaultConfig()
		corsCfg.AllowOrigins = []string{"*"}
		corsCfg.AllowHeaders = []string{"*"}
		cfg.ServerCORS = &corsCfg
	}
	if cfg.Browser == nil {
		cfg.Browser = rod.New()
	}
	if cfg.BrowserServerID <= 0 {
		cfg.BrowserServerID = 1
	}
	if cfg.BrowserServiceURL == "" {
		cfg.BrowserServiceURL = "ws://127.0.0.1:7317"
	}
	if cfg.BrowserUserDataDir == "" {
		cfg.BrowserUserDataDir = "/tmp/rod/user-data/browserBro_userData"
	}
	if cfg.Router == nil {
		cfg.Router = gin.New()
	}

	return &Manager{
		router:    cfg.Router,
		fileStore: cfg.FileStore,
		cors:      *cfg.ServerCORS,
		plugins:   cfg.Plugins,
		browserConnector: newBrowserConnector(
			cfg.Browser,
			cfg.BrowserServerID,
			cfg.BrowserServiceURL,
			cfg.BrowserUserDataDir,
			cfg.BrowserMonitorEnabled,
		),
		server: &http.Server{
			Addr:    cfg.ServerAddress,
			Handler: cfg.Router,
		},
	}, nil
}

func (m *Manager) Run() error {
	m.router.Use(loggerMiddleware(&log.Logger))
	m.router.Use(gin.Recovery())
	m.router.Use(cors.New(m.cors))

	m.router.NoRoute(func(c *gin.Context) {
		c.JSON(
			http.StatusNotFound,
			helper.HTTPMessage{Message: "route not found"},
		)
	})

	api := m.router.Group("/api")
	v1 := api.Group("/v1")
	v1.GET("/health", healthcheck.Healthcheck)

	pluginsGroup := v1.Group("/plugins")
	pluginsGroup.GET("", func(c *gin.Context) {
		pluginNames := make([]string, 0, len(m.plugins))
		for _, plugin := range m.plugins {
			pluginNames = append(pluginNames, plugin.Name())
		}
		c.JSON(http.StatusOK, gin.H{
			"plugins": pluginNames,
		})
	})
	if err := m.loadPlugins(pluginsGroup); err != nil {
		return err
	}

	fsGroup := v1.Group("/files")
	fsGroup.Use(contextMiddleware(m.fileStore))
	fsGroup.GET("/:filename", fsEndpoints.Get)
	fsGroup.DELETE("/:filename", fsEndpoints.Delete)

	if err := m.browserConnector.Connect(); err != nil {
		return err
	}

	go func() {
		if err := m.server.ListenAndServe(); err != nil &&
			err != http.ErrServerClosed {
			log.Fatal().Err(err).Msg("error running the API server")
		}
	}()

	return nil
}

func (m *Manager) Stop() error {
	return m.server.Close()
}

func (m *Manager) loadPlugins(pluginsGroup *gin.RouterGroup) error {
	log.Info().Int("count", len(m.plugins)).Msg("loading plugins")

	for _, plugin := range m.plugins {
		name := plugin.Name()
		pluginsGroup.POST(fmt.Sprintf("/%s", name), func(c *gin.Context) {
			var params map[string]any
			if err := c.ShouldBindJSON(&params); err != nil {
				c.JSON(
					http.StatusBadRequest,
					helper.HTTPMessage{Message: "invalid request body"},
				)
				return
			}
			results, err := plugin.Run(params)
			if err != nil {
				c.JSON(
					http.StatusInternalServerError,
					helper.HTTPMessage{Message: err.Error()},
				)
				return
			}
			c.JSON(http.StatusOK, gin.H{
				name: results,
			})
		})

		log.Info().Str("name", name).Msg("plugin loaded")
	}

	return nil
}
