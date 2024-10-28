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

	// API
	router.GET("/api/ping", ping)
	router.POST("/api/auth", login)

	// DB
	router.GET("/api/dbs", getYpsDbs)
	router.PUT("/api/db", AdminAuthMiddleware(), updateYpsDb)

	// pages
	router.GET("/api/page/:slug", getPage)

	// entries
	router.GET("/api/entry/:slug", getEntry)
	router.GET("/api/browseby", getBrowseByFields)
	router.GET("/api/search", searchEntries)

	return router
}
