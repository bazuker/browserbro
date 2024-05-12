package healthcheck

import (
	"github.com/bazuker/browserbro/pkg/manager/helper"
	"github.com/gin-gonic/gin"
	"net/http"
)

func Healthcheck(c *gin.Context) {
	c.JSON(http.StatusOK, helper.HTTPMessage{Message: "ok"})
}
