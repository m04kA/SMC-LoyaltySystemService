package configure_loyalty

import (
	"context"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/service/loyalty/models"
)

// LoyaltyService интерфейс сервиса лояльности
type LoyaltyService interface {
	ConfigureLoyalty(ctx context.Context, companyID, userID int64, req *models.ConfigureLoyaltyRequest) (*models.LoyaltyConfigResponse, error)
}

// Logger интерфейс для логирования
type Logger interface {
	Info(format string, v ...interface{})
	Warn(format string, v ...interface{})
	Error(format string, v ...interface{})
}
