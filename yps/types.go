package yps

import "time"

// entries

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

type EntryFile struct {
	Filename string
	URL      string
}

type LookedUpEntry struct {
	Item Entry

	// id to title
	AlternateLanguages map[int]string

	// all files, this language and others
	// map of language code -> list of files
	Files map[string][]EntryFile
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
	StartDate       string
	EndDate         string
	Language        string
	rawLanguages    []string
	AltLanguageIDs  []string
	RelatedIDs      []string
}

// others

type Page struct {
	Content      string
	GoogleFormID string
	Updated      time.Time
}

type BrowseByFieldValues map[string][]string
