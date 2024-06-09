package main

import (
	"github.com/joho/godotenv"
	baselog "log"
	"os"
	"url-shortner/internel/config"
	"url-shortner/internel/lib/auth/hash"
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

	hashPassword, err := hash.GetHashPassword(os.Getenv("APP_PASSWORD"))
	if err != nil {
		log.Error("failed to hash password: %v", err)
		os.Exit(1)
	}

	_, err = storage.Query("INSERT INTO users(username, password) VALUES (?, ?)",
		os.Getenv("APP_USER"), hashPassword)
	if err != nil {
		log.Error("failed to create default user: %v", err)
		os.Exit(1)
	}

	log.Info("User created")
}
