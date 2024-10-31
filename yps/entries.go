package yps

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type Entry struct {
	ItemID          string
	Title           string
	Authors         string
	URL             string
	OrgPublishers   []string
	OrgDocID        string
	OrgType         string
	DocType         string
	Abstract        string
	YouthLed        string
	YouthLedDetails string
	Keywords        []string
	StartDate       time.Time
	EndDate         time.Time
	Language        string
	AltLanguageIDs  []string
	RelatedIDs      []string
}

type ImportTryResponse struct {
	TotalEntries int                  `json:"total_entries"`
	NewEntries   int                  `json:"new_entries"`
	Entries      map[string]xlsxEntry `json:"entries"`
	OldEntries   map[string]Entry     `json:"old_entries"`
}

var TheBrowseByFields *BrowseByFieldValues

func UpdateBrowseByFields() error {
	bbf, err := TheDb.GetBrowseByFields()
	if err != nil {
		fmt.Println("Failed to update browse by fields:", err)
	} else {
		TheBrowseByFields = &bbf
	}
	return err
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

	//TODO(dan): upload xlsx file to S3, etc…

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

	//TODO(dan): calculate useful lists – changes, unchanged entries, etc…

	c.JSON(http.StatusOK, ImportTryResponse{
		TotalEntries: len(newEntries.Entries),
		NewEntries:   len(newEntries.Entries) - len(existingEntries),
		Entries:      newEntries.Entries,
		OldEntries:   existingEntries,
	})
}

func getYpsDbs(c *gin.Context) {
}

type SearchEntry struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	Authors            string    `json:"authors"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	DocumentType       string    `json:"document_type"`
	AvailableLanguages []string  `json:"available_languages"`
	Language           string    `json:"language"`
}

type SearchEntriesResponse struct {
	Entries    []SearchEntry `json:"entries"`
	TotalPages int           `json:"total_pages"`
}

func getEntry(c *gin.Context) {
}

type BrowseByFieldValues map[string][]string

type BrowseByFieldsResponse struct {
	Values BrowseByFieldValues `json:"values"`
}

func getBrowseByFields(c *gin.Context) {
	c.JSON(http.StatusOK, BrowseByFieldsResponse{
		Values: *TheBrowseByFields,
	})
}
