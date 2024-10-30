package yps

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	ypsc "github.com/YPS-Database/yps-db-backend/yps/columns"
	ypsl "github.com/YPS-Database/yps-db-backend/yps/languages"
	"github.com/thedatashed/xlsxreader"
)

type EntriesXLSX struct {
	file    *xlsxreader.XlsxFile
	Entries map[string]xlsxEntry
}

func simplifyColumnName(input string) string {
	return strings.TrimSpace(strings.ToLower(input))
}

func trimSpacesOnItemsSkipZero(input []string) (output []string) {
	for _, entry := range input {
		// skip empty items
		if strings.TrimSpace(entry) != "" && entry != "0" {
			output = append(output, strings.TrimSpace(entry))
		}
	}
	return output
}

func getCellValue(row xlsxreader.Row, column string) string {
	for _, cell := range row.Cells {
		if cell.Column == column {
			return cell.Value
		}
	}
	return ""
}

var monthNameToNumber = map[string]int{
	"january":   1,
	"february":  2,
	"march":     3,
	"april":     4,
	"may":       5,
	"june":      6,
	"july":      7,
	"august":    8,
	"september": 9,
	"october":   10,
	"november":  11,
	"december":  12,
}

type xlsxEntry struct {
	ItemID         string
	Title          string
	Authors        string
	URL            string
	OrgPublishers  []string
	OrgDocID       string
	OrgType        string
	DocType        string
	Abstract       string
	YouthLed       string
	Keywords       []string
	StartDate      string
	EndDate        string
	Language       string
	rawLanguages   []string
	AltLanguageIDs []string
	RelatedIDs     []string
}

func ReadEntriesFile(input io.Reader) (*EntriesXLSX, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(input)

	file, err := xlsxreader.NewReader(buf.Bytes())
	if err != nil {
		return nil, err
	}

	var entries EntriesXLSX
	entries.file = file
	entries.Entries = make(map[string]xlsxEntry)

	if len(entries.file.Sheets) != 1 {
		return nil, errors.New("file must have exactly one sheet")
	}

	// read entries columns
	cols := make(map[ypsc.ColumnType]string)

	for row := range file.ReadRows(file.Sheets[0]) {
		if row.Error != nil {
			return nil, fmt.Errorf("error on row [%d]: %s", row.Index, row.Error.Error())
		}

		// processing the first row
		if len(cols) < 1 {
			for _, cell := range row.Cells {
				// get cell type
				name := simplifyColumnName(cell.Value)
				thisColumnType := ypsc.None
				if name == "item" {
					thisColumnType = ypsc.ItemID
				} else if strings.HasPrefix(name, "author") {
					thisColumnType = ypsc.Authors
				} else if name == "year" {
					thisColumnType = ypsc.Year
				} else if name == "title" {
					thisColumnType = ypsc.Title
				} else if strings.Contains(name, "publisher") {
					thisColumnType = ypsc.OrgPublisher
				} else if strings.HasPrefix(name, "doc #") {
					thisColumnType = ypsc.DocNumber
				} else if strings.Contains(name, "month") {
					thisColumnType = ypsc.DayMonth
				} else if name == "url" {
					thisColumnType = ypsc.URL
				} else if name == "languages available" {
					thisColumnType = ypsc.Languages
				} else if name == "alternate languages" {
					thisColumnType = ypsc.AlternateLanguageEntries
				} else if name == "related documents" {
					thisColumnType = ypsc.RelatedEntries
				} else if strings.HasPrefix(name, "youth-led") {
					thisColumnType = ypsc.YouthInvolvement
				} else if strings.HasPrefix(name, "abstract") {
					thisColumnType = ypsc.Abstract
				} else if strings.HasPrefix(name, "type of org") {
					thisColumnType = ypsc.OrgType
				} else if strings.HasPrefix(name, "type of document") {
					thisColumnType = ypsc.DocType
				} else if strings.HasPrefix(name, "keywords") {
					thisColumnType = ypsc.Keywords
				} else if name == "east and southern africa" {
					thisColumnType = ypsc.RegionEastSouthAfrica
				} else if name == "east and central asia" {
					thisColumnType = ypsc.RegionEastCentralAsia
				} else if name == "southeast asia and the pacific" {
					thisColumnType = ypsc.RegionSouthEastAsiaPacific
				} else if name == "europe and eurasia" {
					thisColumnType = ypsc.RegionEuropeEurasia
				} else if name == "latin america and the caribbean" {
					thisColumnType = ypsc.RegionLatinAmericaCaribbean
				} else if name == "middle east and north africa" {
					thisColumnType = ypsc.RegionMiddleEastNorthAfrica
				} else if name == "north america" {
					thisColumnType = ypsc.RegionNorthAmerica
				} else if name == "south asia" {
					thisColumnType = ypsc.RegionSouthAsia
				} else if name == "west and central africa" {
					thisColumnType = ypsc.RegionWestCentralAfrica
				} else if name == "global" {
					thisColumnType = ypsc.RegionGlobal
				} else if name == "n/a" {
					thisColumnType = ypsc.RegionNA
				}

				if thisColumnType != ypsc.None {
					cols[thisColumnType] = cell.Column
					// fmt.Println("Mapping", thisColumnType, "to column", cell.ColumnIndex(), "/", cell.Column)
				}
			}

			// confirm that all required column types are defined
			var missingColumnTypes []ypsc.ColumnType
			for _, columnType := range ypsc.RequiredColumns {
				_, exists := cols[columnType]
				if !exists {
					missingColumnTypes = append(missingColumnTypes, columnType)
				}
			}
			if len(missingColumnTypes) > 0 {
				return nil, errors.New("Cannot find columns: " + ypsc.ColumnNames(missingColumnTypes...))
			}

			continue
		}

		// skip empty rows
		if strings.TrimSpace(getCellValue(row, cols[ypsc.Title])) == "" || getCellValue(row, cols[ypsc.Title]) == "0" {
			continue
		}

		// read the rest of the rows
		var itemID = getCellValue(row, cols[ypsc.ItemID])
		_, alreadyExists := entries.Entries[itemID]
		if alreadyExists {
			return nil, fmt.Errorf("duplicate item ID found on row %d: %s", row.Index, itemID)
		}

		// simple columns
		var title = strings.TrimSpace(getCellValue(row, cols[ypsc.Title]))
		var authors = strings.TrimSpace(getCellValue(row, cols[ypsc.Authors]))
		var orgdocid = strings.TrimSpace(getCellValue(row, cols[ypsc.DocNumber]))
		var url = strings.TrimSpace(getCellValue(row, cols[ypsc.URL]))
		var abstract = strings.TrimSpace(getCellValue(row, cols[ypsc.Abstract]))
		var youthled = strings.TrimSpace(getCellValue(row, cols[ypsc.YouthInvolvement]))
		var orgtype = strings.TrimSpace(getCellValue(row, cols[ypsc.OrgType]))
		var doctype = strings.TrimSpace(getCellValue(row, cols[ypsc.DocType]))

		var orgpublishers = trimSpacesOnItemsSkipZero(strings.Split(getCellValue(row, cols[ypsc.OrgPublisher]), ";"))
		var keywords = trimSpacesOnItemsSkipZero(strings.Split(getCellValue(row, cols[ypsc.Keywords]), ";"))
		var altlangIDs = trimSpacesOnItemsSkipZero(strings.Split(getCellValue(row, cols[ypsc.AlternateLanguageEntries]), ","))
		var relatedIDs = trimSpacesOnItemsSkipZero(strings.Split(getCellValue(row, cols[ypsc.RelatedEntries]), ","))

		// doc number needs to be removed if N/A
		var docnumber = strings.TrimSpace(getCellValue(row, cols[ypsc.DocNumber]))
		if docnumber == "N/A" {
			docnumber = ""
		}

		// languages need special handling
		var langs []string
		for _, languageName := range strings.Split(getCellValue(row, cols[ypsc.Languages]), ",") {
			languageName = strings.TrimSpace(languageName)
			if languageName == "" {
				continue
			}
			languageCode, err := ypsl.GetCode(languageName)
			if err != nil {
				return nil, fmt.Errorf("language error on item %s: %s", itemID, err.Error())
			}
			langs = append(langs, languageCode)
		}

		//TODO(dan): process regions

		// start and end dates
		var startDate, endDate string
		rawYear := strings.TrimSpace(getCellValue(row, cols[ypsc.Year]))
		rawDayMonth := strings.TrimSpace(getCellValue(row, cols[ypsc.DayMonth]))
		if rawYear != "N/A" {
			// try scanning the time. this is a quirk with how dates are entered in the db file
			t, basicScanErr := time.Parse("2006-01-02", rawDayMonth)

			if basicScanErr == nil {
				startDate = fmt.Sprintf("%s-%s", rawYear, t.Format("01-02"))
			} else if rawDayMonth == "N/A" {
				startDate = fmt.Sprintf("%s-01-01", rawYear)
			} else if monthNameToNumber[strings.ToLower(rawDayMonth)] != 0 {
				startDate = fmt.Sprintf("%s-%d-01", rawYear, monthNameToNumber[strings.ToLower(rawDayMonth)])
			} else {
				fmt.Println("can't process", rawDayMonth, "- skipping this date for the import")
				startDate = fmt.Sprintf("%s-01-01", rawYear)
			}

			if endDate == "" {
				endDate = startDate
			}
		}

		// insert into our list of processed entries
		newEntry := xlsxEntry{
			ItemID:         itemID,
			Title:          title,
			Authors:        authors,
			URL:            url,
			OrgPublishers:  orgpublishers,
			OrgDocID:       orgdocid,
			OrgType:        orgtype,
			DocType:        doctype,
			Abstract:       abstract,
			YouthLed:       youthled,
			Keywords:       keywords,
			StartDate:      startDate,
			EndDate:        endDate,
			rawLanguages:   langs,
			AltLanguageIDs: altlangIDs,
			RelatedIDs:     relatedIDs,
		}
		if len(langs) == 1 {
			newEntry.Language = langs[0]
		}
		entries.Entries[itemID] = newEntry
	}

	// post-processing
	for id, entry := range entries.Entries {
		// confirm related documents exist
		for _, altID := range entry.RelatedIDs {
			_, exists := entries.Entries[altID]
			if !exists {
				return nil, fmt.Errorf("item %s lists [%s] as a related item, but item [%s] does not exist", id, altID, altID)
			}
		}

		// set correct language for main entry
		if entry.Language != "" {
			continue
		}

		allAlternates := []string{id}
		languagesToRemove := make(map[string]bool)
		for _, altID := range entry.AltLanguageIDs {
			// skip self being listed in alt language ids
			if altID == id {
				continue
			}

			altEntry, exists := entries.Entries[altID]
			if !exists {
				return nil, fmt.Errorf("item %s lists %s as an alternate language, but item %s does not exist", id, altID, altID)
			}
			if altEntry.Language == "" {
				return nil, fmt.Errorf("item %s is an alternate, and must have only a single language defined", altID)
			}
			allAlternates = append(allAlternates, altID)
			languagesToRemove[altEntry.Language] = true
		}

		var finalLanguages []string
		for _, lang := range entry.rawLanguages {
			if languagesToRemove[lang] {
				continue
			}
			finalLanguages = append(finalLanguages, lang)
		}

		if len(finalLanguages) != 1 {
			return nil, fmt.Errorf("cannot work out which language item %s is, please confirm the alternates list is correct", id)
		}

		entry.Language = finalLanguages[0]

		// set all alternates on all entries
		entry.AltLanguageIDs = allAlternates

		for _, altID := range entry.AltLanguageIDs {
			// skip self being listed in alt language ids
			if altID == id {
				continue
			}

			altEntry, exists := entries.Entries[altID]
			if !exists {
				return nil, fmt.Errorf("item %s lists %s as an alternate language, but item %s does not exist", id, altID, altID)
			}

			altEntry.AltLanguageIDs = allAlternates
			entries.Entries[altEntry.ItemID] = altEntry
		}

		// post processing finished for this item, hooray
		entries.Entries[id] = entry
	}

	return &entries, nil
}
