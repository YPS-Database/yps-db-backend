package yps

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type PingResponse struct {
	OK bool `json:"ok"`
}

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, PingResponse{OK: true})
}

func GetRouter() (router *gin.Engine) {
	router = gin.Default()
	router.GET("/", ping)
	return router
}
