package models

import "time"

// Client mirrors Phorest's client export. We keep a watermark column
// updated_at_phorest so all sync code uses the same convention.
type Client struct {
	// Local PK (surrogate)
	ID uint `gorm:"primaryKey"`

	// Natural key from Phorest (unique)
	ClientID string `gorm:"column:client_id;uniqueIndex"`

	// Core identity
	Version            int    `gorm:"column:version"`
	FirstName          string `gorm:"column:first_name"`
	LastName           string `gorm:"column:last_name"`
	Mobile             string `gorm:"column:mobile"`
	LinkedClientMobile string `gorm:"column:linked_client_mobile"`
	LandLine           string `gorm:"column:land_line"`
	Email              string `gorm:"column:email"`

	// Timestamps from Phorest export
	CreatedAtPhorest *time.Time `gorm:"column:created_at_phorest;type:timestamptz"`
	UpdatedAtPhorest *time.Time `gorm:"column:updated_at_phorest;type:timestamptz;index"`

	// Demographics
	BirthDate *time.Time `gorm:"column:birth_date;type:date"`
	Gender    string     `gorm:"column:gender"`

	// Consents / flags
	SMSMarketingConsent   bool `gorm:"column:sms_marketing_consent"`
	EmailMarketingConsent bool `gorm:"column:email_marketing_consent"`
	SMSReminderConsent    bool `gorm:"column:sms_reminder_consent"`
	EmailReminderConsent  bool `gorm:"column:email_reminder_consent"`
	Archived              bool `gorm:"column:archived"`
	Deleted               bool `gorm:"column:deleted"`
	Banned                bool `gorm:"column:banned"`

	// Links
	MergedToClientID string `gorm:"column:merged_to_client_id"`

	// Address
	StreetAddress1 string `gorm:"column:street_address_1"`
	StreetAddress2 string `gorm:"column:street_address_2"`
	City           string `gorm:"column:city"`
	State          string `gorm:"column:state"`
	PostalCode     string `gorm:"column:postal_code"`
	Country        string `gorm:"column:country"`

	// Dates
	ClientSince *time.Time `gorm:"column:client_since;type:date"`
	FirstVisit  *time.Time `gorm:"column:first_visit;type:date"`
	LastVisit   *time.Time `gorm:"column:last_visit;type:date"`

	// Misc
	Notes                    string   `gorm:"column:notes"`
	PhotoURL                 string   `gorm:"column:photo_url"`
	PreferredStaffID         string   `gorm:"column:preferred_staff_id;index"`
	CreditAccountCreditDays  *int     `gorm:"column:credit_account_credit_days"`
	CreditAccountCreditLimit *float64 `gorm:"column:credit_account_credit_limit"`
	LoyaltyCardSerialNumber  string   `gorm:"column:loyalty_card_serial_number"`
	ExternalID               string   `gorm:"column:external_id"`
	CreatingBranchID         string   `gorm:"column:creating_branch_id"`
	ClientCategoryIDs        string   `gorm:"column:client_category_ids"`

	// GORM metadata
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (Client) TableName() string { return "clients" }
