package loyalty

import (
	"context"
	"errors"
	"fmt"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/domain"
	cardRepo "github.com/m04kA/SMC-LoyaltySystemService/internal/infra/storage/loyalty_card"
	configRepo "github.com/m04kA/SMC-LoyaltySystemService/internal/infra/storage/loyalty_config"
	sellerClient "github.com/m04kA/SMC-LoyaltySystemService/internal/integrations/sellerservice"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/service/loyalty/models"
)

type Service struct {
	cardRepo     LoyaltyCardRepository
	configRepo   LoyaltyConfigRepository
	sellerClient SellerServiceClient
}

func NewService(
	cardRepo LoyaltyCardRepository,
	configRepo LoyaltyConfigRepository,
	sellerClient SellerServiceClient,
) *Service {
	return &Service{
		cardRepo:     cardRepo,
		configRepo:   configRepo,
		sellerClient: sellerClient,
	}
}

// GetCard получает карту лояльности клиента в компании
func (s *Service) GetCard(ctx context.Context, userID, companyID int64) (*models.LoyaltyCardResponse, error) {
	// 1. Сначала проверяем, что программа лояльности включена для компании
	config, err := s.configRepo.GetByCompanyID(ctx, companyID)
	if err != nil {
		if errors.Is(err, configRepo.ErrConfigNotFound) {
			return nil, ErrConfigNotFound
		}
		return nil, fmt.Errorf("%w: GetCard - failed to get config: %v", ErrInternal, err)
	}

	if !config.IsEnabled {
		return nil, ErrConfigDisabled
	}

	// 2. Только если программа лояльности включена - получаем карту клиента
	card, err := s.cardRepo.GetByUserAndCompany(ctx, userID, companyID)
	if err != nil {
		if errors.Is(err, cardRepo.ErrCardNotFound) {
			return nil, ErrCardNotFound
		}
		return nil, fmt.Errorf("%w: GetCard - repository error: %v", ErrInternal, err)
	}

	return models.FromDomainLoyaltyCard(card), nil
}

// CreateCard создает новую карту лояльности для клиента
func (s *Service) CreateCard(ctx context.Context, req *models.CreateLoyaltyCardRequest) (*models.LoyaltyCardResponse, error) {
	// 1. Получаем конфигурацию программы лояльности компании
	config, err := s.configRepo.GetByCompanyID(ctx, req.CompanyID)
	if err != nil {
		if errors.Is(err, configRepo.ErrConfigNotFound) {
			return nil, ErrConfigNotFound
		}
		return nil, fmt.Errorf("%w: CreateCard - failed to get config: %v", ErrInternal, err)
	}

	// 2. Проверяем, что программа лояльности включена
	if !config.IsEnabled {
		return nil, ErrConfigDisabled
	}

	// 3. Создаем карту с параметрами из конфигурации
	card := &domain.LoyaltyCard{
		UserID:             req.UserID,
		CompanyID:          req.CompanyID,
		CardType:           config.CardType,
		Status:             domain.CardStatusActive,
		DiscountPercentage: *config.DiscountPercentage,
	}

	createdCard, err := s.cardRepo.Create(ctx, card)
	if err != nil {
		if errors.Is(err, cardRepo.ErrCardAlreadyExists) {
			return nil, ErrCardAlreadyExists
		}
		return nil, fmt.Errorf("%w: CreateCard - repository error: %v", ErrInternal, err)
	}

	return models.FromDomainLoyaltyCard(createdCard), nil
}

// ConfigureLoyalty настраивает программу лояльности компании
// Требует проверки прав: пользователь должен быть менеджером компании
func (s *Service) ConfigureLoyalty(ctx context.Context, companyID, userID int64, req *models.ConfigureLoyaltyRequest) (*models.LoyaltyConfigResponse, error) {
	// 1. Проверяем права доступа через SellerService
	if err := s.checkManagerAccess(ctx, companyID, userID); err != nil {
		return nil, err
	}

	// 2. Валидируем discount_percentage
	if req.DiscountPercentage < 0 || req.DiscountPercentage > 100 {
		return nil, fmt.Errorf("%w: discount percentage must be between 0 and 100", ErrInvalidInput)
	}

	// 3. Проверяем, существует ли уже конфигурация
	existingConfig, err := s.configRepo.GetByCompanyID(ctx, companyID)

	if err != nil && !errors.Is(err, configRepo.ErrConfigNotFound) {
		return nil, fmt.Errorf("%w: ConfigureLoyalty - failed to check existing config: %v", ErrInternal, err)
	}

	var config *domain.LoyaltyConfig

	// 4. Если конфигурация существует - обновляем
	if existingConfig != nil {
		// Дефолтное значение isEnabled - текущее значение из БД
		isEnabled := existingConfig.IsEnabled
		// Если пришло новое значение - используем его
		if req.IsEnabled != nil {
			isEnabled = *req.IsEnabled
		}

		updateInput := domain.UpdateLoyaltyConfigInput{
			CompanyID:          companyID,
			IsEnabled:          &isEnabled,
			DiscountPercentage: &req.DiscountPercentage,
		}

		config, err = s.configRepo.Update(ctx, updateInput)
		if err != nil {
			return nil, fmt.Errorf("%w: ConfigureLoyalty - failed to update config: %v", ErrInternal, err)
		}
	} else {
		// 5. Если конфигурации нет - создаем новую
		// Дефолтное значение isEnabled = false, если не указано обратное
		isEnabled := false
		if req.IsEnabled != nil {
			isEnabled = *req.IsEnabled
		}

		createInput := domain.CreateLoyaltyConfigInput{
			CompanyID:          companyID,
			IsEnabled:          isEnabled,
			DiscountPercentage: req.DiscountPercentage,
		}

		config, err = s.configRepo.Create(ctx, createInput)
		if err != nil {
			if errors.Is(err, configRepo.ErrConfigAlreadyExists) {
				return nil, ErrConfigAlreadyExists
			}
			return nil, fmt.Errorf("%w: ConfigureLoyalty - failed to create config: %v", ErrInternal, err)
		}
	}

	return models.FromDomainLoyaltyConfig(config), nil
}

// checkManagerAccess проверяет, является ли пользователь менеджером компании
func (s *Service) checkManagerAccess(ctx context.Context, companyID, userID int64) error {
	// Получаем данные компании из SellerService
	company, err := s.sellerClient.GetCompany(ctx, companyID)
	if err != nil {
		// Проверяем типы ошибок от SellerService
		if errors.Is(err, sellerClient.ErrCompanyNotFound) {
			return ErrConfigNotFound // Компания не найдена = нельзя настроить лояльность
		}
		// Все остальные ошибки от SellerService оборачиваем с контекстом
		return fmt.Errorf("%w: seller service error: %v", ErrSellerServiceUnavailable, err)
	}

	// Проверяем, есть ли userID в списке менеджеров
	isManager := false
	for _, managerID := range company.ManagerIDs {
		if managerID == userID {
			isManager = true
			break
		}
	}

	if !isManager {
		return ErrAccessDenied
	}

	return nil
}
