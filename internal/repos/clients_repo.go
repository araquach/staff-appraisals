package repos

import (
	"fmt"
	"log"
	"strings"
	"time"

	"gorm.io/gorm"
	"staff-appraisals/internal/models"
)

type ClientsRepo struct {
	db *gorm.DB
	lg *log.Logger
}

func NewClientsRepo(db *gorm.DB, lg *log.Logger) *ClientsRepo {
	return &ClientsRepo{db: db, lg: lg}
}

// UpsertBatch inserts/updates clients based on client_id, only when EXCLUDED.updated_at_phorest is newer.
func (r *ClientsRepo) UpsertBatch(rows []models.Client, batchSize int) error {
	if len(rows) == 0 {
		return nil
	}
	now := time.Now().UTC()

	cols := []string{
		"client_id", "version", "first_name", "last_name",
		"mobile", "linked_client_mobile", "land_line", "email",
		"created_at_phorest", // was created_at in INSERT — fixed below
		"updated_at_phorest",
		"birth_date", "gender",
		"sms_marketing_consent", "email_marketing_consent", "sms_reminder_consent", "email_reminder_consent",
		"archived", "deleted", "banned",
		"merged_to_client_id",
		"street_address_1", "street_address_2", "city", "state", "postal_code", "country",
		"client_since", "first_visit", "last_visit",
		"notes", "photo_url", "preferred_staff_id",
		"credit_account_credit_days", "credit_account_credit_limit",
		"loyalty_card_serial_number", "external_id", "creating_branch_id", "client_category_ids",
		"created_at", "updated_at", // renamed from created_at_sys/updated_at_sys for clarity
	}

	placeholders := make([]string, 0, len(rows))
	args := make([]any, 0, len(rows)*len(cols))

	flush := func() error {
		if len(placeholders) == 0 {
			return nil
		}
		sql := fmt.Sprintf(`
INSERT INTO clients (
  client_id, version, first_name, last_name,
  mobile, linked_client_mobile, land_line, email,
  created_at_phorest, updated_at_phorest,          -- ✅ fixed here
  birth_date, gender,
  sms_marketing_consent, email_marketing_consent, sms_reminder_consent, email_reminder_consent,
  archived, deleted, banned,
  merged_to_client_id,
  street_address_1, street_address_2, city, state, postal_code, country,
  client_since, first_visit, last_visit,
  notes, photo_url, preferred_staff_id,
  credit_account_credit_days, credit_account_credit_limit,
  loyalty_card_serial_number, external_id, creating_branch_id, client_category_ids,
  created_at, updated_at
) VALUES %s
ON CONFLICT (client_id) DO UPDATE SET
  version = EXCLUDED.version,
  first_name = EXCLUDED.first_name,
  last_name = EXCLUDED.last_name,
  mobile = EXCLUDED.mobile,
  linked_client_mobile = EXCLUDED.linked_client_mobile,
  land_line = EXCLUDED.land_line,
  email = EXCLUDED.email,
  created_at_phorest = EXCLUDED.created_at_phorest,
  updated_at_phorest = EXCLUDED.updated_at_phorest,
  birth_date = EXCLUDED.birth_date,
  gender = EXCLUDED.gender,
  sms_marketing_consent = EXCLUDED.sms_marketing_consent,
  email_marketing_consent = EXCLUDED.email_marketing_consent,
  sms_reminder_consent = EXCLUDED.sms_reminder_consent,
  email_reminder_consent = EXCLUDED.email_reminder_consent,
  archived = EXCLUDED.archived,
  deleted = EXCLUDED.deleted,
  banned = EXCLUDED.banned,
  merged_to_client_id = EXCLUDED.merged_to_client_id,
  street_address_1 = EXCLUDED.street_address_1,
  street_address_2 = EXCLUDED.street_address_2,
  city = EXCLUDED.city,
  state = EXCLUDED.state,
  postal_code = EXCLUDED.postal_code,
  country = EXCLUDED.country,
  client_since = EXCLUDED.client_since,
  first_visit = EXCLUDED.first_visit,
  last_visit = EXCLUDED.last_visit,
  notes = EXCLUDED.notes,
  photo_url = EXCLUDED.photo_url,
  preferred_staff_id = EXCLUDED.preferred_staff_id,
  credit_account_credit_days = EXCLUDED.credit_account_credit_days,
  credit_account_credit_limit = EXCLUDED.credit_account_credit_limit,
  loyalty_card_serial_number = EXCLUDED.loyalty_card_serial_number,
  external_id = EXCLUDED.external_id,
  creating_branch_id = EXCLUDED.creating_branch_id,
  client_category_ids = EXCLUDED.client_category_ids,
  updated_at = EXCLUDED.updated_at
WHERE clients.updated_at_phorest IS NULL
   OR EXCLUDED.updated_at_phorest > clients.updated_at_phorest;`,
			strings.Join(placeholders, ","),
		)
		if err := r.db.Exec(sql, args...).Error; err != nil {
			return err
		}
		r.lg.Printf("Upserted clients: %d", len(placeholders))
		placeholders = placeholders[:0]
		args = args[:0]
		return nil
	}

	for _, c := range rows {
		placeholders = append(placeholders, "("+strings.Repeat("?,", len(cols)-1)+"?)")
		args = append(args,
			c.ClientID, c.Version, c.FirstName, c.LastName,
			c.Mobile, c.LinkedClientMobile, c.LandLine, c.Email,
			c.CreatedAtPhorest, c.UpdatedAtPhorest,
			c.BirthDate, c.Gender,
			c.SMSMarketingConsent, c.EmailMarketingConsent, c.SMSReminderConsent, c.EmailReminderConsent,
			c.Archived, c.Deleted, c.Banned,
			c.MergedToClientID,
			c.StreetAddress1, c.StreetAddress2, c.City, c.State, c.PostalCode, c.Country,
			c.ClientSince, c.FirstVisit, c.LastVisit,
			c.Notes, c.PhotoURL, c.PreferredStaffID,
			c.CreditAccountCreditDays, c.CreditAccountCreditLimit,
			c.LoyaltyCardSerialNumber, c.ExternalID, c.CreatingBranchID, c.ClientCategoryIDs,
			now, now,
		)
		if len(placeholders) >= batchSize {
			if err := flush(); err != nil {
				return err
			}
		}
	}
	return flush()
}
