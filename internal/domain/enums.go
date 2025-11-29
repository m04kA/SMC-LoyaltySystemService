package domain

// CardType типы карт лояльности
type CardType string

const (
	// CardTypeFixedDiscount фиксированная скидка (текущая реализация)
	CardTypeFixedDiscount CardType = "fixed_discount"
	// CardTypeProgressiveDiscount прогрессивная скидка (будущее)
	CardTypeProgressiveDiscount CardType = "progressive_discount"
	// CardTypePointsBased накопительная система баллов (будущее)
	CardTypePointsBased CardType = "points_based"
)

// CardStatus статусы карт лояльности
type CardStatus string

const (
	// CardStatusActive активная карта
	CardStatusActive CardStatus = "active"
	// CardStatusDisabled выключенная карта
	CardStatusDisabled CardStatus = "disabled"
	// CardStatusSuspended приостановленная карта
	CardStatusSuspended CardStatus = "suspended"
	// CardStatusExpired истёкшая карта
	CardStatusExpired CardStatus = "expired"
)
