CREATE TABLE "entries" (
  "id" string PRIMARY KEY,
  "url" string,
  "entry_type" string,
  "entry_language" string,
  "languages" string[],
  "regions" string[],
  "start_date" date,
  "end_date" date,
  "alternates" string[],
  "related" string[],
  "title" string,
  "authors" string[],
  "abstract" string,
  "keywords" string[],
  "org" string,
  "org_doc_id" string,
  "org_type" string,
  "youth_led" string,
  "notes" string
);

CREATE TABLE "entry_files" (
  "entry_id" string,
  "filename" string,
  "url" string,
  PRIMARY KEY ("entry_id", "filename")
);

CREATE TABLE "spreadsheet_files" (
  "id" string PRIMARY KEY,
  "filename" string,
  "url" string,
  "added_at" timestamp
);

CREATE TABLE "pages" (
  "id" integer PRIMARY KEY,
  "slug" string,
  "content" string DEFAULT 'This is a sample page. Login to edit it.',
  "updated_at" timestamp
);

ALTER TABLE "entry_files" ADD FOREIGN KEY ("entry_id") REFERENCES "entries" ("id");
