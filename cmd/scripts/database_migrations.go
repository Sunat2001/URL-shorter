package main

import (
	"github.com/joho/godotenv"
	baselog "log"
	"os"
	"url-shortner/internel/config"
	"url-shortner/internel/lib/logger/scriptLogger"
	"url-shortner/internel/storage/sqlite"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		baselog.Fatal("Error loading .env file")
	}
	cfg := config.MustLoad()
	log := scriptLogger.SetupLogger(cfg.Env)

	storage, err := sqlite.New(cfg.StoragePath)
	if err != nil {
		log.Error("failed to initialize storage: %v", err)
		os.Exit(1)
	}
	defer storage.CloseConnection()

	err = storage.RunMigrations(log)
	if err != nil {
		log.Error("failed to run migrations: %v", err)
		os.Exit(1)
	}
	log.Info("Migrations applied successfully")
}
