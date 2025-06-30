DROP INDEX IF EXISTS events_aggregate_id_idx;
DROP INDEX IF EXISTS events_transaction_id_idx;
ALTER TABLE events DROP CONSTRAINT IF EXISTS events_unique_columns;
DROP TABLE IF EXISTS events;
