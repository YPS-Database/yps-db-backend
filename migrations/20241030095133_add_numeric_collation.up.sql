CREATE COLLATION numeric (provider = icu, locale = 'en@colNumeric=yes');

ALTER TABLE "entries" ALTER COLUMN "id" type TEXT COLLATE numeric;
