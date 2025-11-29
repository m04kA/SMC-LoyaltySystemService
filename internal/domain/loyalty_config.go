package domain

import (
	"errors"
	"time"
)

// LoyaltyConfig представляет конфигурацию программы лояльности компании
type LoyaltyConfig struct {
	ID                 int64
	CompanyID          int64
	CardType           CardType
	IsEnabled          bool
	DiscountPercentage *float64 // Для fixed_discount
	ProgressiveConfig  []byte   // JSONB для progressive_discount (будущее)
	PointsConfig       []byte   // JSONB для points_based (будущее)
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// CreateLoyaltyConfigInput входные данные для создания конфигурации
type CreateLoyaltyConfigInput struct {
	CompanyID          int64
	IsEnabled          bool
	DiscountPercentage float64
}

// UpdateLoyaltyConfigInput входные данные для обновления конфигурации
type UpdateLoyaltyConfigInput struct {
	CompanyID          int64
	IsEnabled          *bool
	DiscountPercentage *float64
}

// Validate проверяет корректность конфигурации программы лояльности
func (c *LoyaltyConfig) Validate() error {
	if c.DiscountPercentage == nil {
		return errors.New("discount percentage is required")
	}

	if *c.DiscountPercentage < 0 || *c.DiscountPercentage > 100 {
		return errors.New("discount percentage must be between 0 and 100")
	}

	return nil
}
