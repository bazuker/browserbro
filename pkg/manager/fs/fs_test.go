package fs

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/bazuker/browserbro/pkg/fs"
	"github.com/bazuker/browserbro/pkg/fs/mock"
	"github.com/bazuker/browserbro/pkg/manager/helper"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestGet(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		var getObjectCalled bool

		rw := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rw)
		c.Set(helper.ContextFileStore, &mock.FileStore{
			GetObjectFn: func(filename string) ([]byte, error) {
				getObjectCalled = true
				return []byte("test"), nil
			},
		})
		c.Params = []gin.Param{{Key: "filename", Value: "test.txt"}}

		Get(c)
		assert.True(t, getObjectCalled)
		assert.Equal(t, http.StatusOK, rw.Code)
		assert.Equal(t, "test", rw.Body.String())
		assert.Equal(t, "text/plain; charset=utf-8", rw.Header().Get("Content-Type"))
	})

	t.Run("handle file not found", func(t *testing.T) {
		var getObjectCalled bool

		rw := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rw)
		c.Set(helper.ContextFileStore, &mock.FileStore{
			GetObjectFn: func(filename string) ([]byte, error) {
				getObjectCalled = true
				return nil, fs.ErrorFileNotFound
			},
		})
		c.Params = []gin.Param{{Key: "filename", Value: "test.txt"}}

		Get(c)
		assert.True(t, getObjectCalled)
		assert.Equal(t, http.StatusNotFound, rw.Code)
		assert.JSONEq(t, `{"error":"file not found"}`, rw.Body.String())
	})

	t.Run("handle internal server error", func(t *testing.T) {
		var getObjectCalled bool

		rw := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(rw)
		c.Set(helper.ContextFileStore, &mock.FileStore{
			GetObjectFn: func(filename string) ([]byte, error) {
				getObjectCalled = true
				return nil, assert.AnError
			},
		})
		c.Params = []gin.Param{{Key: "filename", Value: "test.txt"}}

		Get(c)
		assert.True(t, getObjectCalled)
		assert.Equal(t, http.StatusInternalServerError, rw.Code)
		assert.JSONEq(t, `{"error":"internal server error"}`, rw.Body.String())
	})
}
