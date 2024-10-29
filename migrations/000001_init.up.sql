CREATE TABLE "entries" (
  "id" text PRIMARY KEY,
  "url" text,
  "entry_type" text,
  "entry_language" text,
  "regions" text[],
  "start_date" date,
  "end_date" date,
  "alternates" text[],
  "related" text[],
  "title" text,
  "authors" text,
  "abstract" text,
  "keywords" text[],
  "orgs" text[],
  "org_doc_id" text,
  "org_type" text,
  "youth_led" text,
  "notes" text
);

CREATE TABLE "temp_insert_entries" (
  "id" text PRIMARY KEY,
  "url" text,
  "entry_type" text,
  "entry_language" text,
  "regions" text[],
  "start_date" date,
  "end_date" date,
  "alternates" text[],
  "related" text[],
  "title" text,
  "authors" text,
  "abstract" text,
  "keywords" text[],
  "orgs" text[],
  "org_doc_id" text,
  "org_type" text,
  "youth_led" text,
  "notes" text
);

CREATE TABLE "entry_files" (
  "entry_id" text,
  "filename" text,
  "url" text,
  PRIMARY KEY ("entry_id", "filename")
);

CREATE TABLE "spreadsheet_files" (
  "id" text PRIMARY KEY,
  "filename" text,
  "url" text,
  "added_at" timestamp
);

CREATE TABLE "pages" (
  "id" text PRIMARY KEY,
  "content" text NOT NULL DEFAULT 'This is a sample page. Login to edit it.',
  "google_form_id" text NOT NULL DEFAULT '',
  "updated_at" timestamp NOT NULL DEFAULT (now())
);

ALTER TABLE "entry_files" ADD FOREIGN KEY ("entry_id") REFERENCES "entries" ("id");

-- basic page contents
INSERT INTO pages (id) VALUES ('home'), ('about'), ('publications'), ('data'), ('submit');

-- full text search
CREATE OR REPLACE FUNCTION f_kwarr2text(text[]) 
  RETURNS text LANGUAGE sql IMMUTABLE AS $$SELECT array_to_string($1, '; ', '*')$$;

ALTER TABLE entries
  ADD COLUMN titlesearch_index_col tsvector
    GENERATED ALWAYS AS (to_tsvector('english', title)) STORED;
CREATE INDEX titlesearch_idx ON entries USING GIN (titlesearch_index_col);

ALTER TABLE entries
  ADD COLUMN abstractsearch_index_col tsvector
    GENERATED ALWAYS AS (to_tsvector('english', abstract)) STORED;
CREATE INDEX abstractsearch_idx ON entries USING GIN (abstractsearch_index_col);

ALTER TABLE entries
  ADD COLUMN alltextsearch_index_col tsvector
    GENERATED ALWAYS AS (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(f_kwarr2text(keywords)) || ' ' || coalesce(abstract, ''))) STORED;
CREATE INDEX alltextsearch_idx ON entries USING GIN (alltextsearch_index_col);
