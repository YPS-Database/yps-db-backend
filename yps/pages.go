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

type PageRequest struct {
	ID string `uri:"slug" binding:"required"`
}

type PageResponse struct {
	ID           string `json:"id"`
	MD           string `json:"markdown"`
	GoogleFormID string `json:"google_form_id"`
	Updated      int64  `json:"updated"`
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
		ID:           req.ID,
		MD:           page.Content,
		GoogleFormID: page.GoogleFormID,
		Updated:      page.Updated.Unix(),
	})
}
