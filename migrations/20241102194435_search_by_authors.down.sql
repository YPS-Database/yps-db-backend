DROP INDEX alltextsearch_idx;
ALTER TABLE entries
  DROP COLUMN IF EXISTS alltextsearch_index_col;

ALTER TABLE entries
  ADD COLUMN alltextsearch_index_col tsvector
    GENERATED ALWAYS AS (to_tsvector('english', coalesce(title, '') || ' ' || coalesce(f_kwarr2text(keywords)) || ' ' || coalesce(abstract, ''))) STORED;
CREATE INDEX alltextsearch_idx ON entries USING GIN (alltextsearch_index_col);
