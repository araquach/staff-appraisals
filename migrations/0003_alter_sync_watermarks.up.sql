-- 0003_alter_sync_watermarks_up.sql

-- 1) Add new columns
ALTER TABLE sync_watermarks
    ADD COLUMN id bigserial,
    ADD COLUMN entity text,
    ADD COLUMN last_updated_phorest timestamptz,
    ADD COLUMN created_at timestamptz DEFAULT now(),
    ADD COLUMN updated_at timestamptz DEFAULT now();

-- 2) Copy existing data to new columns
UPDATE sync_watermarks
SET entity               = source_name,
    last_updated_phorest = last_updated_at;

-- 3) Drop the old primary key on (source_name, branch_id)
ALTER TABLE sync_watermarks
    DROP CONSTRAINT IF EXISTS sync_watermarks_pkey;

-- 4) Allow branch_id to be nullable
ALTER TABLE sync_watermarks
    ALTER COLUMN branch_id DROP NOT NULL;

-- 5) Normalise empty string branch_ids to NULL
UPDATE sync_watermarks
SET branch_id = 'ALL'
WHERE branch_id IS NULL OR branch_id = '';

-- 6) Make id the new primary key
ALTER TABLE sync_watermarks
    ALTER COLUMN id SET NOT NULL;

ALTER TABLE sync_watermarks
    ADD CONSTRAINT sync_watermarks_pkey PRIMARY KEY (id);

-- 7) Drop old columns we no longer need
ALTER TABLE sync_watermarks
    DROP COLUMN source_name,
    DROP COLUMN last_updated_at,
    DROP COLUMN last_run_at;

-- 8) Drop old index and add the new unique index on (entity, branch_id)
DROP INDEX IF EXISTS idx_sync_watermarks_last_updated;

CREATE UNIQUE INDEX IF NOT EXISTS idx_watermark_entity_branch
    ON sync_watermarks (entity, branch_id);