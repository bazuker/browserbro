package fs

import (
	"errors"
	"net/http"

	"github.com/bazuker/browserbro/pkg/fs"
	"github.com/bazuker/browserbro/pkg/manager/helper"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

func Get(c *gin.Context) {
	filename := c.Param("filename")
	fileStoreContext := c.MustGet(helper.ContextFileStore)
	fileStore := fileStoreContext.(fs.FileStore)

	data, err := fileStore.GetObject(filename)
	if err != nil {
		if errors.Is(err, fs.ErrorFileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		log.Error().Err(err).Msg("failed to get file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.Data(http.StatusOK, http.DetectContentType(data), data)
}

func Delete(c *gin.Context) {
	filename := c.Param("filename")
	fileStoreContext := c.MustGet(helper.ContextFileStore)
	fileStore := fileStoreContext.(fs.FileStore)

	err := fileStore.DeleteObject(filename)
	if err != nil {
		if errors.Is(err, fs.ErrorFileNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "file not found"})
			return
		}
		log.Error().Err(err).Msg("failed to delete file")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal server error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "file deleted"})
}
