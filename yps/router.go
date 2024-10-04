package yps

import (
	"fmt"
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

type PageRequest struct {
	ID string `uri:"slug" binding:"required"`
}

type PageResponse struct {
	ID      string `json:"id"`
	HTML    string `json:"html"`
	Updated int64  `json:"updated"`
}

func getPage(c *gin.Context) {
	var req PageRequest
	if err := c.ShouldBindUri(&req); err != nil {
		fmt.Println("Could not get page URI binding:", err.Error())
		c.JSON(400, gin.H{"error": "Page must be given"})
		return
	}

	page, err := TheDb.GetPage(req.ID)
	if err != nil {
		fmt.Println("Could not get page:", err.Error())
		c.JSON(400, gin.H{"error": "Could not get page"})
		return
	}

	c.JSON(http.StatusOK, PageResponse{
		ID:      req.ID,
		HTML:    page.Content,
		Updated: page.Updated.Unix(),
	})
}

func GetRouter() (router *gin.Engine) {
	router = gin.Default()
	router.Use(cors.Default())
	router.GET("/api/", ping)
	router.GET("/api/page/:slug", getPage)
	return router
}
