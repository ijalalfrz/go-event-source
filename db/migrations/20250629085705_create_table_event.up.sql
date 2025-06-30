CREATE TABLE IF NOT EXISTS events (
    id bigint PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    transaction_id varchar(100) NOT NULL,
    aggregate_id bigint NOT NULL,
    aggregate_type varchar(100) NOT NULL,
    event_type varchar(100) NOT NULL,
    sequence_number int NOT NULL,
    event_data jsonb NOT NULL,
    version varchar(20) NOT NULL,
    created_at timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE events ADD CONSTRAINT events_unique_columns UNIQUE (aggregate_id, aggregate_type, sequence_number);
CREATE INDEX events_transaction_id_idx ON events (transaction_id);
CREATE INDEX events_aggregate_id_idx ON events (aggregate_id);
