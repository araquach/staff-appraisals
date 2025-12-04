-- 1) Product master table (business-wide)
CREATE TABLE ph_products (
                             id               TEXT PRIMARY KEY,         -- productId from Phorest
                             parent_id        TEXT,
                             name             TEXT NOT NULL,
                             brand_id         TEXT,
                             brand_name       TEXT,
                             category_id      TEXT,
                             category_name    TEXT,
                             code             TEXT,
                             type_raw         TEXT,                     -- e.g. "RETAIL, COLOUR, PROFESSIONAL"
                             measurement_qty  NUMERIC,
                             measurement_unit TEXT,
                             archived         BOOLEAN NOT NULL DEFAULT FALSE,
                             created_at_ph    TIMESTAMPTZ,
                             updated_at_ph    TIMESTAMPTZ,
                             inserted_at      TIMESTAMPTZ NOT NULL DEFAULT now(),
                             updated_at       TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_ph_products_brand ON ph_products(brand_name);
CREATE INDEX idx_ph_products_category ON ph_products(category_name);


-- 2) Current stock per branch
CREATE TABLE ph_product_stock (
                                  id                BIGSERIAL PRIMARY KEY,
                                  product_id        TEXT NOT NULL REFERENCES ph_products(id) ON DELETE CASCADE,
                                  branch_id         TEXT NOT NULL,               -- Phorest branchId
                                  price             NUMERIC,                     -- selling price
                                  min_quantity      NUMERIC,
                                  max_quantity      NUMERIC,
                                  quantity_in_stock NUMERIC,
                                  reorder_count     NUMERIC,
                                  reorder_cost      NUMERIC,                     -- cost price per unit
                                  archived          BOOLEAN NOT NULL DEFAULT FALSE,
                                  created_at_ph     TIMESTAMPTZ,
                                  updated_at_ph     TIMESTAMPTZ,
                                  last_synced_at    TIMESTAMPTZ NOT NULL DEFAULT now(),

                                  UNIQUE (product_id, branch_id)
);

CREATE INDEX idx_ph_product_stock_branch ON ph_product_stock(branch_id);
CREATE INDEX idx_ph_product_stock_product ON ph_product_stock(product_id);
CREATE INDEX idx_ph_product_stock_updated_ph ON ph_product_stock(updated_at_ph);


-- 3) History of stock levels over time (append-only)
CREATE TABLE ph_product_stock_history (
                                          id                BIGSERIAL PRIMARY KEY,
                                          product_id        TEXT NOT NULL REFERENCES ph_products(id) ON DELETE CASCADE,
                                          branch_id         TEXT NOT NULL,
                                          snapshot_time     TIMESTAMPTZ NOT NULL DEFAULT now(),
                                          quantity_in_stock NUMERIC,
                                          price             NUMERIC,
                                          source            TEXT NOT NULL DEFAULT 'sync'  -- 'sync', 'manual', etc.
);

CREATE INDEX idx_ph_stock_hist_branch_product_time
    ON ph_product_stock_history (branch_id, product_id, snapshot_time);


-- 4) Daily DV view (last known quantity per day)
CREATE VIEW dv_daily_stock_levels AS
SELECT
    branch_id,
    product_id,
    date_trunc('day', snapshot_time)::date AS stock_date,
    (ARRAY_AGG(quantity_in_stock ORDER BY snapshot_time DESC))[1] AS quantity_in_stock,
    (ARRAY_AGG(price ORDER BY snapshot_time DESC))[1]             AS price
FROM ph_product_stock_history
GROUP BY branch_id, product_id, date_trunc('day', snapshot_time);