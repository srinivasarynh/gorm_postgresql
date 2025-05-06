package db

import (
	"context"
	"fmt"
	"time"

	"go_postgres/internal/config"

	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"
)

type PostgresDB struct {
	DB *gorm.DB
}

type GormLogAdapter struct {
	Logger *zap.Logger
}

func NewPostgresDB(cfg *config.DatabaseConfig, zapLogger *zap.Logger) (*PostgresDB, error) {
	gormLogger := logger.New(
		NewGormLogAdapter(zapLogger),
		logger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  false,
		},
	)

	db, err := gorm.Open(postgres.Open(cfg.GetDSN()), &gorm.Config{
		Logger: gormLogger,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   "app_",
			SingularTable: false,
		},
		PrepareStmt: true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetConnMaxLifetime(cfg.ConnMaxLife)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	zapLogger.Info("successfully connected to the database")
	return &PostgresDB{DB: db}, nil
}

func NewGormLogAdapter(zapLogger *zap.Logger) *GormLogAdapter {
	return &GormLogAdapter{Logger: zapLogger}
}

func (l *GormLogAdapter) Printf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	l.Logger.Info(msg)
}
