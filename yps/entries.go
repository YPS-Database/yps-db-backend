package yps

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Entry struct {
	ItemID         string
	Title          string
	Authors        string
	URL            string
	OrgPublisher   string
	OrgDocID       string
	OrgType        string
	DocType        string
	Abstract       string
	YouthLed       string
	Keywords       []string
	StartDate      time.Time
	EndDate        time.Time
	Language       string
	AltLanguageIDs []string
	RelatedIDs     []string
}

type ImportTryResponse struct {
	TotalEntries int                  `json:"total_entries"`
	NewEntries   int                  `json:"new_entries"`
	Entries      map[string]xlsxEntry `json:"entries"`
	OldEntries   map[string]Entry     `json:"old_entries"`
}

func updateYpsDb(c *gin.Context) {
	// whether to apply the changes or not
	_, apply := c.GetQuery("apply")
	fmt.Println("Apply changes?", apply)

	if apply {
		applyYpsDbUpdate(c)
	} else {
		testYpsDbUpdate(c)
	}
}

func applyYpsDbUpdate(c *gin.Context) {
	// load passed db file
	fileHeader, err := c.FormFile("db")
	if err != nil {
		fmt.Println("Could not get file from updateYpsDb call:", err.Error())
		c.JSON(400, gin.H{"error": "Could not get 'db' file in form body."})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		fmt.Println("Could not open file from updateYpsDb call:", err.Error())
		c.JSON(400, gin.H{"error": "Could not open 'db' file in form body."})
		return
	}

	newEntries, err := ReadEntriesFile(file)
	if err != nil {
		fmt.Println("Could not read entries file:", err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = TheDb.UploadEntries(newEntries.Entries)
	if err != nil {
		fmt.Println("Could not upload entries:", err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"ok": true})
}

func testYpsDbUpdate(c *gin.Context) {
	// load passed db file
	fileHeader, err := c.FormFile("db")
	if err != nil {
		fmt.Println("Could not get file from updateYpsDb call:", err.Error())
		c.JSON(400, gin.H{"error": "Could not get 'db' file in form body."})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		fmt.Println("Could not open file from updateYpsDb call:", err.Error())
		c.JSON(400, gin.H{"error": "Could not open 'db' file in form body."})
		return
	}

	newEntries, err := ReadEntriesFile(file)
	if err != nil {
		fmt.Println("Could not read entries file:", err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	existingEntries, err := TheDb.GetAllEntries()
	if err != nil {
		fmt.Println("Could not existing entries:", err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("entries:", len(existingEntries))

	c.JSON(http.StatusOK, ImportTryResponse{
		TotalEntries: len(newEntries.Entries),
		NewEntries:   len(newEntries.Entries) - len(existingEntries),
		// Entries:      newEntries.Entries,
		Entries:    nil,
		OldEntries: existingEntries,
	})
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
	values, err := TheDb.GetBrowseByFields()
	if err != nil {
		fmt.Println("Could not get browse by fields:", err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, BrowseByFieldsResponse{
		Values: values,
	})
}
