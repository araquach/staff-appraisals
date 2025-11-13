package phorest

import (
	"staff-appraisals/internal/repos"
)

// SyncStaffFromAPI fetches staff for each configured branch and upserts them.
func (r *Runner) SyncStaffFromAPI() error {
	c := NewStaffClient(r.Cfg.PhorestUsername, r.Cfg.PhorestPassword, r.Cfg.PhorestBusiness, r.Logger)
	repo := repos.NewStaffRepo(r.DB, r.Logger)

	for _, b := range r.Cfg.Branches {
		if b.BranchID == "" {
			r.Logger.Printf("⚠️  Skipping branch with empty BranchID (name=%q)", b.Name)
			continue
		}
		r.Logger.Printf("Fetching staff for %s (%s)", b.Name, b.BranchID)

		rows, err := c.FetchStaff(b.BranchID)
		if err != nil {
			r.Logger.Printf("❌ staff fetch failed for %s (%s): %v", b.Name, b.BranchID, err)
			continue
		}
		if len(rows) == 0 {
			r.Logger.Printf("No staff to upsert for %s (%s)", b.Name, b.BranchID)
			continue
		}
		if err := repo.UpsertMany(rows); err != nil {
			r.Logger.Printf("❌ staff upsert failed for %s (%s): %v", b.Name, b.BranchID, err)
			continue
		}
		r.Logger.Printf("✅ staff upserted for %s (%s): %d", b.Name, b.BranchID, len(rows))
	}
	return nil
}
