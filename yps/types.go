package yps

import (
	"slices"
	"time"
)

// entries

type Entry struct {
	ItemID          string    `json:"id"`
	Title           string    `json:"title"`
	Authors         string    `json:"authors"`
	URL             string    `json:"url"`
	OrgPublishers   []string  `json:"orgs"`
	OrgDocID        string    `json:"org_doc_id"`
	OrgType         string    `json:"org_type"`
	DocType         string    `json:"entry_type"`
	Abstract        string    `json:"abstract"`
	YouthLed        string    `json:"youth_led"`
	YouthLedDetails string    `json:"youth_led_details"`
	Keywords        []string  `json:"keywords"`
	Regions         []string  `json:"regions"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Language        string    `json:"language"`
	AltLanguageIDs  []string  `json:"alt_language_ids"`
	RelatedIDs      []string  `json:"related_ids"`
}

type EntryFile struct {
	Filename string `json:"filename"`
	URL      string `json:"url"`
}

type LookedUpAltLanguageEntry struct {
	Language string      `json:"language"`
	Title    string      `json:"title"`
	Files    []EntryFile `json:"files"`
}

type LookedUpEntry struct {
	Entry Entry `json:"entry"`

	Files []EntryFile `json:"files"`

	Alternates map[string]LookedUpAltLanguageEntry `json:"alternates"`
}

func (luEntry *LookedUpEntry) AsEntryResponse() (response GetEntryResponse) {
	response.Entry = luEntry.Entry
	response.Files = luEntry.Files
	response.Alternates = luEntry.Alternates
	return response
}

type SearchEntry struct {
	ID                 string    `json:"id"`
	Title              string    `json:"title"`
	Authors            string    `json:"authors"`
	StartDate          time.Time `json:"start_date"`
	EndDate            time.Time `json:"end_date"`
	DocumentType       string    `json:"document_type"`
	AvailableLanguages []string  `json:"available_languages"`
	Regions            []string  `json:"regions"`
	Language           string    `json:"language"`
}

type XlsxEntry struct {
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
	Regions         []string
	StartDate       string
	EndDate         string
	Language        string
	rawLanguages    []string
	AltLanguageIDs  []string
	RelatedIDs      []string
}

func (newEntry *XlsxEntry) Matches(oldEntry Entry) bool {
	return (newEntry.Title == oldEntry.Title &&
		newEntry.Authors == oldEntry.Authors &&
		newEntry.URL == oldEntry.URL &&
		slices.Equal(newEntry.OrgPublishers, oldEntry.OrgPublishers) &&
		newEntry.OrgDocID == oldEntry.OrgDocID &&
		newEntry.OrgType == oldEntry.OrgType &&
		newEntry.DocType == oldEntry.DocType &&
		newEntry.Abstract == oldEntry.Abstract &&
		newEntry.YouthLed == oldEntry.YouthLed &&
		newEntry.YouthLedDetails == oldEntry.YouthLedDetails &&
		slices.Equal(newEntry.Keywords, oldEntry.Keywords) &&
		slices.Equal(newEntry.Regions, oldEntry.Regions) &&
		// newEntry.StartDate == oldEntry.StartDate.Format("2006-01-02") &&
		// newEntry.EndDate == oldEntry.EndDate.Format("2006-01-02") &&
		newEntry.Language == oldEntry.Language &&
		slices.Equal(newEntry.RelatedIDs, oldEntry.RelatedIDs))
}

// others

type ypsDbInfo struct {
	NumberOfEntries   int
	NumberOfLanguages int
}

type Page struct {
	Content      string
	GoogleFormID string
	Updated      time.Time
}

type BrowseByFieldValues map[string][]string
