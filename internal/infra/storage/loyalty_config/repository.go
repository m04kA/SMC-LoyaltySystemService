package loyalty_config

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/domain"
	"github.com/m04kA/SMC-LoyaltySystemService/pkg/psqlbuilder"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

// Repository репозиторий для работы с конфигурациями программ лояльности
type Repository struct {
	db DBExecutor
}

// NewRepository создает новый экземпляр репозитория конфигураций
func NewRepository(db DBExecutor) *Repository {
	return &Repository{db: db}
}

// GetByCompanyID получает конфигурацию программы лояльности компании
func (r *Repository) GetByCompanyID(ctx context.Context, companyID int64) (*domain.LoyaltyConfig, error) {
	query, args, err := psqlbuilder.Select(
		"id", "company_id", "card_type", "is_enabled", "discount_percentage",
		"progressive_config", "points_config", "created_at", "updated_at",
	).
		From("loyalty_configs").
		Where(squirrel.Eq{"company_id": companyID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%w: GetByCompanyID - build select query: %v", ErrBuildQuery, err)
	}

	var config domain.LoyaltyConfig
	var cardType string
	var createdAt, updatedAt sql.NullTime
	var discountPercentage sql.NullFloat64
	var progressiveConfig, pointsConfig []byte

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&config.ID,
		&config.CompanyID,
		&cardType,
		&config.IsEnabled,
		&discountPercentage,
		&progressiveConfig,
		&pointsConfig,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrConfigNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: GetByCompanyID - scan config: %v", ErrScanRow, err)
	}

	config.CardType = domain.CardType(cardType)
	config.CreatedAt = createdAt.Time
	config.UpdatedAt = updatedAt.Time

	if discountPercentage.Valid {
		config.DiscountPercentage = &discountPercentage.Float64
	}

	if len(progressiveConfig) > 0 {
		config.ProgressiveConfig = progressiveConfig
	}

	if len(pointsConfig) > 0 {
		config.PointsConfig = pointsConfig
	}

	return &config, nil
}

// Create создает новую конфигурацию программы лояльности
// Пока что поддерживается только discount_percentage для fixed_discount типа
// В будущем добавятся progressive_config и points_config для других типов
func (r *Repository) Create(ctx context.Context, input domain.CreateLoyaltyConfigInput) (*domain.LoyaltyConfig, error) {
	query, args, err := psqlbuilder.Insert("loyalty_configs").
		Columns("company_id", "card_type", "is_enabled", "discount_percentage").
		Values(input.CompanyID, string(domain.CardTypeFixedDiscount), input.IsEnabled, input.DiscountPercentage).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%w: Create - build insert query: %v", ErrBuildQuery, err)
	}

	var configID int64
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&configID, &createdAt, &updatedAt)
	if err != nil {
		// Проверяем на duplicate key (UNIQUE constraint violation на company_id)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pqErrCodeUniqueViolation {
			return nil, ErrConfigAlreadyExists
		}
		return nil, fmt.Errorf("%w: Create - insert config: %v", ErrExecQuery, err)
	}

	return &domain.LoyaltyConfig{
		ID:                 configID,
		CompanyID:          input.CompanyID,
		CardType:           domain.CardTypeFixedDiscount,
		IsEnabled:          input.IsEnabled,
		DiscountPercentage: &input.DiscountPercentage,
		CreatedAt:          createdAt.Time,
		UpdatedAt:          updatedAt.Time,
	}, nil
}

// Update обновляет конфигурацию программы лояльности
// Поддерживается обновление is_enabled и discount_percentage
func (r *Repository) Update(ctx context.Context, input domain.UpdateLoyaltyConfigInput) (*domain.LoyaltyConfig, error) {
	updateBuilder := psqlbuilder.Update("loyalty_configs").
		Where(squirrel.Eq{"company_id": input.CompanyID})

	hasUpdates := false

	if input.IsEnabled != nil {
		updateBuilder = updateBuilder.Set("is_enabled", *input.IsEnabled)
		hasUpdates = true
	}

	if input.DiscountPercentage != nil {
		updateBuilder = updateBuilder.Set("discount_percentage", *input.DiscountPercentage)
		hasUpdates = true
	}

	if !hasUpdates {
		return nil, fmt.Errorf("%w: Update - no fields to update", ErrBuildQuery)
	}

	// Добавляем RETURNING для получения обновлённых данных
	updateBuilder = updateBuilder.Suffix("RETURNING id, company_id, card_type, is_enabled, discount_percentage, progressive_config, points_config, created_at, updated_at")

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: Update - build update query: %v", ErrBuildQuery, err)
	}

	var config domain.LoyaltyConfig
	var cardType string
	var createdAt, updatedAt sql.NullTime
	var discountPercentage sql.NullFloat64
	var progressiveConfig, pointsConfig []byte

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&config.ID,
		&config.CompanyID,
		&cardType,
		&config.IsEnabled,
		&discountPercentage,
		&progressiveConfig,
		&pointsConfig,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrConfigNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: Update - scan updated config: %v", ErrScanRow, err)
	}

	config.CardType = domain.CardType(cardType)
	config.CreatedAt = createdAt.Time
	config.UpdatedAt = updatedAt.Time

	if discountPercentage.Valid {
		config.DiscountPercentage = &discountPercentage.Float64
	}

	if len(progressiveConfig) > 0 {
		config.ProgressiveConfig = progressiveConfig
	}

	if len(pointsConfig) > 0 {
		config.PointsConfig = pointsConfig
	}

	return &config, nil
}
