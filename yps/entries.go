package yps

import (
	"bytes"
	"fmt"
	"net/http"

	ypss3 "github.com/YPS-Database/yps-db-backend/yps/s3"
	"github.com/gin-gonic/gin"
)

type ImportTryResponse struct {
	TotalEntries int                  `json:"total_entries"`
	NewEntries   int                  `json:"new_entries"`
	Entries      map[string]XlsxEntry `json:"entries"`
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
	buf := new(bytes.Buffer)
	buf.ReadFrom(file)

	s3fn := fmt.Sprintf("dbs/%s", fileHeader.Filename)

	exists, err := TheS3.FileExists(s3fn)
	if err != nil {
		fmt.Println("Could not check file existence:", err.Error())
		c.JSON(400, gin.H{"error": "Could not check whether db file exists."})
		return
	}
	if exists {
		c.JSON(400, gin.H{"error": "Spreadsheet with this filename already exists. Please rename the file before you upload it."})
		return
	}

	err = TheDb.UploadDbFile(s3fn, bytes.NewReader(buf.Bytes()))
	if err != nil {
		fmt.Println("Could not upload new db file:", err.Error())
		c.JSON(400, gin.H{"error": "Could not upload 'db' file from form body."})
		return
	}

	newEntries, err := ReadEntriesFile(bytes.NewReader(buf.Bytes()))
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

	exists, err := TheS3.FileExists(fmt.Sprintf("dbs/%s", fileHeader.Filename))
	if err != nil {
		fmt.Println("Could not check file existence:", err.Error())
		c.JSON(400, gin.H{"error": "Could not check whether db file exists."})
		return
	}
	if exists {
		c.JSON(400, gin.H{"error": "Spreadsheet with this filename already exists. Please rename the file before you upload it."})
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

type GetDbFilesResponse struct {
	Files []ypss3.S3Upload `json:"files"`
}

func getYpsDbs(c *gin.Context) {
	files, err := TheDb.GetDbFiles()
	if err != nil {
		fmt.Println("Could not get db files:", err.Error())
		c.JSON(400, gin.H{"error": "Could not get db files"})
		return
	}

	c.JSON(http.StatusOK, GetDbFilesResponse{
		Files: files,
	})
}

type GetEntryRequest struct {
	ID string `uri:"slug" binding:"required"`
}

type GetEntryResponse struct {
	LookedUpEntry
}

func getEntry(c *gin.Context) {
	var req GetEntryRequest
	if err := c.ShouldBindUri(&req); err != nil {
		fmt.Println("Could not get entry URI binding:", err.Error())
		c.JSON(400, gin.H{"error": "Entry must be given"})
		return
	}

	luEntry, err := TheDb.GetSingleEntry(req.ID)
	if err != nil {
		fmt.Println("Could not get entry:", req.ID, ":", err.Error())
		c.JSON(400, gin.H{"error": "Could not get entry"})
		return
	}

	var response GetEntryResponse = luEntry.AsEntryResponse()

	c.JSON(http.StatusOK, response)
}

type BrowseByFieldsResponse struct {
	Values BrowseByFieldValues `json:"values"`
}

func getBrowseByFields(c *gin.Context) {
	c.JSON(http.StatusOK, BrowseByFieldsResponse{
		Values: *TheBrowseByFields,
	})
}
