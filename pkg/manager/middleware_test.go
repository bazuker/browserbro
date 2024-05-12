package manager

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/bazuker/browserbro/pkg/fs/mock"
	"github.com/bazuker/browserbro/pkg/manager/helper"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func Test_loggerMiddleware(t *testing.T) {
	buffer := new(bytes.Buffer)
	var memLogger = zerolog.New(buffer).With().Timestamp().Logger()
	r := gin.New()
	r.Use(loggerMiddleware(&memLogger))
	r.GET("/example", func(c *gin.Context) {})

	logData := struct {
		IP         string `json:"ip"`
		Method     string `json:"method"`
		StatusCode int    `json:"status_code"`
		Path       string `json:"path"`
		Latency    string `json:"latency"`
		Level      string `json:"level"`
		Time       string `json:"time"`
	}{}

	performRequest(r, "GET", "/example?n=42", nil)
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &logData))
	assert.Equal(t, logData.StatusCode, http.StatusOK)
	assert.Equal(t, logData.Method, "GET")
	assert.Equal(t, logData.Path, "/example?n=42")
	assert.Equal(t, logData.Level, "info")
	assert.NotEmpty(t, logData.Time)
	assert.NotEmpty(t, logData.Latency)
	assert.NotEmpty(t, logData.IP)

	buffer.Reset()
	performRequest(r, "GET", "/notfound", nil)
	require.NoError(t, json.Unmarshal(buffer.Bytes(), &logData))
	assert.Equal(t, logData.StatusCode, http.StatusNotFound)
	assert.Equal(t, logData.Method, "GET")
	assert.Equal(t, logData.Path, "/notfound")
	assert.Equal(t, logData.Level, "info")
	assert.NotEmpty(t, logData.Time)
	assert.NotEmpty(t, logData.Latency)
	assert.NotEmpty(t, logData.IP)
}

func Test_contextMiddleware(t *testing.T) {
	var endpointCalled bool
	r := gin.New()
	fs := &mock.FileStore{}
	r.Use(contextMiddleware(fs))

	r.GET("/example", func(c *gin.Context) {
		endpointCalled = true
		fileStore, ok := c.Get(helper.ContextFileStore)
		require.True(t, ok)
		assert.Equal(t, fs, fileStore)
	})

	performRequest(r, "GET", "/example", nil)
	assert.True(t, endpointCalled)
}
