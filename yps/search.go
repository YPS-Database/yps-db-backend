package yps

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SearchRequest struct {
	Query         string `form:"query"`
	SearchContext string `form:"queryContext"`
	EntryLanguage string `form:"language"`
	FilterKey     string `form:"filterKey"`
	FilterValue   string `form:"filterValue"`
	Sort          string `form:"sort"`
	Page          int    `form:"page"`
}

type SearchResponse struct {
	Page         int           `json:"page"`
	TotalPages   int           `json:"total_pages"`
	TotalEntries int           `json:"total_entries"`
	StartEntry   int           `json:"start_entry"`
	EndEntry     int           `json:"end_entry"`
	Entries      []SearchEntry `json:"entries"`
}

func searchEntries(c *gin.Context) {
	var req SearchRequest
	if err := c.ShouldBind(&req); err != nil {
		fmt.Println("Could not get search binding:", err.Error())
		c.JSON(400, gin.H{"error": "Search params not found"})
		return
	}

	response, err := TheDb.Search(req)
	if err != nil {
		fmt.Println("Could not search:", err.Error())
		c.JSON(400, gin.H{"error": "Could not search"})
		return
	}

	c.JSON(http.StatusOK, response)
}
