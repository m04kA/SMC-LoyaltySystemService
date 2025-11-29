package create_loyalty_card

import (
	"context"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/service/loyalty/models"
)

// LoyaltyService интерфейс сервиса лояльности
type LoyaltyService interface {
	CreateCard(ctx context.Context, req *models.CreateLoyaltyCardRequest) (*models.LoyaltyCardResponse, error)
}

// Logger интерфейс для логирования
type Logger interface {
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}
