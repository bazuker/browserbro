package manager

import (
	"time"

	"github.com/bazuker/browserbro/pkg/fs"
	"github.com/bazuker/browserbro/pkg/manager/helper"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
)

func loggerMiddleware(logger *zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		param := gin.LogFormatterParams{}
		param.TimeStamp = time.Now()
		param.Latency = param.TimeStamp.Sub(start)
		if param.Latency > time.Minute {
			param.Latency = param.Latency.Truncate(time.Second)
		}

		param.ClientIP = c.ClientIP()
		param.Method = c.Request.Method
		param.StatusCode = c.Writer.Status()
		param.ErrorMessage = c.Errors.ByType(gin.ErrorTypePrivate).String()
		param.BodySize = c.Writer.Size()
		if raw != "" {
			path = path + "?" + raw
		}
		param.Path = path

		var logEvent *zerolog.Event
		if c.Writer.Status() >= 500 {
			logEvent = logger.Error()
		} else {
			logEvent = logger.Info()
		}

		logEvent.Str("ip", param.ClientIP).
			Str("method", param.Method).
			Int("status_code", param.StatusCode).
			Str("path", param.Path).
			Str("latency", param.Latency.String()).
			Msg(param.ErrorMessage)
	}
}

func contextMiddleware(
	fileStore fs.FileStore,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(helper.ContextFileStore, fileStore)
		c.Next()
	}
}
