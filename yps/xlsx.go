package yps

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	ypsc "github.com/YPS-Database/yps-db-backend/yps/columns"
	ypsl "github.com/YPS-Database/yps-db-backend/yps/languages"
	"github.com/thedatashed/xlsxreader"
)

type EntriesXLSX struct {
	file    *xlsxreader.XlsxFile
	Entries map[string]XlsxEntry
	Nits    []string
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

func parseDayFromString(input string) (day string, err error) {
	for _, sub := range strings.Split(input, " ") {
		num, err := strconv.Atoi(sub)

		if err == nil && num > 0 && num < 31 {
			day = strconv.Itoa(num)
			return day, nil
		}
	}

	return day, errors.New("Could not parse day")
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

func ReadEntriesFile(input io.Reader) (*EntriesXLSX, error) {
	buf := new(bytes.Buffer)
	buf.ReadFrom(input)

	file, err := xlsxreader.NewReader(buf.Bytes())
	if err != nil {
		return nil, err
	}

	var entries EntriesXLSX
	entries.file = file
	entries.Entries = make(map[string]XlsxEntry)

	if len(entries.file.Sheets) < 1 {
		return nil, errors.New("no sheets found in the supplied database")
	}

	sheetToUse := 0
	for i, sheetName := range entries.file.Sheets {
		if strings.Contains(strings.ToLower(sheetName), "database") {
			sheetToUse = i
		}
	}

	entries.Nits = append(entries.Nits, fmt.Sprintf("Reading sheet %d (%s)", sheetToUse, file.Sheets[sheetToUse]))

	// read entries columns
	cols := make(map[ypsc.ColumnType]string)

	for row := range file.ReadRows(file.Sheets[sheetToUse]) {
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
				} else if strings.HasPrefix(name, "youth-led") || strings.HasPrefix(name, "youth authored") {
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
		var youthleddesc = strings.TrimSpace(getCellValue(row, cols[ypsc.YouthInvolvement]))
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

		// work out whether the description is youth led or not
		youthled := "Unknown"
		if strings.Contains(strings.ToLower(youthleddesc), "co-authored") || strings.Contains(strings.ToLower(youthleddesc), "co authored") {
			// this check is before No, because "No, co-authored with adults" and "Yes, co-authored with adults" should both be Co-authored
			youthled = "Co-authored"
		} else if strings.HasPrefix(strings.ToLower(youthleddesc), "yes") || strings.HasPrefix(strings.ToLower(youthleddesc), "youth-led") {
			youthled = "Yes"
		} else if strings.HasPrefix(strings.ToLower(youthleddesc), "no") {
			youthled = "No"
		} else if strings.HasPrefix(strings.ToLower(youthleddesc), "n/a") {
			youthled = "N/A"
		}
		if youthled == "Unknown" {
			entries.Nits = append(entries.Nits, fmt.Sprintf("[Item %s] Could not work out the 'youth-led' status, please make it start with 'Yes' or 'No', or include the text 'Co-authored'.", itemID))
		}

		// regions
		var regions []string
		if getCellValue(row, cols[ypsc.RegionEastSouthAfrica]) == "1" {
			regions = append(regions, "East and Southern Africa")
		}
		if getCellValue(row, cols[ypsc.RegionEastCentralAsia]) == "1" {
			regions = append(regions, "East and Central Asia")
		}
		if getCellValue(row, cols[ypsc.RegionSouthEastAsiaPacific]) == "1" {
			regions = append(regions, "Southeast Asia and the Pacific")
		}
		if getCellValue(row, cols[ypsc.RegionEuropeEurasia]) == "1" {
			regions = append(regions, "Europe and Eurasia")
		}
		if getCellValue(row, cols[ypsc.RegionLatinAmericaCaribbean]) == "1" {
			regions = append(regions, "Latin America and the Caribbean")
		}
		if getCellValue(row, cols[ypsc.RegionMiddleEastNorthAfrica]) == "1" {
			regions = append(regions, "Middle East and North Africa")
		}
		if getCellValue(row, cols[ypsc.RegionNorthAmerica]) == "1" {
			regions = append(regions, "North America")
		}
		if getCellValue(row, cols[ypsc.RegionSouthAsia]) == "1" {
			regions = append(regions, "South Asia")
		}
		if getCellValue(row, cols[ypsc.RegionWestCentralAfrica]) == "1" {
			regions = append(regions, "West and Central Africa")
		}
		if getCellValue(row, cols[ypsc.RegionGlobal]) == "1" {
			regions = append(regions, "Global")
		}
		if getCellValue(row, cols[ypsc.RegionNA]) == "1" {
			regions = append(regions, "N/A")
		}
		if len(regions) < 1 {
			regions = append(regions, "N/A")
			entries.Nits = append(entries.Nits, fmt.Sprintf("[Item %s] No regions defined, marking as N/A.", itemID))
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
				monthList := strings.Split(strings.ToLower(rawDayMonth), "-")
				if len(monthList) == 2 {
					var startMonth, endMonth string
					startDay := "01"
					endDay := "01"

					// parse month
					for monthName, monthNumber := range monthNameToNumber {
						if strings.Contains(monthList[0], monthName) {
							startMonth = strconv.Itoa(monthNumber)
						}
						if strings.Contains(monthList[1], monthName) {
							endMonth = strconv.Itoa(monthNumber)
						}

						if startMonth == "" && endMonth != "" {
							startMonth = endMonth
						}
						if endMonth == "" && startMonth != "" {
							endMonth = startMonth
						}
					}

					// parse day
					parsedDay, err := parseDayFromString(monthList[0])
					if err == nil {
						startDay = parsedDay
					}

					parsedDay, err = parseDayFromString(monthList[1])
					if err == nil {
						endDay = parsedDay
					} else if startMonth == endMonth {
						endDay = startDay
					}

					if startMonth != "" && endMonth != "" {
						startDate = fmt.Sprintf("%s-%s-%s", rawYear, startMonth, startDay)
						endDate = fmt.Sprintf("%s-%s-%s", rawYear, endMonth, endDay)
					}
				}
			}

			if startDate == "" {
				fmt.Println("can't process", rawDayMonth, "- skipping this date for the import")
				startDate = fmt.Sprintf("%s-01-01", rawYear)
				entries.Nits = append(entries.Nits, fmt.Sprintf("[Item %s] Could not work out the start/end dates.", itemID))
			}

			if endDate == "" {
				endDate = startDate
			}
		}

		// insert into our list of processed entries
		newEntry := XlsxEntry{
			ItemID:          itemID,
			Title:           title,
			Authors:         authors,
			URL:             url,
			OrgPublishers:   orgpublishers,
			OrgDocID:        orgdocid,
			OrgType:         orgtype,
			DocType:         doctype,
			Abstract:        abstract,
			YouthLed:        youthled,
			YouthLedDetails: youthleddesc,
			Keywords:        keywords,
			StartDate:       startDate,
			EndDate:         endDate,
			Regions:         regions,
			rawLanguages:    langs,
			AltLanguageIDs:  altlangIDs,
			RelatedIDs:      relatedIDs,
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
