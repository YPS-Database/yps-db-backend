package yps

import "time"

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
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	Language        string    `json:"language"`
	AltLanguageIDs  []string  `json:"alt_language_ids"`
	RelatedIDs      []string  `json:"related_ids"`
}

type EntryFile struct {
	Filename string
	URL      string
}

type LookedUpAltLanguageEntry struct {
	Language string `json:"language"`
	Title    string `json:"title"`
}

type LookedUpEntry struct {
	Entry Entry `json:"entry"`

	// id to entry
	AlternateLanguages map[string]LookedUpAltLanguageEntry `json:"alternate_languages"`

	// all files, this language and others
	// map of entry id -> list of files
	Files map[string][]EntryFile `json:"files"`
}

func (luEntry *LookedUpEntry) AsEntryResponse() (response GetEntryResponse) {
	response.Entry = luEntry.Entry
	response.AlternateLanguages = luEntry.AlternateLanguages
	response.Files = luEntry.Files
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
