package usecase

import (
	"errors"
	"time"

	"github.com/google/uuid"
	"subscriptions/internal/model"
	"subscriptions/internal/repository"
)

var ErrSubscriptionNotFound = errors.New("subscription not found")

type Usecase struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Usecase {
	return &Usecase{repo: repo}
}

func (s *Usecase) CreateSubscription(sub *model.Subscription) error {
	return s.repo.Create(sub)
}

func (s *Usecase) GetSubscription(id uuid.UUID) (*model.Subscription, error) {
	return s.repo.GetByID(id)
}

func (s *Usecase) UpdateSubscription(sub *model.Subscription) error {
	return s.repo.Update(sub)
}

func (s *Usecase) DeleteSubscription(id uuid.UUID) error {
	return s.repo.Delete(id)
}

func (s *Usecase) ListSubscriptions(userID *uuid.UUID, serviceName *string, limit, offset int) (*model.SubscriptionList, error) {
	return s.repo.List(userID, serviceName, limit, offset)
}

// CalculateTotal Подсчёт суммарной стоимости подписок за период
func (s *Usecase) CalculateTotal(userID *uuid.UUID, serviceName *string, from, to time.Time) (int, error) {
	return s.repo.CalculateTotal(userID, serviceName, from, to)
}
