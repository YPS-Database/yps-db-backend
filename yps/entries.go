package yps

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func updateYpsDb(c *gin.Context) {
}

func getYpsDbs(c *gin.Context) {
}

type SearchEntry struct {
	ID                 string   `json:"id"`
	Title              string   `json:"title"`
	Authors            string   `json:"authors"`
	Year               string   `json:"year"`
	DocumentType       string   `json:"document_type"`
	AvailableLanguages []string `json:"available_languages"`
	Language           string   `json:"language"`
}

type SearchEntriesResponse struct {
	Entries    []SearchEntry `json:"entries"`
	TotalPages int           `json:"total_pages"`
}

func searchEntries(c *gin.Context) {
	entries := []SearchEntry{
		{
			ID:                 "1",
			Title:              "Maaaapping a Sector: Bridging the Evidence Gap on Youth-Driven Peacebuilding: Findings of the Global Survey of Youth-Led Organisations Working on Peace and Security",
			Authors:            "UNOY Peacebuilders, Search for Common Ground",
			Year:               "2017",
			DocumentType:       "Consultation Report",
			AvailableLanguages: []string{"English"},
			Language:           "English",
		},
		{
			ID:                 "2",
			Title:              "Validation Consultation with Youth for the Progress Study on Youth, Peace & Security",
			Authors:            "",
			Year:               "2017",
			DocumentType:       "Concept Note",
			AvailableLanguages: []string{"English", "French", "German"},
			Language:           "English",
		},
		{
			ID:                 "3",
			Title:              "Meeting Report: Youth, Peace, and Security in the Arab States Region: A Consultation and High-level Dialogue",
			Authors:            "Altiok, Ali: Secretariat for the Progress Study on Youth, Peace and Security",
			Year:               "2016",
			DocumentType:       "Consultation Report",
			AvailableLanguages: []string{"English"},
			Language:           "English",
		},
	}

	c.JSON(http.StatusOK, SearchEntriesResponse{
		Entries:    entries,
		TotalPages: 13,
	})
}

func getEntry(c *gin.Context) {
}

type BrowseByFieldValues map[string][]string

type BrowseByFieldsResponse struct {
	Values BrowseByFieldValues `json:"values"`
}

func getBrowseByFields(c *gin.Context) {
	values := BrowseByFieldValues{
		"Org":  {"a", "b", "c"},
		"Year": {"2024", "2023", "2022"},
	}

	c.JSON(http.StatusOK, BrowseByFieldsResponse{
		Values: values,
	})
}
