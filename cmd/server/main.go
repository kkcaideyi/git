package main

import (
	sqlite "github.com/glebarez/sqlite"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"issue-pm/config"
	"issue-pm/model"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal(err)
	}
	if cfg.DatabaseURL == "" || cfg.JWTSecret == "" {
		log.Fatal("DATABASE_URL and JWT_SECRET are required")
	}
	var db *gorm.DB
	if strings.HasPrefix(cfg.DatabaseURL, "sqlite://") {
		path := strings.TrimPrefix(cfg.DatabaseURL, "sqlite://")
		if dir := filepath.Dir(path); dir != "." {
			if err = os.MkdirAll(dir, 0755); err != nil {
				log.Fatal(err)
			}
		}
		db, err = gorm.Open(sqlite.Open(path), &gorm.Config{})
	} else {
		db, err = gorm.Open(postgres.Open(cfg.DatabaseURL), &gorm.Config{})
	}
	if err != nil {
		log.Fatal(err)
	}
	if cfg.AppEnv != "production" {
		if err = db.AutoMigrate(&model.User{}, &model.Project{}, &model.ProjectMember{}, &model.Issue{}, &model.Comment{}, &model.AuditLog{}); err != nil {
			log.Fatal(err)
		}
	}
	log.Printf("listening on :%s", cfg.HTTPPort)
	log.Fatal(newRouter(cfg, db).Run(":" + cfg.HTTPPort))
}
