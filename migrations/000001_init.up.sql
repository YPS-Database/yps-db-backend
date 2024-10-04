CREATE TABLE "entries" (
  "id" text PRIMARY KEY,
  "url" text,
  "entry_type" text,
  "entry_language" text,
  "languages" text[],
  "regions" text[],
  "start_date" date,
  "end_date" date,
  "alternates" text[],
  "related" text[],
  "title" text,
  "authors" text[],
  "abstract" text,
  "keywords" text[],
  "org" text,
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
INSERT INTO pages (id) VALUES ('about'), ('publications'), ('data'), ('submit');