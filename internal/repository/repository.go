package repository

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"subscriptions/internal/config"
	"subscriptions/internal/model"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func InitDB(cfg *config.Config) (*gorm.DB, error) {
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.DBHost, cfg.DBPort, cfg.DBUser, cfg.DBPass, cfg.DBName,
	)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&model.Subscription{}); err != nil {
		return nil, fmt.Errorf("failed to migrate: %w", err)
	}

	return db, nil
}

type Repository interface {
	Create(sub *model.Subscription) error
	GetByID(id uuid.UUID) (*model.Subscription, error)
	Update(sub *model.Subscription) error
	Delete(id uuid.UUID) error
	List(userID *uuid.UUID, serviceName *string) ([]model.Subscription, error)
}

type repo struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
	return &repo{db: db}
}

func (r *repo) Create(sub *model.Subscription) error {
	if sub.ID == uuid.Nil {
		sub.ID = uuid.New()
	}
	log.Printf("Creating subscription with ID: %s", sub.ID.String())
	return r.db.Create(sub).Error
}

func (r *repo) GetByID(id uuid.UUID) (*model.Subscription, error) {
	var sub model.Subscription
	if err := r.db.First(&sub, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &sub, nil
}

func (r *repo) Update(sub *model.Subscription) error {
	return r.db.Save(sub).Error
}

func (r *repo) Delete(id uuid.UUID) error {
	return r.db.Delete(&model.Subscription{}, "id = ?", id).Error
}

func (r *repo) List(userID *uuid.UUID, serviceName *string) ([]model.Subscription, error) {
	var subs []model.Subscription
	query := r.db.Model(&model.Subscription{})
	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if serviceName != nil && *serviceName != "" {
		query = query.Where("service_name = ?", *serviceName)
	}
	if err := query.Find(&subs).Error; err != nil {
		return nil, err
	}
	return subs, nil
}
