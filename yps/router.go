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

func GetRouter(trustedProxies []string, corsAllowedFrom []string) (router *gin.Engine) {
	router = gin.Default()

	router.SetTrustedProxies(trustedProxies)

	corsConfig := cors.DefaultConfig()
	corsConfig.AllowOrigins = corsAllowedFrom
	router.Use(cors.New(corsConfig))

	router.GET("/api/", ping)
	router.GET("/api/page/:slug", getPage)
	router.POST("/api/auth", login)

	router.GET("/api/test", AdminAuthMiddleware(), ping)
	return router
}
