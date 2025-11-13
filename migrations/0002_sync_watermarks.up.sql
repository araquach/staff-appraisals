CREATE TABLE IF NOT EXISTS sync_watermarks (
                                               source_name      text    NOT NULL,          -- e.g. 'clients_csv', 'transactions_csv', 'reviews_api'
                                               branch_id        text    NOT NULL,          -- '' (empty) for global sources like clients
                                               last_updated_at  timestamptz,
                                               last_run_at      timestamptz NOT NULL DEFAULT now(),
                                               PRIMARY KEY (source_name, branch_id)
);

CREATE INDEX IF NOT EXISTS idx_sync_watermarks_last_updated
    ON sync_watermarks (source_name, branch_id, last_updated_at);