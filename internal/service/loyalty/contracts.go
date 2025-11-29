package loyalty

import (
	"context"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/domain"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/integrations/sellerservice"
)

// LoyaltyCardRepository интерфейс репозитория карт лояльности
type LoyaltyCardRepository interface {
	GetByUserAndCompany(ctx context.Context, userID, companyID int64) (*domain.LoyaltyCard, error)
	Create(ctx context.Context, card *domain.LoyaltyCard) (*domain.LoyaltyCard, error)
	Update(ctx context.Context, input domain.UpdateLoyaltyCardInput) (*domain.LoyaltyCard, error)
}

// LoyaltyConfigRepository интерфейс репозитория конфигураций программ лояльности
type LoyaltyConfigRepository interface {
	GetByCompanyID(ctx context.Context, companyID int64) (*domain.LoyaltyConfig, error)
	Create(ctx context.Context, input domain.CreateLoyaltyConfigInput) (*domain.LoyaltyConfig, error)
	Update(ctx context.Context, input domain.UpdateLoyaltyConfigInput) (*domain.LoyaltyConfig, error)
}

// SellerServiceClient интерфейс клиента для взаимодействия с SellerService
type SellerServiceClient interface {
	// GetCompany получает данные компании по ID
	GetCompany(ctx context.Context, companyID int64) (*sellerservice.Company, error)
}
