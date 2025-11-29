package models

import (
	"time"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/domain"
)

// CreateLoyaltyCardRequest запрос на создание карты лояльности
type CreateLoyaltyCardRequest struct {
	UserID    int64 `json:"user_id"`
	CompanyID int64 `json:"company_id"`
}

// LoyaltyCardResponse ответ с данными карты лояльности
type LoyaltyCardResponse struct {
	CardID             int64     `json:"card_id"`
	UserID             int64     `json:"user_id"`
	CompanyID          int64     `json:"company_id"`
	CardType           string    `json:"card_type"`
	Status             string    `json:"status"`
	DiscountPercentage float64   `json:"discount_percentage"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// ConfigureLoyaltyRequest запрос на настройку программы лояльности
type ConfigureLoyaltyRequest struct {
	DiscountPercentage float64 `json:"discount_percentage"`
	IsEnabled          *bool   `json:"is_enabled,omitempty"`
}

// LoyaltyConfigResponse ответ с данными конфигурации программы лояльности
type LoyaltyConfigResponse struct {
	CompanyID          int64     `json:"company_id"`
	CardType           string    `json:"card_type"`
	IsEnabled          bool      `json:"is_enabled"`
	DiscountPercentage float64   `json:"discount_percentage"`
	CreatedAt          time.Time `json:"created_at"`
	UpdatedAt          time.Time `json:"updated_at"`
}

// FromDomainLoyaltyCard конвертирует domain модель карты в DTO
func FromDomainLoyaltyCard(card *domain.LoyaltyCard) *LoyaltyCardResponse {
	return &LoyaltyCardResponse{
		CardID:             card.ID,
		UserID:             card.UserID,
		CompanyID:          card.CompanyID,
		CardType:           string(card.CardType),
		Status:             string(card.Status),
		DiscountPercentage: card.DiscountPercentage,
		CreatedAt:          card.CreatedAt,
		UpdatedAt:          card.UpdatedAt,
	}
}

// FromDomainLoyaltyConfig конвертирует domain модель конфигурации в DTO
func FromDomainLoyaltyConfig(config *domain.LoyaltyConfig) *LoyaltyConfigResponse {
	discountPercentage := 0.0
	if config.DiscountPercentage != nil {
		discountPercentage = *config.DiscountPercentage
	}

	return &LoyaltyConfigResponse{
		CompanyID:          config.CompanyID,
		CardType:           string(config.CardType),
		IsEnabled:          config.IsEnabled,
		DiscountPercentage: discountPercentage,
		CreatedAt:          config.CreatedAt,
		UpdatedAt:          config.UpdatedAt,
	}
}
