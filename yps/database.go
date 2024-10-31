package yps

import (
	"context"
	"fmt"
	"math"
	"os"
	"slices"
	"strconv"
	"strings"
	"time"

	ypsl "github.com/YPS-Database/yps-db-backend/yps/languages"
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

func GetDatabase() *YPSDatabase {
	return TheDb
}

func (db *YPSDatabase) Close() {
	db.pool.Close()
}

// entries
//

func (db *YPSDatabase) GetBrowseByFields() (values BrowseByFieldValues, err error) {
	values = make(BrowseByFieldValues)

	// always set youth-led to have this order
	values["youth_led"] = []string{"Yes", "Co-authored", "No", "N/A"}

	// year
	rows, err := db.pool.Query(context.Background(), `
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
select entry_type, count(*) as number_of_rows from entries group by entry_type order by number_of_rows desc
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
		if count < 10 {
			break
		}
		entryTypes = append(entryTypes, entryType)
	}
	rows.Close()
	if len(entryTypes) > 0 {
		values["entry_type"] = entryTypes
	}

	return values, err
}

func (db *YPSDatabase) GetAllEntries() (entries map[string]Entry, err error) {
	entries = make(map[string]Entry)

	rows, err := db.pool.Query(context.Background(), `
select id, url, entry_type, entry_language, start_date, end_date, alternates,
	related, title, authors, abstract, keywords, orgs, org_doc_id, org_type, youth_led,
	youth_led_distilled
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
			&e.OrgPublishers, &e.OrgDocID, &e.OrgType, &e.YouthLedDetails, &e.YouthLed)
		if err != nil {
			return entries, err
		}
		entries[e.ItemID] = e
	}

	return entries, err
}

func (db *YPSDatabase) UploadEntries(entryMap map[string]xlsxEntry) error {
	// assemble rows slice
	var rows [][]any
	for id, entry := range entryMap {
		startDate, _ := time.Parse(time.DateOnly, entry.StartDate)
		endDate, _ := time.Parse(time.DateOnly, entry.EndDate)
		rows = append(rows, []any{
			id, entry.URL, entry.DocType, entry.Language, startDate, endDate,
			entry.AltLanguageIDs, entry.RelatedIDs, entry.Title, entry.Authors, entry.Abstract,
			entry.Keywords, entry.OrgPublishers, entry.OrgDocID, entry.OrgType, entry.YouthLed,
			entry.YouthLedDetails,
		})
	}
	source := pgx.CopyFromRows(rows)

	// delete rows from temp table on exiting this call
	_, err := db.pool.Exec(context.Background(), `truncate table temp_insert_entries`)
	if err != nil {
		return err
	}
	// defer db.pool.Exec(context.Background(), `truncate table temp_insert_entries`)

	fmt.Println("starting copy into temp table")
	_, err = db.pool.CopyFrom(context.Background(), pgx.Identifier{`temp_insert_entries`}, []string{
		"id", "url", "entry_type", "entry_language", "start_date", "end_date", "alternates",
		"related", "title", "authors", "abstract", "keywords", "orgs", "org_doc_id", "org_type",
		"youth_led_distilled", "youth_led"}, source)
	fmt.Println("ended copy into temp table")

	if err != nil {
		return err
	}

	// transfer rows to real table
	_, err = db.pool.Exec(context.Background(), `
insert into entries (id, url, entry_type, entry_language, start_date, end_date, alternates, related, title, authors, abstract, keywords, orgs, org_doc_id, org_type, youth_led, youth_led_distilled)
select id, url, entry_type, entry_language, start_date, end_date, alternates, related, title, authors, abstract, keywords, orgs, org_doc_id, org_type, youth_led, youth_led_distilled
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

func (db *YPSDatabase) Search(params SearchRequest) (values SearchResponse, err error) {
	values.Entries = []SearchEntry{}

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
		whereClauses = append(whereClauses, fmt.Sprintf(`youth_led_distilled = $%d`, newParamNumber))
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

	// search for actual entries
	startEntry := max(params.Page-1, 0) * DefaultEntriesPerPage
	values.StartEntry = startEntry + 1
	values.EndEntry = values.StartEntry + min(DefaultEntriesPerPage, values.TotalEntries-(values.StartEntry-1)) - 1
	if values.TotalEntries == 0 {
		values.StartEntry = 0
	}

	sortClause := `rank desc, id asc`
	if params.Sort == "dateasc" {
		sortClause = `start_date asc, id asc`
	} else if params.Sort == "datedesc" {
		sortClause = `start_date desc, id desc`
	} else if params.Sort == "abc" {
		sortClause = `title asc, id asc`
	}

	assembledSearchQuery := fmt.Sprintf(`
SELECT id, title, authors, start_date, end_date, entry_type, entry_language, array(select entry_language from entries e where e.id=entries.id or entries.id=ANY(alternates)) as languages, %s AS rank
FROM entries
%s
ORDER BY %s
LIMIT %d
OFFSET %d
`, rankQuery, assembledWhereClause, sortClause, DefaultEntriesPerPage, startEntry)

	fmt.Println("SEARCH:", []any{assembledSearchQuery, len(assembledParams), assembledParams})

	rows, err := db.pool.Query(context.Background(), assembledSearchQuery, assembledParams...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Search Query failed: %v\n", err)
		return values, err
	}
	defer rows.Close()

	for rows.Next() {
		var e SearchEntry

		var rank float32
		var langCodes []string
		err = rows.Scan(&e.ID, &e.Title, &e.Authors, &e.StartDate, &e.EndDate, &e.DocumentType, &e.Language, &langCodes, &rank)
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
