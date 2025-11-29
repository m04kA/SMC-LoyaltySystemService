package get_loyalty_card

import (
	"context"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/service/loyalty/models"
)

// LoyaltyService интерфейс сервиса лояльности
type LoyaltyService interface {
	GetCard(ctx context.Context, userID, companyID int64) (*models.LoyaltyCardResponse, error)
}

// Logger интерфейс для логирования
type Logger interface {
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}
