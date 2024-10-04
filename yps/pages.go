package yps

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Page struct {
	Content      string
	GoogleFormID string
	Updated      time.Time
}

func (p *Page) ToHTML() string {
	//TODO(dan): return google form too
	return p.Content
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
		HTML:    page.ToHTML(),
		Updated: page.Updated.Unix(),
	})
}
