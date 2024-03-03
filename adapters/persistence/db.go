package persistence

import (
	"fmt"
	"time"

	"github.com/Buddy-Git/JITScheduler-svc/adapters/logger"
	"github.com/Buddy-Git/JITScheduler-svc/config"
	"github.com/Buddy-Git/JITScheduler-svc/model"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	log "gorm.io/gorm/logger"
)

var db *gorm.DB

func InitPostgresDatabase() (*gorm.DB, error) {
	cfg := config.GetConfig()
	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s", cfg.PgUser, cfg.PgPassword, cfg.PgHost, cfg.PgPort, cfg.DB)

	db, err := gorm.Open(postgres.Open(dbURL), &gorm.Config{
		Logger: log.Default.LogMode(log.Silent),
	})

	if err != nil {
		logger.Error("err", zap.Error(err))
		time.Sleep(10 * time.Second)
		InitPostgresDatabase()
	}

	//Auto Migration
	logger.Info("Auto migration")
	db.AutoMigrate(&model.Tenant{}, &model.Event{})

	// Get generic database object sql.DB to use its functions
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(cfg.MaxOpenConnections - 1)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConnections)
	logger.Info("Successfully created db connection")

	return db, nil
}

func GetDBConn() *gorm.DB {
	return db
}

func SetDBConn(conn *gorm.DB) {
	db = conn
}
