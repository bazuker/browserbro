package manager

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/bazuker/browserbro/pkg/fs/mock"
	"github.com/bazuker/browserbro/pkg/plugins"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNew(t *testing.T) {
	fs := &mock.FileStore{}
	cfg := Config{
		ServerAddress:         ":10001",
		ServerCORS:            nil,
		FileStore:             fs,
		BrowserServerID:       1,
		BrowserServiceURL:     "ws://example.com:7317",
		UserDataDir:           "/tmp/my_data",
		BrowserMonitorEnabled: true,
	}
	m := New(cfg)
	require.NotNil(t, m)

	assert.NotNil(t, m.router)
	assert.NotNil(t, m.browser)

	assert.Equal(t, ":10001", m.cfg.ServerAddress)
	assert.NotNil(t, m.cfg.ServerCORS)
	assert.Equal(t, fs, m.cfg.FileStore)
	assert.Equal(t, 1, m.cfg.BrowserServerID)
	assert.Equal(t, "ws://example.com:7317", m.cfg.BrowserServiceURL)
	assert.Equal(t, "/tmp/my_data", m.cfg.UserDataDir)
	assert.True(t, m.cfg.BrowserMonitorEnabled)
}

func TestManager_Start(t *testing.T) {
	var pluginRunCalled bool
	plugins.AddCustom(&mockPlugin{
		name: "test",
		runFn: func(params map[string]interface{}) (
			map[string]interface{},
			error,
		) {
			pluginRunCalled = true
			return nil, nil
		},
	})
	plugins.AddCustom(&mockPlugin{
		name: "error",
		runFn: func(params map[string]interface{}) (
			map[string]interface{},
			error,
		) {
			return nil, fmt.Errorf("plugin error")
		},
	})

	m := New(Config{
		ServerAddress:     ":10001",
		BrowserServiceURL: "ws://example.com:7317",
		FileStore:         &mock.FileStore{},
	})
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
