package model

import (
	"github.com/google/uuid"
	"time"
)

type Subscription struct {
	ID          uuid.UUID  `gorm:"type:uuid;default:uuid_generate_v4();primaryKey" json:"id,omitempty"`
	ServiceName string     `json:"service_name" db:"service_name"`
	Price       int        `json:"price" db:"price"`
	UserID      uuid.UUID  `json:"user_id" db:"user_id"`
	StartDate   time.Time  `json:"start_date" db:"start_date"` // формат "07-2025"
	EndDate     *time.Time `json:"end_date,omitempty" db:"end_date"`
}

// SubscriptionReq represents a subscription creation request
// swagger:model
type SubscriptionReq struct {
	// ServiceName is the name of the subscription service
	// required: true
	ServiceName string `json:"service_name" binding:"required"`
	// Price subscription price, must be positive
	// required: true
	Price int `json:"price" binding:"required"`
	// UserID owner of the subscription (UUID)
	// required: true
	UserID uuid.UUID `json:"user_id" binding:"required"`
	// StartDate subscription start date in MM-YYYY format
	// required: true
	StartDate string `json:"start_date" binding:"required" example:"07-2025"`
	// EndDate optional subscription end date in MM-YYYY format
	EndDate *string `json:"end_date,omitempty"`
}

type SubscriptionList struct {
	Total int64          `json:"total"`
	Items []Subscription `json:"items"`
}
