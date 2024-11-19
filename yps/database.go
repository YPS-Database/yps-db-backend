package yps

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"io"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	ypsl "github.com/YPS-Database/yps-db-backend/yps/languages"
	uuid "github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

const DefaultEntriesPerPage = 30

type YPSDatabase struct {
	pool *pgxpool.Pool
}

var TheDb *YPSDatabase

func OpenDatabase(connectionUrl string) error {
	pool, err := pgxpool.New(context.Background(), connectionUrl)
	if err != nil {
		return err
	}

	TheDb = &YPSDatabase{
		pool,
	}

	return nil
}

func (db *YPSDatabase) Close() {
	db.pool.Close()
}

// db files

func (db *YPSDatabase) UploadDbFile(filename string, body io.Reader) error {
	uploaded, err := TheS3.Upload(filename, body)
	if err != nil {
		return err
	}

	// check for existing db file
	var existingId string
	err = db.pool.QueryRow(context.Background(), `
select id
from spreadsheet_files
where filename=$1
`, uploaded.Filename).Scan(&existingId)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return err
	}

	// update existing if exists, else upload it brand new
	id := existingId
	if id == "" {
		newId, err := uuid.NewV7()
		if err != nil {
			return err
		}
		id = newId.String()
	}

	// upload file to db
	_, err = db.pool.Exec(context.Background(), `
insert into spreadsheet_files (id, filename, url, added_at)
values ($1, $2, $3, $4)
on conflict (id)
do update
set
	filename=excluded.filename,
	url=excluded.url,
	added_at=excluded.added_at
`, id, uploaded.Filename, uploaded.URL, time.Now())
	return err
}

func (db *YPSDatabase) GetLatestDbInfo() (info ypsDbInfo, err error) {
	err = db.pool.QueryRow(context.Background(), `
select count(*) from entries
`).Scan(&info.NumberOfEntries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Count QueryRow failed: %v\n", err)
		return info, err
	}

	err = db.pool.QueryRow(context.Background(), `
select count(distinct entry_language) from entries
`).Scan(&info.NumberOfLanguages)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Count QueryRow failed: %v\n", err)
		return info, err
	}

	return info, nil
}

func (db *YPSDatabase) GetDbFiles() (files []DbFile, err error) {
	files = []DbFile{}

	rows, err := db.pool.Query(context.Background(), `
select id, filename, url
from spreadsheet_files
order by added_at desc
`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query for db files failed: %v\n", err)
		return files, err
	}
	defer rows.Close()

	for rows.Next() {
		var thisFile DbFile

		err = rows.Scan(&thisFile.ID, &thisFile.Filename, &thisFile.URL)
		if err != nil {
			return files, err
		}

		// remove all but the actual filename from the returned fn
		splitUp := strings.Split(thisFile.Filename, "/")
		thisFile.Filename = splitUp[len(splitUp)-1]

		files = append(files, thisFile)
	}

	return files, nil
}

// entries
//

func (db *YPSDatabase) GetBrowseByFields() (values BrowseByFieldValues, err error) {
	values = make(BrowseByFieldValues)

	// youth-led
	rows, err := db.pool.Query(context.Background(), `
select distinct youth_led from entries order by youth_led desc
`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Youth-led query failed: %v\n", err)
		return values, err
	}
	var youthLed []string
	for rows.Next() {
		var youth string
		err = rows.Scan(&youth)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not cast youth: %v\n", err)
			return values, err
		}
		youthLed = append(youthLed, youth)
	}
	rows.Close()
	if len(youthLed) > 0 {
		values["youth_led"] = youthLed
	}

	// year
	rows, err = db.pool.Query(context.Background(), `
select distinct DATE_PART('year', start_date) AS year from entries where start_date > '1800-01-01' order by year desc
`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Year query failed: %v\n", err)
		return values, err
	}
	var years []string
	for rows.Next() {
		var year int
		err = rows.Scan(&year)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not cast year: %v\n", err)
			return values, err
		}
		years = append(years, strconv.Itoa(year))
	}
	rows.Close()
	if len(years) > 0 {
		values["year"] = years
	}

	// entry type
	rows, err = db.pool.Query(context.Background(), `
select entry_type, count(*) as number_of_rows from entries group by entry_type order by entry_type asc
-- number_of_rows desc
`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Entry type query failed: %v\n", err)
		return values, err
	}
	var entryTypes []string
	for rows.Next() {
		var entryType string
		var count int
		err = rows.Scan(&entryType, &count)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not cast entry type or count: %v\n", err)
			return values, err
		}
		// if count < 10 {
		// 	break
		// }
		entryTypes = append(entryTypes, entryType)
	}
	rows.Close()
	if len(entryTypes) > 0 {
		values["entry_type"] = entryTypes
	}

	// regions
	rows, err = db.pool.Query(context.Background(), `
	select distinct unnest(regions) as region_name from entries order by region_name asc
	`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Entry type query failed: %v\n", err)
		return values, err
	}
	var regions []string
	for rows.Next() {
		var region string
		err = rows.Scan(&region)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Could not cast region name: %v\n", err)
			return values, err
		}
		if region != "Global" && region != "N/A" {
			regions = append(regions, region)
		}
	}
	rows.Close()
	if len(regions) > 0 {
		regions = append(regions, "Global")
		regions = append(regions, "N/A")
		values["region"] = regions
	}

	return values, err
}

func (db *YPSDatabase) GetAllEntries() (entries map[string]Entry, err error) {
	entries = make(map[string]Entry)

	rows, err := db.pool.Query(context.Background(), `
select id, url, entry_type, entry_language, start_date, end_date, alternates,
	related, title, authors, abstract, keywords, regions, orgs, org_doc_id, org_type,
	youth_led, youth_led_distilled
from entries
`)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query failed: %v\n", err)
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var e Entry

		err = rows.Scan(&e.ItemID, &e.URL, &e.DocType, &e.Language, &e.StartDate, &e.EndDate,
			&e.AltLanguageIDs, &e.RelatedIDs, &e.Title, &e.Authors, &e.Abstract, &e.Keywords,
			&e.Regions, &e.OrgPublishers, &e.OrgDocID, &e.OrgType, &e.YouthLedDetails, &e.YouthLed)
		if err != nil {
			return entries, err
		}
		entries[e.ItemID] = e
	}

	return entries, err
}

func (db *YPSDatabase) GetSingleEntry(id string) (entry LookedUpEntry, err error) {
	entry.Files = []EntryFile{}
	entry.Alternates = make(map[string]LookedUpAltLanguageEntry)
	entry.Related = make(map[string]string)

	// get the main entry
	err = db.pool.QueryRow(context.Background(), `
select id, url, entry_type, entry_language, start_date, end_date, alternates, related, title, authors, abstract, keywords, regions, orgs, org_doc_id, org_type, youth_led, youth_led_distilled
from entries
where id=$1
`, id).Scan(
		&entry.Entry.ItemID, &entry.Entry.URL, &entry.Entry.DocType, &entry.Entry.Language, &entry.Entry.StartDate,
		&entry.Entry.EndDate, &entry.Entry.AltLanguageIDs, &entry.Entry.RelatedIDs, &entry.Entry.Title,
		&entry.Entry.Authors, &entry.Entry.Abstract, &entry.Entry.Keywords, &entry.Entry.Regions,
		&entry.Entry.OrgPublishers, &entry.Entry.OrgDocID, &entry.Entry.OrgType, &entry.Entry.YouthLedDetails,
		&entry.Entry.YouthLed,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow for entry failed: %v\n", err)
		return entry, err
	}

	// get the alternate languages
	rows, err := db.pool.Query(context.Background(), `
select id, entry_language, title
from entries
where id=any($1)
`, entry.Entry.AltLanguageIDs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query for alt rows failed: %v\n", err)
		return entry, err
	}
	for rows.Next() {
		var leID string
		var le LookedUpAltLanguageEntry
		le.Files = []EntryFile{}

		err = rows.Scan(&leID, &le.Language, &le.Title)
		if err != nil {
			return entry, err
		}

		if leID != id {
			entry.Alternates[leID] = le
		}
	}
	rows.Close()

	// get all files
	all_ids := slices.Concat(entry.Entry.AltLanguageIDs, []string{id}) // not necessary, but just in case

	rows, err = db.pool.Query(context.Background(), `
select entry_id, filename, url
from entry_files
where entry_id=any($1)
`, all_ids)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query for files failed: %v\n", err)
		return entry, err
	}
	for rows.Next() {
		var entryID, filename, url string

		err = rows.Scan(&entryID, &filename, &url)
		if err != nil {
			return entry, err
		}

		newFile := EntryFile{
			Filename: filename,
			URL:      url,
		}
		if entryID == id {
			entry.Files = append(entry.Files, newFile)
		} else {
			for leID, le := range entry.Alternates {
				if entryID == leID {
					le.Files = append(le.Files, newFile)
					entry.Alternates[leID] = le
				}
			}
		}
	}
	rows.Close()

	// get the related entries
	rows, err = db.pool.Query(context.Background(), `
select id, title
from entries
where id=any($1)
`, entry.Entry.RelatedIDs)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Query for alt rows failed: %v\n", err)
		return entry, err
	}
	for rows.Next() {
		var reID, reTitle string

		err = rows.Scan(&reID, &reTitle)
		if err != nil {
			return entry, err
		}

		if reID != id {
			entry.Related[reID] = reTitle
		}
	}
	rows.Close()

	//TODO(dan): look up related files for this language and others

	return entry, err
}

func (db *YPSDatabase) UploadEntries(entryMap map[string]XlsxEntry) error {
	// assemble rows slice
	var rows [][]any
	for id, entry := range entryMap {
		startDate, _ := time.Parse(time.DateOnly, entry.StartDate)
		endDate, _ := time.Parse(time.DateOnly, entry.EndDate)
		rows = append(rows, []any{
			id, entry.URL, entry.DocType, entry.Language, startDate, endDate,
			entry.AltLanguageIDs, entry.RelatedIDs, entry.Title, entry.Authors, entry.Abstract,
			entry.Keywords, entry.Regions, entry.OrgPublishers, entry.OrgDocID, entry.OrgType,
			entry.YouthLed, entry.YouthLedDetails,
		})
	}
	source := pgx.CopyFromRows(rows)

	// delete rows from temp table
	_, err := db.pool.Exec(context.Background(), `truncate table temp_insert_entries`)
	if err != nil {
		return err
	}

	fmt.Println("starting copy into temp table")
	_, err = db.pool.CopyFrom(context.Background(), pgx.Identifier{`temp_insert_entries`}, []string{
		"id", "url", "entry_type", "entry_language", "start_date", "end_date", "alternates",
		"related", "title", "authors", "abstract", "keywords", "regions", "orgs", "org_doc_id",
		"org_type", "youth_led_distilled", "youth_led"}, source)
	fmt.Println("ended copy into temp table")

	if err != nil {
		return err
	}

	// transfer rows to real table
	_, err = db.pool.Exec(context.Background(), `
insert into entries (id, url, entry_type, entry_language, start_date, end_date, alternates, related, title, authors, abstract, keywords, regions, orgs, org_doc_id, org_type, youth_led, youth_led_distilled)
select id, url, entry_type, entry_language, start_date, end_date, alternates, related, title, authors, abstract, keywords, regions, orgs, org_doc_id, org_type, youth_led, youth_led_distilled
from temp_insert_entries
on conflict (id)
do update
set
	url=excluded.url,
	entry_type=excluded.entry_type,
	entry_language=excluded.entry_language,
	start_date=excluded.start_date,
	end_date=excluded.end_date,
	alternates=excluded.alternates,
	related=excluded.related,
	title=excluded.title,
	authors=excluded.authors,
	abstract=excluded.abstract,
	keywords=excluded.keywords,
	regions=excluded.regions,
	orgs=excluded.orgs,
	org_doc_id=excluded.org_doc_id,
	org_type=excluded.org_type,
	youth_led=excluded.youth_led,
	youth_led_distilled=excluded.youth_led_distilled
`)
	if err != nil {
		return err
	}

	// remove rows in real table but not in temp table
	_, err = db.pool.Exec(context.Background(), `
delete from entries
where not exists (
	select from temp_insert_entries
	where temp_insert_entries."id" = entries."id"
)
`)
	if err != nil {
		return err
	}

	err = UpdateBrowseByFields()

	return err
}

func (db *YPSDatabase) RemoveDbFile(id string) (err error) {
	_, err = db.pool.Exec(context.Background(), `
delete from spreadsheet_files
where id = $1
`, id)

	return err
}

func (db *YPSDatabase) ImportFileList(fileList FileList) (err error) {
	// assemble rows slice
	var rows [][]any
	for id, entry := range fileList.Entries {
		for _, filename := range entry {
			rows = append(rows, []any{
				id, filename, TheS3.EntryFileURL(id, filename),
			})
		}
	}
	source := pgx.CopyFromRows(rows)

	// delete rows from temp table
	_, err = db.pool.Exec(context.Background(), `truncate table temp_insert_entry_files`)
	if err != nil {
		return err
	}

	fmt.Println("starting copy into temp table")
	_, err = db.pool.CopyFrom(context.Background(), pgx.Identifier{`temp_insert_entry_files`}, []string{
		"entry_id", "filename", "url"}, source)
	fmt.Println("ended copy into temp table")

	if err != nil {
		return err
	}

	// transfer rows to real table
	_, err = db.pool.Exec(context.Background(), `
	insert into entry_files (entry_id, filename, url)
	select entry_id, filename, url
	from temp_insert_entry_files
	on conflict (entry_id, filename)
	do update
	set
		url=excluded.url
	`)
	if err != nil {
		return err
	}

	return err
}

func (db *YPSDatabase) AddEntryFile(entry, filename, url string) (err error) {
	_, err = db.pool.Exec(context.Background(), `
insert into entry_files (entry_id, filename, url)
values ($1, $2, $3)
on conflict (entry_id, filename)
do update
set
	url=excluded.url
	`, entry, filename, url)

	return err
}

func (db *YPSDatabase) RemoveEntryFile(entry, filename string) (err error) {
	_, err = db.pool.Exec(context.Background(), `
delete from entry_files
where entry_id = $1 and filename = $2
`, entry, filename)

	return err
}

func (db *YPSDatabase) Search(params SearchRequest) (values SearchResponse, err error) {
	values.Entries = []SearchEntry{}
	values.Filters = []SearchFilter{}

	var assembledParams []any
	newParamNumber := 1
	var whereClauses []string

	rankQuery := `1`
	if strings.TrimSpace(params.Query) != "" {
		queryColumn := `alltextsearch_index_col`
		if params.SearchContext == "title" {
			queryColumn = `titlesearch_index_col`
		} else if params.SearchContext == "abstract" {
			queryColumn = `abstractsearch_index_col`
		}

		rankQuery = fmt.Sprintf(`ts_rank_cd(%s, websearch_to_tsquery('english', $%d))`, queryColumn, newParamNumber)
		whereClauses = append(whereClauses, fmt.Sprintf(`%s @@ websearch_to_tsquery('english', $%d)`, queryColumn, newParamNumber))
		assembledParams = append(assembledParams, params.Query)
		newParamNumber += 1
	}

	if params.EntryLanguage != "all" {
		whereClauses = append(whereClauses, fmt.Sprintf(`entry_language = $%d`, newParamNumber))
		assembledParams = append(assembledParams, params.EntryLanguage)
		newParamNumber += 1
	}

	// arbitrary filters
	if params.FilterKey == "entry_type" {
		whereClauses = append(whereClauses, fmt.Sprintf(`entry_type = $%d`, newParamNumber))
		assembledParams = append(assembledParams, params.FilterValue)
		newParamNumber += 1
	}
	if params.FilterKey == "year" {
		whereClauses = append(whereClauses, fmt.Sprintf(`date_part('year', start_date) = $%d`, newParamNumber))
		assembledParams = append(assembledParams, params.FilterValue)
		newParamNumber += 1
	}
	if params.FilterKey == "youth_led" {
		whereClauses = append(whereClauses, fmt.Sprintf(`youth_led = $%d`, newParamNumber))
		assembledParams = append(assembledParams, params.FilterValue)
		newParamNumber += 1
	}
	if params.FilterKey == "keyword" {
		whereClauses = append(whereClauses, fmt.Sprintf(`$%d = ANY(keywords)`, newParamNumber))
		assembledParams = append(assembledParams, params.FilterValue)
		newParamNumber += 1
	}
	if params.FilterKey == "region" {
		whereClauses = append(whereClauses, fmt.Sprintf(`$%d = ANY(regions)`, newParamNumber))
		assembledParams = append(assembledParams, params.FilterValue)
		newParamNumber += 1
	}

	var assembledWhereClause string
	if len(whereClauses) > 0 {
		assembledWhereClause = fmt.Sprintf(`where %s`, strings.Join(whereClauses, ` AND `))
	}

	// count total response rows for total pages
	assembledCountQuery := fmt.Sprintf(`
SELECT count(*)
FROM entries
%s
`, assembledWhereClause)

	fmt.Println("COUNT:", []any{assembledCountQuery, len(assembledParams), assembledParams})

	var totalEntries int
	err = db.pool.QueryRow(context.Background(), assembledCountQuery, assembledParams...).Scan(&totalEntries)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Count QueryRow failed: %v\n", err)
		return values, err
	}
	values.TotalEntries = totalEntries
	values.TotalPages = int(math.Ceil(float64(totalEntries) / float64(DefaultEntriesPerPage)))

	// youth-led filter
	assembledYouthLedQuery := fmt.Sprintf(`
SELECT youth_led, count(*)
FROM entries
%s
GROUP BY youth_led
ORDER BY youth_led desc
`, assembledWhereClause)

	rows, err := db.pool.Query(context.Background(), assembledYouthLedQuery, assembledParams...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Youth-led Query failed: %v\n", err)
		return values, err
	}

	youthLed := SearchFilter{
		Key: "youth_led",
	}
	for rows.Next() {
		var fv SearchFilterValue

		err = rows.Scan(&fv.Value, &fv.Count)
		if err != nil {
			return values, err
		}
		youthLed.Values = append(youthLed.Values, fv)
	}
	rows.Close()
	if len(youthLed.Values) > 0 {
		values.Filters = append(values.Filters, youthLed)
	}

	// search for actual entries
	startEntry := max(params.Page-1, 0) * DefaultEntriesPerPage
	values.StartEntry = startEntry + 1
	values.EndEntry = values.StartEntry + min(DefaultEntriesPerPage, values.TotalEntries-(values.StartEntry-1)) - 1
	if values.TotalEntries == 0 {
		values.StartEntry = 0
	}

	sortClause := `rank desc, start_date desc, id asc`
	if params.Sort == "dateasc" {
		sortClause = `start_date asc, id asc`
	} else if params.Sort == "datedesc" {
		sortClause = `start_date desc, id desc`
	} else if params.Sort == "abc" {
		sortClause = `title asc, start_date desc, id asc`
	}

	assembledSearchQuery := fmt.Sprintf(`
SELECT id, title, authors, start_date, end_date, entry_type, entry_language, array(select entry_language from entries e where e.id=entries.id or entries.id=ANY(alternates)) as languages, regions, %s AS rank
FROM entries
%s
ORDER BY %s
LIMIT %d
OFFSET %d
`, rankQuery, assembledWhereClause, sortClause, DefaultEntriesPerPage, startEntry)

	fmt.Println("SEARCH:", []any{assembledSearchQuery, len(assembledParams), assembledParams})

	rows, err = db.pool.Query(context.Background(), assembledSearchQuery, assembledParams...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Search Query failed: %v\n", err)
		return values, err
	}
	defer rows.Close()

	for rows.Next() {
		var e SearchEntry

		var rank float32
		var langCodes []string
		err = rows.Scan(&e.ID, &e.Title, &e.Authors, &e.StartDate, &e.EndDate, &e.DocumentType, &e.Language, &langCodes, &e.Regions, &rank)
		if err != nil {
			return values, err
		}
		e.Language = ypsl.GetName(e.Language)
		for _, lc := range langCodes {
			e.AvailableLanguages = append(e.AvailableLanguages, ypsl.GetName(lc))
		}
		slices.SortFunc(e.AvailableLanguages, func(a, b string) int {
			return strings.Compare(strings.ToLower(a), strings.ToLower(b))
		})
		values.Entries = append(values.Entries, e)
	}

	return values, nil
}

// dynamic pages
//

func (db *YPSDatabase) GetPage(id string) (page *Page, err error) {
	var rawPage Page

	err = db.pool.QueryRow(context.Background(), `
select content, google_form_id, updated_at
from pages
where id=$1
`, id).Scan(&rawPage.Content, &rawPage.GoogleFormID, &rawPage.Updated)
	if err != nil {
		fmt.Fprintf(os.Stderr, "QueryRow failed: %v\n", err)
		return nil, err
	}

	return &rawPage, nil
}

func (db *YPSDatabase) SetPage(id string, content string, googleFormID string) (err error) {
	_, err = db.pool.Exec(context.Background(), `
insert into pages (id, content, google_form_id, updated_at)
values ($1, $2, $3, default)
on conflict (id)
do update
set
	content=excluded.content,
	google_form_id=excluded.google_form_id,
	updated_at=default
`, id, content, googleFormID)
	return err
}
