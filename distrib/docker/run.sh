#!/bin/sh

exec env DATABASE_MIGRATIONS_PATH=/ypsdb-bin/migrations /ypsdb-bin/yps-db-backend
