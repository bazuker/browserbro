package healthcheck

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestHealthcheck(t *testing.T) {
	rw := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rw)

	Healthcheck(c)
	assert.Equal(t, http.StatusOK, rw.Code)
	assert.JSONEq(t, `{"message":"ok"}`, rw.Body.String())
}
