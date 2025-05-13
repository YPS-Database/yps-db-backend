CREATE TABLE logs (
  id SERIAL PRIMARY KEY,
  ts TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  log_level VARCHAR(20) NOT NULL CHECK(log_level IN ('DEBUG', 'INFO', 'WARNING', 'ERROR')),
  
  -- e.g. login, import_database, etc
  event_type VARCHAR(50) NOT NULL,

  message TEXT NOT NULL,
  extra_data JSONB NOT NULL DEFAULT '{}'::JSONB
);
