#!/bin/bash
set -e

# Initialize postgres user and test database for integration tests
psql -v ON_ERROR_STOP=1 --username "$POSTGRES_USER" --dbname "$POSTGRES_DB" <<-EOSQL
	CREATE USER docker;
	CREATE DATABASE transaction_test;
	GRANT pg_read_all_data, pg_write_all_data TO docker;
EOSQL
