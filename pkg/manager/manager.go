package manager

import (
	"fmt"
	"net/http"

	"github.com/bazuker/browserbro/pkg/fs"
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
	cfg              Config
	plugins          []pluginsRegistry.Plugin
	browser          *rod.Browser
	browserConnector connector
}

type connector interface {
	Connect() error
}

type Config struct {
	// ServerAddress is server HTTP address.
	ServerAddress string
	// ServerCORS is cross-origin resource sharing configuration
	ServerCORS *cors.Config
	// FileStore is a file storage provider interface.
	FileStore fs.FileStore
	// BrowserServerID is a unique identifier for the browser server.
	BrowserServerID int
	// BrowserServiceURL is the URL of the browser service.
	BrowserServiceURL string
	// UserDataDir is the directory where the browser stores its data in the browser service.
	UserDataDir string
	// BrowserMonitorEnabled enables the browser monitor.
	BrowserMonitorEnabled bool
}

func New(cfg Config) *Manager {
	if cfg.ServerCORS == nil {
		corsCfg := cors.DefaultConfig()
		corsCfg.AllowOrigins = []string{"*"}
		corsCfg.AllowHeaders = []string{"*"}
		cfg.ServerCORS = &corsCfg
	}

	browser := rod.New()
	router := gin.New()

	return &Manager{
		router:  router,
		browser: browser,
		cfg:     cfg,
		browserConnector: newBrowserConnector(
			browser,
			cfg.BrowserServerID,
			cfg.BrowserServiceURL,
			cfg.UserDataDir,
			cfg.BrowserMonitorEnabled,
		),
		server: &http.Server{
			Addr:    cfg.ServerAddress,
			Handler: router,
		},
		plugins: pluginsRegistry.Initialize(
			browser,
			cfg.FileStore,
		),
	}
}

func (m *Manager) Run() error {
	m.router.Use(loggerMiddleware(&log.Logger))
	m.router.Use(gin.Recovery())
	m.router.Use(cors.New(*m.cfg.ServerCORS))

	m.router.NoRoute(func(c *gin.Context) {
		c.JSON(
			http.StatusNotFound,
			helper.HTTPMessage{Message: "route not found"},
		)
	})

	api := m.router.Group("/api")
	v1 := api.Group("/v1")
	v1.GET("/health", healthcheck.Healthcheck)
	v1.Use(contextMiddleware(m.cfg.FileStore))

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
