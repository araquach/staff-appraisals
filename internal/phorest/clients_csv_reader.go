package phorest

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"staff-appraisals/internal/models"
)

type ParsedClients struct {
	Clients []models.Client
}

// ParseClientsCSV reads a Phorest clients CSV and returns a unique set by client_id.
// If multiple rows per client_id exist, we keep the one with the newest UpdatedAtPhorest.
func ParseClientsCSV(path string, lg *log.Logger) (*ParsedClients, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.ReuseRecord = true

	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	idx := make(map[string]int, len(header))
	for i, h := range header {
		idx[strings.TrimSpace(strings.ToLower(h))] = i
	}

	get := func(rec []string, name string) string {
		i, ok := idx[name]
		if !ok || i >= len(rec) {
			return ""
		}
		return rec[i]
	}
	parseInt := func(s string) int {
		s = strings.TrimSpace(s)
		if s == "" {
			return 0
		}
		v, _ := strconv.Atoi(s)
		return v
	}
	parseFloatPtr := func(s string) *float64 {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil
		}
		return &v
	}
	parseIntPtr := func(s string) *int {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		v, err := strconv.Atoi(s)
		if err != nil {
			return nil
		}
		return &v
	}
	parseBool := func(s string) bool {
		s = strings.TrimSpace(strings.ToLower(s))
		return s == "1" || s == "true" || s == "t" || s == "yes" || s == "y"
	}
	parseDate := func(s string) *time.Time {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		layouts := []string{"2006-01-02"}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, s); err == nil {
				return &t
			}
		}
		return nil
	}
	parseTS := func(s string) *time.Time {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		for _, layout := range []string{
			time.RFC3339Nano,
			time.RFC3339,
			"2006-01-02 15:04:05",
			"2006-01-02T15:04:05.000",
			"2006-01-02T15:04:05",
		} {
			if t, err := time.Parse(layout, s); err == nil {
				return &t
			}
		}
		return nil
	}

	// Deduplicate by newest UpdatedAtPhorest (a.k.a. CSV "updated_at")
	byID := map[string]models.Client{}
	row := 0

	for {
		rec, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read row %d: %w", row, err)
		}
		row++

		clientID := get(rec, "client_id")
		if clientID == "" {
			continue
		}

		updated := get(rec, "updated_at") // CSV header
		created := get(rec, "created_at")

		c := models.Client{
			ClientID:           clientID,
			Version:            parseInt(get(rec, "version")),
			FirstName:          get(rec, "first_name"),
			LastName:           get(rec, "last_name"),
			Mobile:             get(rec, "mobile"),
			LinkedClientMobile: get(rec, "linked_client_mobile"),
			LandLine:           get(rec, "land_line"),
			Email:              get(rec, "email"),
			CreatedAtPhorest:   parseTS(created),
			UpdatedAtPhorest:   parseTS(updated),
			BirthDate:          parseDate(get(rec, "birth_date")),
			Gender:             get(rec, "gender"),

			SMSMarketingConsent:   parseBool(get(rec, "sms_marketing_consent")),
			EmailMarketingConsent: parseBool(get(rec, "email_marketing_consent")),
			SMSReminderConsent:    parseBool(get(rec, "sms_reminder_consent")),
			EmailReminderConsent:  parseBool(get(rec, "email_reminder_consent")),
			Archived:              parseBool(get(rec, "archived")),
			Deleted:               parseBool(get(rec, "deleted")),
			Banned:                parseBool(get(rec, "banned")),

			MergedToClientID: get(rec, "merged_to_client_id"),

			StreetAddress1: get(rec, "street_address_1"),
			StreetAddress2: get(rec, "street_address_2"),
			City:           get(rec, "city"),
			State:          get(rec, "state"),
			PostalCode:     get(rec, "postal_code"),
			Country:        get(rec, "country"),

			ClientSince: parseDate(get(rec, "client_since")),
			FirstVisit:  parseDate(get(rec, "first_visit")),
			LastVisit:   parseDate(get(rec, "last_visit")),

			Notes:                    get(rec, "notes"),
			PhotoURL:                 get(rec, "photo_url"),
			PreferredStaffID:         get(rec, "preferred_staff_id"),
			CreditAccountCreditDays:  parseIntPtr(get(rec, "credit_account_credit_days")),
			CreditAccountCreditLimit: parseFloatPtr(get(rec, "credit_account_credit_limit")),
			LoyaltyCardSerialNumber:  get(rec, "loyalty_card_serial_number"),
			ExternalID:               get(rec, "external_id"),
			CreatingBranchID:         get(rec, "creating_branch_id"),
			ClientCategoryIDs:        get(rec, "client_category_ids"),
		}

		if existing, ok := byID[clientID]; ok {
			switch {
			case existing.UpdatedAtPhorest == nil && c.UpdatedAtPhorest != nil:
				byID[clientID] = c
			case existing.UpdatedAtPhorest != nil && c.UpdatedAtPhorest != nil &&
				c.UpdatedAtPhorest.After(*existing.UpdatedAtPhorest):
				byID[clientID] = c
			default:
				// keep existing
			}
		} else {
			byID[clientID] = c
		}
	}

	out := make([]models.Client, 0, len(byID))
	for _, v := range byID {
		out = append(out, v)
	}

	lg.Printf("Parsed clients CSV: %d unique clients", len(out))
	return &ParsedClients{Clients: out}, nil
}
