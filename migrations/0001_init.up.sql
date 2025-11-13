create table branches
(
    id               bigserial
        primary key,
    branch_id        text not null,
    name             text,
    time_zone        text,
    latitude         numeric,
    longitude        numeric,
    street_address_1 text,
    street_address_2 text,
    city             text,
    state            text,
    postal_code      text,
    country          text,
    currency_code    text,
    account_id       bigint,
    created_at       timestamp with time zone,
    updated_at       timestamp with time zone
);

alter table branches
    owner to postgres;

create unique index idx_branches_branch_id
    on branches (branch_id);

create table staff
(
    id                           bigserial
        primary key,
    staff_id                     text not null,
    branch_id                    text not null,
    staff_category_id            text,
    user_id                      text,
    staff_category_name          text,
    first_name                   text,
    last_name                    text,
    birth_date                   date,
    start_date                   date,
    self_employed                boolean,
    archived                     boolean,
    mobile                       text,
    email                        text,
    gender                       text,
    notes                        text,
    online_profile               text,
    hide_from_online_bookings    boolean,
    hide_from_appointment_screen boolean,
    image_url                    text,
    created_at                   timestamp with time zone,
    updated_at                   timestamp with time zone
);

alter table staff
    owner to postgres;

create unique index ux_staff_branch
    on staff (staff_id, branch_id);

create table clients
(
    id                          bigserial
        primary key,
    client_id                   text,
    version                     bigint,
    first_name                  text,
    last_name                   text,
    mobile                      text,
    linked_client_mobile        text,
    land_line                   text,
    email                       text,
    created_at_phorest          timestamp with time zone,
    updated_at_phorest          timestamp with time zone,
    birth_date                  date,
    gender                      text,
    sms_marketing_consent       boolean,
    email_marketing_consent     boolean,
    sms_reminder_consent        boolean,
    email_reminder_consent      boolean,
    archived                    boolean,
    deleted                     boolean,
    banned                      boolean,
    merged_to_client_id         text,
    street_address_1            text,
    street_address_2            text,
    city                        text,
    state                       text,
    postal_code                 text,
    country                     text,
    client_since                date,
    first_visit                 date,
    last_visit                  date,
    notes                       text,
    photo_url                   text,
    preferred_staff_id          text,
    credit_account_credit_days  bigint,
    credit_account_credit_limit numeric,
    loyalty_card_serial_number  text,
    external_id                 text,
    creating_branch_id          text,
    client_category_ids         text,
    created_at                  timestamp with time zone,
    updated_at                  timestamp with time zone
);

alter table clients
    owner to postgres;

create index idx_clients_preferred_staff_id
    on clients (preferred_staff_id);

create index idx_clients_updated_at_phorest
    on clients (updated_at_phorest);

create unique index idx_clients_client_id
    on clients (client_id);

create table transactions
(
    id                 bigserial
        primary key,
    transaction_id     text,
    branch_id          text,
    branch_name        text,
    client_id          text,
    client_first_name  text,
    client_last_name   text,
    client_source      text,
    purchased_date     date,
    purchase_time      time,
    updated_at_phorest timestamp with time zone,
    created_at         timestamp with time zone,
    updated_at         timestamp with time zone
);

alter table transactions
    owner to postgres;

create index idx_transactions_updated_at_phorest
    on transactions (updated_at_phorest);

create index idx_transactions_client_id
    on transactions (client_id);

create index idx_transactions_branch_id
    on transactions (branch_id);

create unique index idx_transactions_transaction_id
    on transactions (transaction_id);

create table transaction_items
(
    id                                   bigserial
        primary key,
    transaction_item_id                  text,
    transaction_id                       text,
    branch_id                            text,
    branch_name                          text,
    client_id                            text,
    client_first_name                    text,
    client_last_name                     text,
    client_source                        text,
    purchased_date                       date,
    purchase_time                        time,
    item_type                            text,
    description                          text,
    quantity                             numeric,
    purchase_voucher_discount_percentage numeric,
    purchase_online_deposit              numeric,
    purchase_online_discount_amount      numeric,
    service_id                           text,
    service_name                         text,
    service_category_id                  text,
    service_category_name                text,
    package_id                           text,
    package_name                         text,
    special_offer_id                     text,
    special_offer_name                   text,
    product_id                           text,
    product_name                         text,
    product_brand_id                     text,
    product_brand_name                   text,
    product_category_id                  text,
    product_category_name                text,
    product_barcode                      text,
    product_code                         text,
    course_id                            text,
    course_name                          text,
    client_course_name                   text,
    voucher_serial                       text,
    service_reward_id                    text,
    service_reward_name                  text,
    product_reward_id                    text,
    product_reward_name                  text,
    unit_price                           numeric,
    original_price                       numeric,
    discount_type                        numeric,
    discount_value                       numeric,
    item_online_deposit                  numeric,
    item_online_discount                 numeric,
    loyalty_points_awarded               numeric,
    tax_rate                             numeric,
    total_amount                         numeric,
    total_amount_pre_vouch_disc          numeric,
    net_total_amount                     numeric,
    gross_total_amount                   numeric,
    net_price                            numeric,
    gross_price                          numeric,
    discount_amount                      numeric,
    tax_amount                           numeric,
    staff_tips                           numeric,
    product_cost_price                   numeric,
    service_cost                         numeric,
    service_cost_type                    text,
    gross_total_with_discount            numeric,
    gross_total_with_discount_minus_tax  numeric,
    simple_discount_amount               numeric,
    membership_benefit_used              bigint,
    membership_discount_amount           numeric,
    deal                                 numeric,
    session_net_amount                   numeric,
    session_gross_amount                 numeric,
    phorest_tips                         numeric,
    payment_type                         text,
    payment_type_ids                     text,
    payment_type_amounts                 numeric,
    payment_type_codes                   text,
    payment_type_names                   text,
    payment_type_voucher_serials         text,
    payment_type_prepaid_tax_amounts     text,
    outstanding_balance_pmt              bigint,
    open_sale                            boolean,
    open_sale_type                       text,
    purchase_type                        text,
    online_booking                       bigint,
    void                                 bigint,
    voided_transaction_id                text,
    void_reason                          text,
    department_id                        text,
    department_name                      text,
    staff_id                             text,
    staff_first_name                     text,
    staff_last_name                      text,
    staff_category_id                    text,
    staff_category_name                  text,
    is_requested_staff                   bigint,
    primary_staff_id                     text,
    preferred_staff_id                   text,
    preferred_staff_name                 text,
    appointment_id                       text,
    appointment_date                     date,
    appointment_created                  timestamp with time zone,
    appointment_rating                   bigint,
    client_birthday                      date,
    client_gender                        text,
    client_email                         text,
    client_first_visit                   date,
    appt_client_id                       text,
    appt_client_first_name               text,
    appt_client_last_name                text,
    appt_client_birthday                 date,
    appt_client_gender                   text,
    appt_client_email                    text,
    appt_client_first_visit              date,
    internet_category_ids                text,
    internet_category_names              text,
    branch_product_id                    text,
    fixed_discount_id                    text,
    fixed_discount_name                  text,
    client_course_id                     text,
    creating_user                        text,
    tax_rate_name                        text,
    sale_fee_id                          text,
    updated_at_phorest                   timestamp with time zone,
    created_at                           timestamp with time zone,
    updated_at                           timestamp with time zone
);

ALTER TABLE transaction_items
    DROP CONSTRAINT IF EXISTS fk_transactions_items;

ALTER TABLE transaction_items
    ADD CONSTRAINT fk_transactions_items
        FOREIGN KEY (transaction_id)
            REFERENCES transactions (transaction_id)
            ON UPDATE CASCADE
            ON DELETE CASCADE;

alter table transaction_items
    owner to postgres;

create index idx_transaction_items_updated_at_phorest
    on transaction_items (updated_at_phorest);

create index idx_transaction_items_appointment_id
    on transaction_items (appointment_id);

create index idx_transaction_items_staff_id
    on transaction_items (staff_id);

create index idx_transaction_items_item_type
    on transaction_items (item_type);

create index idx_transaction_items_client_id
    on transaction_items (client_id);

create index idx_transaction_items_branch_id
    on transaction_items (branch_id);

create index idx_transaction_items_transaction_id
    on transaction_items (transaction_id);

create unique index idx_transaction_items_transaction_item_id
    on transaction_items (transaction_item_id);

create table reviews
(
    id                bigserial
        primary key,
    review_id         text not null,
    branch_id         text not null,
    client_id         text,
    client_first_name text,
    client_last_name  text,
    review_date       date,
    visit_date        date,
    staff_id          text,
    staff_first_name  text,
    staff_last_name   text,
    text              text,
    rating            bigint,
    facebook_review   boolean,
    twitter_review    boolean,
    created_at        timestamp with time zone,
    updated_at        timestamp with time zone
);

alter table reviews
    owner to postgres;

create index idx_reviews_rating
    on reviews (rating);

create index idx_reviews_staff_id
    on reviews (staff_id);

create index idx_reviews_visit_date
    on reviews (visit_date);

create index idx_reviews_review_date
    on reviews (review_date);

create index idx_reviews_client_id
    on reviews (client_id);

create index idx_reviews_branch_id
    on reviews (branch_id);

create unique index idx_reviews_review_id
    on reviews (review_id);

