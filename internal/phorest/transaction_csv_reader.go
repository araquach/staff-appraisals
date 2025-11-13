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

type ParsedBatch struct {
	Transactions []models.Transaction
	Items        []models.TransactionItem
}

// ParseTransactionsCSV reads a Phorest transactions CSV and returns split header/items.
// It’s header-driven (no hard-coded positions), tolerant to extra/missing columns,
// and builds one Transaction per unique transaction_id (choosing newest by updated_at_phorest).
func ParseTransactionsCSV(path string, lg *log.Logger) (*ParsedBatch, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open csv: %w", err)
	}
	defer f.Close()

	r := csv.NewReader(f)
	r.ReuseRecord = true

	// Header map
	header, err := r.Read()
	if err != nil {
		return nil, fmt.Errorf("read header: %w", err)
	}
	idx := make(map[string]int, len(header))
	for i, h := range header {
		idx[strings.TrimSpace(strings.ToLower(h))] = i
	}

	// Helpers to read fields safely
	get := func(rec []string, name string) string {
		i, ok := idx[name]
		if !ok || i >= len(rec) {
			return ""
		}
		return rec[i]
	}
	parseFloat := func(s string) float64 {
		s = strings.TrimSpace(s)
		if s == "" {
			return 0
		}
		v, _ := strconv.ParseFloat(s, 64)
		return v
	}
	parseInt64 := func(s string) int64 {
		s = strings.TrimSpace(s)
		if s == "" {
			return 0
		}
		v, _ := strconv.ParseInt(s, 10, 64)
		return v
	}
	parseInt := func(s string) int {
		s = strings.TrimSpace(s)
		if s == "" {
			return 0
		}
		v, _ := strconv.Atoi(s)
		return v
	}
	parseBool := func(s string) bool {
		// CSV often uses 0/1 or true/false
		s = strings.TrimSpace(strings.ToLower(s))
		return s == "1" || s == "true" || s == "t" || s == "yes" || s == "y"
	}
	parseDate := func(s string) *time.Time {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		// Common formats in your samples
		layouts := []string{"2006-01-02"}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, s); err == nil {
				return &t
			}
		}
		return nil
	}
	parseClock := func(s string) *time.Time {
		s = strings.TrimSpace(s)
		if s == "" {
			return nil
		}
		for _, layout := range []string{"15:04:05.000", "15:04:05", "15:04"} {
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
		// observed formats (no TZ, or ISO with T)
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

	items := make([]models.TransactionItem, 0, 2048)
	// Use a map to keep one header per transaction_id, selecting the newest updated_at_phorest
	txByID := map[string]models.Transaction{}

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

		transactionID := get(rec, "transaction_id")
		if transactionID == "" {
			// Skip malformed rows
			continue
		}

		// Map line → TransactionItem (1:1)
		item := models.TransactionItem{
			TransactionItemID: get(rec, "transaction_item_id"),
			TransactionID:     transactionID,

			BranchID:        get(rec, "branch_id"),
			BranchName:      get(rec, "branch_name"),
			ClientID:        get(rec, "client_id"),
			ClientFirstName: get(rec, "client_first_name"),
			ClientLastName:  get(rec, "client_last_name"),
			ClientSource:    get(rec, "client_source"),
			PurchasedDate:   parseDate(get(rec, "purchased_date")),
			PurchaseTime:    parseClock(get(rec, "purchase_time")),

			ItemType:    get(rec, "item_type"),
			Description: get(rec, "description"),
			Quantity:    parseFloat(get(rec, "quantity")),

			PurchaseVoucherDiscountPercentage: parseFloat(get(rec, "purchase_voucher_discount_percentage")),
			PurchaseOnlineDeposit:             parseFloat(get(rec, "purchase_online_deposit")),
			PurchaseOnlineDiscountAmount:      parseFloat(get(rec, "purchase_online_discount_amount")),

			ServiceID:           get(rec, "service_id"),
			ServiceName:         get(rec, "service_name"),
			ServiceCategoryID:   get(rec, "service_category_id"),
			ServiceCategoryName: get(rec, "service_category_name"),

			PackageID:        get(rec, "package_id"),
			PackageName:      get(rec, "package_name"),
			SpecialOfferID:   get(rec, "special_offer_id"),
			SpecialOfferName: get(rec, "special_offer_name"),

			ProductID:           get(rec, "product_id"),
			ProductName:         get(rec, "product_name"),
			ProductBrandID:      get(rec, "product_brand_id"),
			ProductBrandName:    get(rec, "product_brand_name"),
			ProductCategoryID:   get(rec, "product_category_id"),
			ProductCategoryName: get(rec, "product_category_name"),
			ProductBarcode:      get(rec, "product_barcode"),
			ProductCode:         get(rec, "product_code"),

			CourseID:          get(rec, "course_id"),
			CourseName:        get(rec, "course_name"),
			ClientCourseName:  get(rec, "client_course_name"),
			VoucherSerial:     get(rec, "voucher_serial"),
			ServiceRewardID:   get(rec, "service_reward_id"),
			ServiceRewardName: get(rec, "service_reward_name"),
			ProductRewardID:   get(rec, "product_reward_id"),
			ProductRewardName: get(rec, "product_reward_name"),

			UnitPrice:                      parseFloat(get(rec, "unit_price")),
			OriginalPrice:                  parseFloat(get(rec, "original_price")),
			DiscountType:                   parseFloat(get(rec, "discount_type")),
			DiscountValue:                  parseFloat(get(rec, "discount_value")),
			ItemOnlineDeposit:              parseFloat(get(rec, "item_online_deposit")),
			ItemOnlineDiscount:             parseFloat(get(rec, "item_online_discount")),
			LoyaltyPointsAwarded:           parseFloat(get(rec, "loyalty_points_awarded")),
			TaxRate:                        parseFloat(get(rec, "tax_rate")),
			TotalAmount:                    parseFloat(get(rec, "total_amount")),
			TotalAmountPreVouchDisc:        parseFloat(get(rec, "total_amount_pre_vouch_disc")),
			NetTotalAmount:                 parseFloat(get(rec, "net_total_amount")),
			GrossTotalAmount:               parseFloat(get(rec, "gross_total_amount")),
			NetPrice:                       parseFloat(get(rec, "net_price")),
			GrossPrice:                     parseFloat(get(rec, "gross_price")),
			DiscountAmount:                 parseFloat(get(rec, "discount_amount")),
			TaxAmount:                      parseFloat(get(rec, "tax_amount")),
			StaffTips:                      parseFloat(get(rec, "staff_tips")),
			ProductCostPrice:               parseFloat(get(rec, "product_cost_price")),
			ServiceCost:                    parseFloat(get(rec, "service_cost")),
			ServiceCostType:                get(rec, "service_cost_type"),
			GrossTotalWithDiscount:         parseFloat(get(rec, "gross_total_with_discount")),
			GrossTotalWithDiscountMinusTax: parseFloat(get(rec, "gross_total_with_discount_minus_tax")),
			SimpleDiscountAmount:           parseFloat(get(rec, "simple_discount_amount")),
			MembershipBenefitUsed:          parseInt(get(rec, "membership_benefit_used")),
			MembershipDiscountAmount:       parseFloat(get(rec, "membership_discount_amount")),
			Deal:                           parseFloat(get(rec, "deal")),
			SessionNetAmount:               parseFloat(get(rec, "session_net_amount")),
			SessionGrossAmount:             parseFloat(get(rec, "session_gross_amount")),
			PhorestTips:                    parseFloat(get(rec, "phorest_tips")),

			PaymentType:                  get(rec, "payment_type"),
			PaymentTypeIDs:               get(rec, "payment_type_ids"),
			PaymentTypeAmounts:           parseFloat(get(rec, "payment_type_amounts")),
			PaymentTypeCodes:             get(rec, "payment_type_codes"),
			PaymentTypeNames:             get(rec, "payment_type_names"),
			PaymentTypeVoucherSerials:    get(rec, "payment_type_voucher_serials"),
			PaymentTypePrepaidTaxAmounts: get(rec, "payment_type_prepaid_tax_amounts"),

			OutstandingBalancePMT: parseInt64(get(rec, "outstanding_balance_pmt")),
			OpenSale:              parseBool(get(rec, "open_sale")),
			OpenSaleType:          get(rec, "open_sale_type"),
			PurchaseType:          get(rec, "purchase_type"),
			OnlineBooking:         parseInt64(get(rec, "online_booking")),
			Void:                  parseInt64(get(rec, "void")),
			VoidedTransactionID:   get(rec, "voided_transaction_id"),
			VoidReason:            get(rec, "void_reason"),

			DepartmentID:   get(rec, "department_id"),
			DepartmentName: get(rec, "department_name"),

			StaffID:            get(rec, "staff_id"),
			StaffFirstName:     get(rec, "staff_first_name"),
			StaffLastName:      get(rec, "staff_last_name"),
			StaffCategoryID:    get(rec, "staff_category_id"),
			StaffCategoryName:  get(rec, "staff_category_name"),
			IsRequestedStaff:   parseInt(get(rec, "is_requested_staff")),
			PrimaryStaffID:     get(rec, "primary_staff_id"),
			PreferredStaffID:   get(rec, "preferred_staff_id"),
			PreferredStaffName: get(rec, "preferred_staff_name"),

			AppointmentID:      get(rec, "appointment_id"),
			AppointmentDate:    parseDate(get(rec, "appointment_date")),
			AppointmentCreated: parseTS(get(rec, "appointment_created")),
			AppointmentRating:  parseInt64(get(rec, "appointment_rating")),

			ClientBirthday:   parseDate(get(rec, "client_birthday")),
			ClientGender:     get(rec, "client_gender"),
			ClientEmail:      get(rec, "client_email"),
			ClientFirstVisit: parseDate(get(rec, "client_first_visit")),

			ApptClientID:         get(rec, "appt_client_id"),
			ApptClientFirstName:  get(rec, "appt_client_first_name"),
			ApptClientLastName:   get(rec, "appt_client_last_name"),
			ApptClientBirthday:   parseDate(get(rec, "appt_client_birthday")),
			ApptClientGender:     get(rec, "appt_client_gender"),
			ApptClientEmail:      get(rec, "appt_client_email"),
			ApptClientFirstVisit: parseDate(get(rec, "appt_client_first_visit")),

			InternetCategoryIDs:   get(rec, "internet_category_ids"),
			InternetCategoryNames: get(rec, "internet_category_names"),
			BranchProductID:       get(rec, "branch_product_id"),
			FixedDiscountID:       get(rec, "fixed_discount_id"),
			FixedDiscountName:     get(rec, "fixed_discount_name"),
			ClientCourseID:        get(rec, "client_course_id"),
			CreatingUser:          get(rec, "creating_user"),
			TaxRateName:           get(rec, "tax_rate_name"),
			SaleFeeID:             get(rec, "sale_fee_id"),

			UpdatedAtPhorest: parseTS(get(rec, "purchase_updated_at")),
		}
		items = append(items, item)

		// Build/refresh header (Transaction) by newest updated_at_phorest per transaction_id
		headerUpdated := item.UpdatedAtPhorest
		txNew := models.Transaction{
			TransactionID:    transactionID,
			BranchID:         item.BranchID,
			BranchName:       item.BranchName,
			ClientID:         item.ClientID,
			ClientFirstName:  item.ClientFirstName,
			ClientLastName:   item.ClientLastName,
			ClientSource:     item.ClientSource,
			PurchasedDate:    item.PurchasedDate,
			PurchaseTime:     item.PurchaseTime,
			UpdatedAtPhorest: headerUpdated,
		}
		if existing, ok := txByID[transactionID]; ok {
			// keep the newest watermark
			switch {
			case existing.UpdatedAtPhorest == nil && headerUpdated != nil:
				txByID[transactionID] = txNew
			case existing.UpdatedAtPhorest != nil && headerUpdated != nil &&
				headerUpdated.After(*existing.UpdatedAtPhorest):
				txByID[transactionID] = txNew
			default:
				// keep existing
			}
		} else {
			txByID[transactionID] = txNew
		}
	}

	transactions := make([]models.Transaction, 0, len(txByID))
	for _, v := range txByID {
		transactions = append(transactions, v)
	}

	lg.Printf("Parsed CSV: %d transactions, %d items", len(transactions), len(items))
	return &ParsedBatch{Transactions: transactions, Items: items}, nil
}
