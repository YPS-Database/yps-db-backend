package yps

import (
	"bytes"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ImportTryResponse struct {
	TotalEntries      int      `json:"total_entries"`
	UnmodifiedEntries int      `json:"unmodified_entries"`
	ModifiedEntries   int      `json:"modified_entries"`
	NewEntries        int      `json:"new_entries"`
	DeletedEntries    int      `json:"deleted_entries"`
	Nits              []string `json:"nits"`
	FileAlreadyExists bool     `json:"file_already_exists"`
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
	overwriteRaw, exists := c.GetQuery("overwrite")
	overwrite := exists && overwriteRaw == "true"

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

	exists, err = TheS3.FileExists(s3fn)
	if err != nil {
		fmt.Println("Could not check file existence:", err.Error())
		c.JSON(400, gin.H{"error": "Could not check whether db file exists."})
		return
	}
	if exists && !overwrite {
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

	Log(LogLevelInfo, "database-update", "Applied database update", map[string]string{
		"filename": fileHeader.Filename,
	})

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

	alreadyExists, err := TheS3.FileExists(fmt.Sprintf("dbs/%s", fileHeader.Filename))
	if err != nil {
		fmt.Println("Could not check file existence:", err.Error())
		c.JSON(400, gin.H{"error": "Could not check whether db file exists."})
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

	var unmodifiedEntriesCount, modifiedEntriesCount, newEntriesCount, deletedEntriesCount int
	for id, newEntry := range newEntries.Entries {
		oldEntry, exists := existingEntries[id]
		if !exists {
			newEntriesCount += 1
			continue
		}

		if newEntry.Matches(oldEntry) {
			unmodifiedEntriesCount += 1
		} else {
			modifiedEntriesCount += 1
		}
	}

	for id := range existingEntries {
		_, exists := newEntries.Entries[id]
		if !exists {
			deletedEntriesCount += 1
		}
	}

	response := ImportTryResponse{
		TotalEntries:      len(newEntries.Entries),
		UnmodifiedEntries: unmodifiedEntriesCount,
		ModifiedEntries:   modifiedEntriesCount,
		NewEntries:        newEntriesCount,
		DeletedEntries:    deletedEntriesCount,
		Nits:              newEntries.Nits,
		FileAlreadyExists: alreadyExists,
	}

	Log(LogLevelInfo, "database-update-test", "Tested database update", response)

	c.JSON(http.StatusOK, response)
}

type DeleteYpsDbRequest struct {
	ID string `uri:"slug" binding:"required"`
}

func deleteYpsDb(c *gin.Context) {
	var req DeleteYpsDbRequest
	if err := c.ShouldBindUri(&req); err != nil {
		fmt.Println("Could not get entry URI binding:", err.Error())
		c.JSON(400, gin.H{"error": "Entry must be given"})
		return
	}

	err := TheDb.RemoveDbFile(req.ID)
	if err != nil {
		fmt.Println("Could not delete db file:", err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	Log(LogLevelInfo, "database-delete", "Deleted database file", map[string]string{
		"database_id": req.ID,
	})

	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

type GetDbFilesResponse struct {
	Files []DbFile `json:"files"`
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

type GetLatestDbResponse struct {
	NumberOfEntries   int `json:"number_of_entries"`
	NumberOfLanguages int `json:"number_of_languages"`
}

func getLatestYpsDb(c *gin.Context) {
	info, err := TheDb.GetLatestDbInfo()
	if err != nil {
		fmt.Println("Could not get db info:", err.Error())
		c.JSON(400, gin.H{"error": "Could not get db info"})
		return
	}

	c.JSON(http.StatusOK, GetLatestDbResponse(info))
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

type UploadEntryFileRequest struct {
	ID string `uri:"slug" binding:"required"`
}

func uploadEntryFile(c *gin.Context) {
	var req UploadEntryFileRequest
	if err := c.ShouldBindUri(&req); err != nil {
		fmt.Println("Could not get entry URI binding:", err.Error())
		c.JSON(400, gin.H{"error": "Entry must be given"})
		return
	}

	// load passed upload file
	fileHeader, err := c.FormFile("upload")
	if err != nil {
		fmt.Println("Could not get file from uploadEntryFile call:", err.Error())
		c.JSON(400, gin.H{"error": "Could not get 'upload' file in form body."})
		return
	}

	file, err := fileHeader.Open()
	if err != nil {
		fmt.Println("Could not open file from uploadEntryFile call:", err.Error())
		c.JSON(400, gin.H{"error": "Could not open 'upload' file in form body."})
		return
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(file)

	s3fn := TheS3.EntryFileKey(req.ID, fileHeader.Filename)
	fmt.Println(s3fn)

	uploaded, err := TheS3.Upload(s3fn, bytes.NewReader(buf.Bytes()))
	if err != nil {
		fmt.Println("Could not upload new entry file:", err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = TheDb.AddEntryFile(req.ID, fileHeader.Filename, uploaded.URL)
	if err != nil {
		fmt.Println("Could not add new entry file:", err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	Log(LogLevelInfo, "entry-file-upload", "Added file to entry", map[string]string{
		"entry":    req.ID,
		"filename": fileHeader.Filename,
	})

	c.JSON(http.StatusCreated, gin.H{"ok": true})
}

type DeleteEntryFileRequest struct {
	ID string `uri:"slug" binding:"required"`
}

type DeleteEntryFileParams struct {
	Filename string `json:"file" binding:"required"`
}

func deleteEntryFile(c *gin.Context) {
	var req DeleteEntryFileRequest
	if err := c.ShouldBindUri(&req); err != nil {
		fmt.Println("Could not get entry URI binding:", err.Error())
		c.JSON(400, gin.H{"error": "Entry must be given"})
		return
	}

	var params DeleteEntryFileParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	fmt.Println("Deleting file", params.Filename, "from item", req.ID)

	s3fn := TheS3.EntryFileKey(req.ID, params.Filename)
	err := TheS3.Delete(s3fn)
	if err != nil {
		fmt.Println("Could not get delete entry from S3:", err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	err = TheDb.RemoveEntryFile(req.ID, params.Filename)
	if err != nil {
		fmt.Println("Could not get remove entry from db:", err.Error())
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	Log(LogLevelInfo, "entry-file-delete", "Deleted file from entry", map[string]string{
		"entry":    req.ID,
		"filename": params.Filename,
	})

	c.JSON(http.StatusOK, gin.H{"ok": true})
}

type BrowseByFieldsResponse struct {
	Values BrowseByFieldValues `json:"values"`
}

func getBrowseByFields(c *gin.Context) {
	c.JSON(http.StatusOK, BrowseByFieldsResponse{
		Values: *TheBrowseByFields,
	})
}

type ImportFileListParams struct {
	Entries map[string][]string `json:"entries" binding:"required"`
}

func importFileList(c *gin.Context) {
	var params ImportFileListParams
	if err := c.ShouldBindJSON(&params); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err := TheDb.ImportFileList(FileList(params))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	Log(LogLevelInfo, "entry-files", "Imported new bulk file list", nil)

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
