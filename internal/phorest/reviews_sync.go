package phorest

import (
	"github.com/araquach/phorest-datahub/internal/repos"
)

func (r *Runner) SyncReviewsFromAPI() error {
	client := NewReviewsClient(r.Cfg.PhorestUsername, r.Cfg.PhorestPassword, r.Cfg.PhorestBusiness)
	repo := repos.NewReviewsRepo(r.DB, r.Logger)

	for _, b := range r.Cfg.Branches {
		if b.BranchID == "" {
			r.Logger.Printf("⚠️  Skipping reviews: empty BranchID (name=%q)", b.Name)
			continue
		}

		// Watermark: last review_date we have for this branch
		since, err := repo.MaxReviewDate(b.BranchID)
		if err != nil {
			r.Logger.Printf("❌ reviews watermark error for %s: %v", b.Name, err)
			continue
		}
		if since != nil {
			r.Logger.Printf("Reviews watermark for %s: since %s", b.Name, *since)
		} else {
			r.Logger.Printf("Reviews watermark for %s: none (full sync)", b.Name)
		}

		page, totalPages := 0, 1
		for page < totalPages {
			rows, tp, err := client.FetchReviews(b.BranchID, valueOrEmpty(since), page, 200)
			if err != nil {
				r.Logger.Printf("❌ reviews fetch failed for %s p%d: %v", b.Name, page, err)
				break
			}
			totalPages = tp
			if len(rows) == 0 {
				r.Logger.Printf("No reviews on page %d for %s", page, b.Name)
				page++
				continue
			}
			if err := repo.UpsertMany(rows); err != nil {
				r.Logger.Printf("❌ reviews upsert failed for %s p%d: %v", b.Name, page, err)
				break
			}
			r.Logger.Printf("✅ reviews upserted for %s p%d: %d", b.Name, page, len(rows))
			page++
		}
	}
	return nil
}

// SyncLatestReviewsFromAPI fetches only the latest N reviews per branch and upserts them.
func (r *Runner) SyncLatestReviewsFromAPI(n int) error {
	client := NewReviewsClient(r.Cfg.PhorestUsername, r.Cfg.PhorestPassword, r.Cfg.PhorestBusiness)
	repo := repos.NewReviewsRepo(r.DB, r.Logger)

	for _, b := range r.Cfg.Branches {
		if b.BranchID == "" {
			r.Logger.Printf("⚠️  Skipping reviews: empty BranchID (name=%q)", b.Name)
			continue
		}
		r.Logger.Printf("Fetching latest %d reviews for %s (%s)", n, b.Name, b.BranchID)

		rows, err := client.FetchLatestN(b.BranchID, n)
		if err != nil {
			r.Logger.Printf("❌ reviews fetch failed for %s: %v", b.Name, err)
			continue
		}
		if len(rows) == 0 {
			r.Logger.Printf("No reviews returned for %s", b.Name)
			continue
		}
		if err := repo.UpsertMany(rows); err != nil {
			r.Logger.Printf("❌ reviews upsert failed for %s: %v", b.Name, err)
			continue
		}
		r.Logger.Printf("✅ upserted %d latest reviews for %s", len(rows), b.Name)
	}
	return nil
}

func valueOrEmpty(ps *string) string {
	if ps == nil {
		return ""
	}
	return *ps
}
