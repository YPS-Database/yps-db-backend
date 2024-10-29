package ypsc

import "strings"

type ColumnType int

const (
	None ColumnType = iota
	ItemID
	Authors
	Year
	Title
	OrgPublisher
	DocNumber
	DayMonth
	URL
	Languages
	AlternateLanguageEntries
	RelatedEntries
	YouthInvolvement
	Abstract
	OrgType
	DocType
	Keywords
	RegionEastSouthAfrica
	RegionEastCentralAsia
	RegionSouthEastAsiaPacific
	RegionEuropeEurasia
	RegionLatinAmericaCaribbean
	RegionMiddleEastNorthAfrica
	RegionNorthAmerica
	RegionSouthAsia
	RegionWestCentralAfrica
	RegionGlobal
	RegionNA
)

func (ct ColumnType) String() string {
	return [...]string{"None", "Item", "Authors", "Year", "Title", "Org / Publisher", "Doc #", "Day / Month", "URL", "Languages available", "Alternate Languages", "Related Documents", "Youth-led/ authored", "Abstract/ Exec Summary", "Type of org", "Type of document", "Keywords", "East and Southern Africa", "East and Central Asia", "Southeast Asia and the Pacific", "Europe and Eurasia", "Latin America and the Caribbean", "Middle East and North Africa", "North America", "South Asia", "West and Central Africa", "Global", "N/A"}[ct]
}

var RequiredColumns = []ColumnType{
	ItemID, Authors, Year, Title, OrgPublisher, DocNumber, DayMonth, URL, Languages,
	AlternateLanguageEntries, RelatedEntries, YouthInvolvement, Abstract, OrgType, DocType,
	Keywords, RegionEastSouthAfrica, RegionEastCentralAsia, RegionSouthEastAsiaPacific,
	RegionEuropeEurasia, RegionLatinAmericaCaribbean, RegionMiddleEastNorthAfrica,
	RegionNorthAmerica, RegionSouthAsia, RegionWestCentralAfrica, RegionGlobal, RegionNA,
}

func ColumnNames(input ...ColumnType) string {
	var names []string

	for _, ct := range input {
		names = append(names, ct.String())
	}

	return strings.Join(names, ", ")
}
