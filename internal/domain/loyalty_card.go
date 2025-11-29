package domain

import "time"

// LoyaltyCard представляет карту лояльности клиента
type LoyaltyCard struct {
	ID                 int64
	UserID             int64
	CompanyID          int64
	CardType           CardType
	Status             CardStatus
	DiscountPercentage float64
	CreatedAt          time.Time
	UpdatedAt          time.Time
}

// CreateLoyaltyCardInput входные данные для создания карты лояльности
type CreateLoyaltyCardInput struct {
	UserID    int64
	CompanyID int64
}

// UpdateLoyaltyCardInput входные данные для обновления карты лояльности
type UpdateLoyaltyCardInput struct {
	// Поиск карты либо по ID
	CardID *int64
	// Либо по комбинации UserID + CompanyID
	UserID    *int64
	CompanyID *int64

	// Обновляемые поля
	CardType           *CardType
	Status             *CardStatus
	DiscountPercentage *float64
}
