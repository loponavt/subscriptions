package repository

import (
	"fmt"
	"github.com/google/uuid"
	"log"
	"subscriptions/internal/config"
	"subscriptions/internal/model"
	"time"

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
	List(userID *uuid.UUID, serviceName *string, limit, offset int) (*model.SubscriptionList, error)
	CalculateTotal(userID *uuid.UUID, serviceName *string, from, to time.Time) (int, error)
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

func (r *repo) List(userID *uuid.UUID, serviceName *string, limit, offset int) (*model.SubscriptionList, error) {
	var subs []model.Subscription
	baseQuery := r.db.Model(&model.Subscription{})

	if userID != nil {
		baseQuery = baseQuery.Where("user_id = ?", *userID)
	}
	if serviceName != nil {
		baseQuery = baseQuery.Where("service_name = ?", *serviceName)
	}

	// Получаем total
	var total int64
	if err := baseQuery.Count(&total).Error; err != nil {
		return nil, err
	}

	// Получаем данные с пагинацией
	if err := baseQuery.
		Limit(limit).
		Offset(offset).
		Find(&subs).Error; err != nil {
		return nil, err
	}

	return &model.SubscriptionList{
		Total: total,
		Items: subs,
	}, nil
}

func (r *repo) CalculateTotal(userID *uuid.UUID, serviceName *string, from, to time.Time) (int, error) {
	var total int

	// считаем количество месяцев пересечения и умножаем на price
	// Используем GREATEST/LEAST для выбора пересекающегося диапазона
	// CASE WHEN months < 1 THEN 1 ELSE months END — чтобы минимальный период был 1 месяц
	query := r.db.Model(&model.Subscription{}).
		Select(`
			COALESCE(SUM(price * GREATEST(1, DATE_PART('month', AGE(LEAST(COALESCE(end_date, NOW()), ?), GREATEST(start_date, ?))))), 0) as total
		`, to, from)

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if serviceName != nil {
		query = query.Where("service_name = ?", *serviceName)
	}

	// Пересечение диапазонов
	query = query.Where("(start_date <= ?) AND (COALESCE(end_date, NOW()) >= ?)", to, from)

	if err := query.Pluck("total", &total).Error; err != nil {
		return 0, err
	}

	return total, nil
}
