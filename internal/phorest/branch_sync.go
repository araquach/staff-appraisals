package phorest

import "staff-appraisals/internal/repos"

func (r *Runner) SyncBranchesFromAPI() error {
	c := NewBranchClient(r.Cfg.PhorestUsername, r.Cfg.PhorestPassword, r.Cfg.PhorestBusiness, r.Logger)
	repo := repos.NewBranchRepo(r.DB, r.Logger)

	rows, err := c.FetchBranches()
	if err != nil {
		r.Logger.Printf("❌ branch fetch failed: %v", err)
		return err
	}
	if len(rows) == 0 {
		r.Logger.Printf("⚠️  no branches found from API")
		return nil
	}
	if err := repo.UpsertMany(rows); err != nil {
		r.Logger.Printf("❌ branch upsert failed: %v", err)
		return err
	}
	r.Logger.Printf("✅ branches upserted: %d", len(rows))
	return nil
}
