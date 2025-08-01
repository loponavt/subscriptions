package usecase

import (
	"time"

	"github.com/google/uuid"
	"subscriptions/internal/model"
	"subscriptions/internal/repository"
)

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

func (s *Usecase) ListSubscriptions(userID *uuid.UUID, serviceName *string) ([]model.Subscription, error) {
	return s.repo.List(userID, serviceName)
}

// CalculateTotal Подсчёт суммарной стоимости подписок за период
func (s *Usecase) CalculateTotal(userID *uuid.UUID, serviceName *string, from, to time.Time) (int, error) {
	subs, err := s.repo.List(userID, serviceName)
	if err != nil {
		return 0, err
	}

	total := 0
	for _, sub := range subs {
		start := sub.StartDate
		end := sub.EndDate
		if end == nil {
			now := time.Now()
			end = &now
		}

		// проверяем пересечение с периодом [from, to]
		if (start.Before(to) || start.Equal(to)) && (end.After(from) || end.Equal(from)) {
			// считаем количество месяцев пересечения
			startPeriod := maxTime(start, from)
			endPeriod := minTime(*end, to)

			months := monthsBetween(startPeriod, endPeriod)
			if months < 1 {
				months = 1
			}
			total += sub.Price * months
		}
	}

	return total, nil
}

func maxTime(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}

func minTime(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

// monthsBetween считает количество полных месяцев между двумя датами включительно
func monthsBetween(from, to time.Time) int {
	yearDiff := to.Year() - from.Year()
	monthDiff := int(to.Month()) - int(from.Month())
	return yearDiff*12 + monthDiff + 1
}
