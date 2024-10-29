package yps

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	values["year"] = years

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
	values["entry_type"] = entryTypes

	return values, err
}

func (db *YPSDatabase) GetAllEntries() (entries map[string]Entry, err error) {
	entries = make(map[string]Entry)

	rows, err := db.pool.Query(context.Background(), `
select id, url, entry_type, entry_language, start_date, end_date, alternates,
	related, title, authors, abstract, keywords, org, org_doc_id, org_type, youth_led
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
			&e.OrgPublisher, &e.OrgDocID, &e.OrgType, &e.YouthLed)
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
			entry.Keywords, entry.OrgPublisher, entry.OrgDocID, entry.OrgType, entry.YouthLed,
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
		"related", "title", "authors", "abstract", "keywords", "org", "org_doc_id", "org_type",
		"youth_led"}, source)
	fmt.Println("ended copy into temp table")

	if err != nil {
		return err
	}

	// transfer rows to real table
	_, err = db.pool.Exec(context.Background(), `
insert into entries (id, url, entry_type, entry_language, start_date, end_date, alternates, related, title, authors, abstract, keywords, org, org_doc_id, org_type, youth_led)
select id, url, entry_type, entry_language, start_date, end_date, alternates, related, title, authors, abstract, keywords, org, org_doc_id, org_type, youth_led
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
	org=excluded.org,
	org_doc_id=excluded.org_doc_id,
	org_type=excluded.org_type,
	youth_led=excluded.youth_led
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

	return err
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
