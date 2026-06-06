package db

import (
	"context"
	"database/sql"
	"time"

	"chatApp/internal/config"
	"chatApp/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	_ "github.com/lib/pq"
)

func NewPostgres(dsn string) (*sql.DB, error) {
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		return nil, err
	}

	// Optimized connection pool settings
	db.SetMaxOpenConns(25)                      // Increased for better concurrency
	db.SetMaxIdleConns(10)                      // Keep more idle connections
	db.SetConnMaxLifetime(time.Hour)            // Recycle connections after 1 hour
	db.SetConnMaxIdleTime(time.Minute * 30)     // Close idle connections after 30 minutes

	// Test connection with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}

	return db, nil
}

// InitGormDB initializes a GORM DB connection and runs AutoMigrate for models.
func InitGormDB(cfg *config.Config) (*gorm.DB, error) {
	gdb, err := gorm.Open(postgres.Open(cfg.DBURL), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	// configure underlying sql.DB connection pool
	sqlDB, err := gdb.DB()
	if err == nil {
		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(10)
		sqlDB.SetConnMaxLifetime(time.Hour)
		sqlDB.SetConnMaxIdleTime(30 * time.Minute)
	}

	// Auto-migrate models
	if err := gdb.AutoMigrate(&model.User{}, &model.Chat{}, &model.ChatParticipant{}, &model.Message{}); err != nil {
		return nil, err
	}

	return gdb, nil
}