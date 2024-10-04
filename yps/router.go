package yps

import (
	"net/http"

	"github.com/gin-contrib/cors"
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
	router.Use(cors.Default())
	router.GET("/api/", ping)
	router.GET("/api/page/:slug", getPage)
	return router
}
