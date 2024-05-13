package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/bazuker/browserbro/pkg/fs/mock"
	"github.com/bazuker/browserbro/pkg/plugins"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	t.Run("default config", func(t *testing.T) {
		fs := &mock.FileStore{}
		cfg := Config{
			ServerAddress: ":10001",
			FileStore:     fs,
			Plugins: []plugins.Plugin{
				&mockPlugin{name: "test"},
			},
		}
		m, err := New(cfg)
		require.NoError(t, err)
		require.NotNil(t, m)

		assert.NotNil(t, m.router)
		assert.Equal(t, ":10001", m.server.Addr)
		assert.NotNil(t, m.cors)
		assert.Equal(t, fs, m.fileStore)
		require.Implements(t, (*connector)(nil), m.browserConnector)
		bc := m.browserConnector.(*browserConnector)
		assert.Equal(t, 1, bc.serverID)
		assert.Equal(t, "ws://127.0.0.1:7317", bc.serviceURL)
		assert.Equal(t, "/tmp/rod/user-data/browserBro_userData", bc.userDataDir)
		assert.False(t, bc.browserMonitoringEnabled)
	})

	t.Run("custom properties", func(t *testing.T) {
		corsCfg := cors.DefaultConfig()
		corsCfg.AllowOrigins = []string{"custom.example.com"}
		g := gin.New()
		plugin := &mockPlugin{name: "test"}
		cfg := Config{
			ServerAddress:         ":10001",
			ServerCORS:            &corsCfg,
			FileStore:             &mock.FileStore{},
			Router:                g,
			BrowserMonitorEnabled: true,
			BrowserServiceURL:     "ws://example.com:7317",
			BrowserUserDataDir:    "/tmp/rod/user-data/my/data",
			Plugins: []plugins.Plugin{
				plugin,
			},
		}
		m, err := New(cfg)
		require.NoError(t, err)
		require.NotNil(t, m)

		assert.Equal(t, corsCfg, m.cors)
		assert.Equal(t, g, m.router)
		bc := m.browserConnector.(*browserConnector)
		assert.Equal(t, 1, bc.serverID)
		assert.Equal(t, "ws://example.com:7317", bc.serviceURL)
		assert.Equal(t, "/tmp/rod/user-data/my/data", bc.userDataDir)
		assert.True(t, bc.browserMonitoringEnabled)
		require.Len(t, m.plugins, 1)
		assert.Equal(t, plugin, m.plugins[0])
	})

	t.Run("missing server address", func(t *testing.T) {
		_, err := New(Config{})
		require.Error(t, err)
		assert.EqualError(t, err, "server address is required")
	})

	t.Run("missing file store", func(t *testing.T) {
		_, err := New(Config{
			ServerAddress: ":10001",
		})
		require.Error(t, err)
		assert.EqualError(t, err, "file store is required")
	})

	t.Run("missing plugins", func(t *testing.T) {
		_, err := New(Config{
			ServerAddress: ":10001",
			FileStore:     &mock.FileStore{},
		})
		require.Error(t, err)
		assert.EqualError(t, err, "at least one plugin is required")
	})
}

func TestManager_Start(t *testing.T) {
	var pluginRunCalled bool
	m, err := New(Config{
		ServerAddress:     ":10001",
		BrowserServiceURL: "ws://example.com:7317",
		FileStore:         &mock.FileStore{},
		Plugins: []plugins.Plugin{
			&mockPlugin{
				name: "test",
				runFn: func(params map[string]interface{}) (
					map[string]interface{},
					error,
				) {
					pluginRunCalled = true
					return nil, nil
				},
			},
			&mockPlugin{
				name: "error",
				runFn: func(params map[string]interface{}) (
					map[string]interface{},
					error,
				) {
					return nil, fmt.Errorf("plugin error")
				},
			},
		},
	})
	require.NoError(t, err)
	// Use mock connector to avoid actual browser service dialing.
	m.browserConnector = &mockConnector{}

	require.NoError(t, m.Run())
	defer func() {
		require.NoError(t, m.Stop())
	}()

	t.Run("verify that endpoints are registered", func(t *testing.T) {
		resp := performRequest(m.router, http.MethodGet, "/DNE", nil)
		assert.Equal(t, http.StatusNotFound, resp.Code)
		assert.JSONEq(t, `{"message":"route not found"}`, resp.Body.String())

		require.True(t, routeExists(m.router, http.MethodGet, "/api/v1/health"))
		require.True(t, routeExists(m.router, http.MethodGet, "/api/v1/files/:filename"))
		require.True(t, routeExists(m.router, http.MethodDelete, "/api/v1/files/:filename"))
		require.True(t, routeExists(m.router, http.MethodGet, "/api/v1/plugins"))
		for _, plugin := range m.plugins {
			assert.True(
				t,
				routeExists(
					m.router,
					http.MethodPost,
					fmt.Sprintf("/api/v1/plugins/%s", plugin.Name()),
				),
			)
		}
	})

	t.Run("plugins list", func(t *testing.T) {
		resp := performRequest(m.router, http.MethodGet, "/api/v1/plugins", nil)
		assert.Equal(t, http.StatusOK, resp.Code)
		pluginNames := make([]string, 0, len(m.plugins))
		for _, plugin := range m.plugins {
			pluginNames = append(pluginNames, plugin.Name())
		}
		var respPluginNames map[string][]string
		require.NoError(t, json.Unmarshal(resp.Body.Bytes(), &respPluginNames))
		assert.ElementsMatch(t, pluginNames, respPluginNames["plugins"])
	})

	t.Run("run plugin", func(t *testing.T) {
		resp := performRequest(
			m.router,
			http.MethodPost,
			"/api/v1/plugins/test",
			bytes.NewBuffer([]byte("{}")),
		)
		assert.Equal(t, http.StatusOK, resp.Code)
		require.True(t, pluginRunCalled)
	})

	t.Run("run plugin with invalid request", func(t *testing.T) {
		resp := performRequest(
			m.router,
			http.MethodPost,
			"/api/v1/plugins/test",
			nil,
		)
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		require.JSONEq(
			t,
			`{"message":"invalid request body"}`,
			resp.Body.String(),
		)
	})

	t.Run("handle plugin error", func(t *testing.T) {
		resp := performRequest(
			m.router,
			http.MethodPost,
			"/api/v1/plugins/error",
			bytes.NewBuffer([]byte("{}")),
		)
		assert.Equal(t, http.StatusInternalServerError, resp.Code)
		require.JSONEq(
			t,
			`{"message":"plugin error"}`,
			resp.Body.String(),
		)
	})
}

func routeExists(router *gin.Engine, method, path string) bool {
	for _, route := range router.Routes() {
		if route.Method == method && route.Path == path {
			return true
		}
	}
	return false
}

type mockConnector struct{}

func (mr *mockConnector) Connect() error {
	return nil
}

type mockPlugin struct {
	name  string
	runFn func(params map[string]interface{}) (map[string]interface{}, error)
}

func (mp *mockPlugin) Name() string {
	return mp.name
}

func (mp *mockPlugin) Run(params map[string]interface{}) (
	map[string]interface{},
	error,
) {
	if mp.runFn == nil {
		return nil, nil
	}
	return mp.runFn(params)
}
