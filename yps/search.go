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

	// entries = []SearchEntry{
	// 	{
	// 		ID:                 "1",
	// 		Title:              "Maaaapping a Sector: Bridging the Evidence Gap on Youth-Driven Peacebuilding: Findings of the Global Survey of Youth-Led Organisations Working on Peace and Security",
	// 		Authors:            "UNOY Peacebuilders, Search for Common Ground",
	// 		Year:               "2017",
	// 		DocumentType:       "Consultation Report",
	// 		AvailableLanguages: []string{"English"},
	// 		Language:           "English",
	// 	},
	// 	{
	// 		ID:                 "2",
	// 		Title:              "Validation Consultation with Youth for the Progress Study on Youth, Peace & Security",
	// 		Authors:            "",
	// 		Year:               "2017",
	// 		DocumentType:       "Concept Note",
	// 		AvailableLanguages: []string{"English", "French", "German"},
	// 		Language:           "English",
	// 	},
	// 	{
	// 		ID:                 "3",
	// 		Title:              "Meeting Report: Youth, Peace, and Security in the Arab States Region: A Consultation and High-level Dialogue",
	// 		Authors:            "Altiok, Ali: Secretariat for the Progress Study on Youth, Peace and Security",
	// 		Year:               "2016",
	// 		DocumentType:       "Consultation Report",
	// 		AvailableLanguages: []string{"English"},
	// 		Language:           "English",
	// 	},
	// }

	c.JSON(http.StatusOK, response)
}
