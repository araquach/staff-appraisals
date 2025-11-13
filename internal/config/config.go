package config

import (
	"log"
	"os"

	"staff-appraisals/internal/util"
)

// BranchConfig holds the name + ID of each branch.
type BranchConfig struct {
	Name     string
	BranchID string
}

// Config centralises all environment and runtime configuration.
type Config struct {
	Logger          *log.Logger
	DatabaseURL     string
	PhorestUsername string
	PhorestPassword string
	PhorestBusiness string

	Branches []BranchConfig

	AutoMigrate bool
}

// Load builds the Config struct, validating critical env vars.
func Load() *Config {
	logger := util.NewLogger()
	logger.Println("Loading environment configuration...")

	cfg := &Config{
		Logger:          logger,
		DatabaseURL:     getEnvOrFail(logger, "DATABASE_URL"),
		PhorestUsername: getEnvOrFail(logger, "PHOREST_USERNAME"),
		PhorestPassword: getEnvOrFail(logger, "PHOREST_PASSWORD"),
		PhorestBusiness: getEnvOrFail(logger, "PHOREST_BUSINESS"),
		AutoMigrate:     os.Getenv("AUTO_MIGRATE") == "1",
		Branches: []BranchConfig{
			{
				Name:     getEnvOrDefault("SITE_1_NAME", "Jakata"),
				BranchID: getEnvOrFail(logger, "SITE_1_BRANCH_ID"),
			},
			{
				Name:     getEnvOrDefault("SITE_2_NAME", "PK"),
				BranchID: getEnvOrFail(logger, "SITE_2_BRANCH_ID"),
			},
			{
				Name:     getEnvOrDefault("SITE_3_NAME", "Base"),
				BranchID: getEnvOrFail(logger, "SITE_3_BRANCH_ID"),
			},
		},
	}

	logger.Printf("✅ Loaded config for %d branches\n", len(cfg.Branches))
	return cfg
}

func getEnvOrFail(logger *log.Logger, key string) string {
	val := os.Getenv(key)
	if val == "" {
		logger.Fatalf("❌ Environment variable %s is required but not set", key)
	}
	return val
}

func getEnvOrDefault(key, def string) string {
	val := os.Getenv(key)
	if val == "" {
		return def
	}
	return val
}
