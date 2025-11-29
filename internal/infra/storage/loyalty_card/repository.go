package loyalty_card

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/domain"
	"github.com/m04kA/SMC-LoyaltySystemService/pkg/psqlbuilder"

	"github.com/Masterminds/squirrel"
	"github.com/lib/pq"
)

// Repository репозиторий для работы с картами лояльности
type Repository struct {
	db DBExecutor
}

// NewRepository создает новый экземпляр репозитория карт лояльности
func NewRepository(db DBExecutor) *Repository {
	return &Repository{db: db}
}

// GetByUserAndCompany получает карту лояльности клиента в компании
func (r *Repository) GetByUserAndCompany(ctx context.Context, userID, companyID int64) (*domain.LoyaltyCard, error) {
	query, args, err := psqlbuilder.Select(
		"id", "user_id", "company_id", "card_type", "status", "discount_percentage", "created_at", "updated_at",
	).
		From("loyalty_cards").
		Where(squirrel.Eq{"user_id": userID, "company_id": companyID}).
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%w: GetByUserAndCompany - build select query: %v", ErrBuildQuery, err)
	}

	var card domain.LoyaltyCard
	var cardType, status string
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&card.ID,
		&card.UserID,
		&card.CompanyID,
		&cardType,
		&status,
		&card.DiscountPercentage,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrCardNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: GetByUserAndCompany - scan card: %v", ErrScanRow, err)
	}

	card.CardType = domain.CardType(cardType)
	card.Status = domain.CardStatus(status)
	card.CreatedAt = createdAt.Time
	card.UpdatedAt = updatedAt.Time

	return &card, nil
}

// Create создает новую карту лояльности
func (r *Repository) Create(ctx context.Context, card *domain.LoyaltyCard) (*domain.LoyaltyCard, error) {
	query, args, err := psqlbuilder.Insert("loyalty_cards").
		Columns("user_id", "company_id", "card_type", "status", "discount_percentage").
		Values(card.UserID, card.CompanyID, string(card.CardType), string(card.Status), card.DiscountPercentage).
		Suffix("RETURNING id, created_at, updated_at").
		ToSql()

	if err != nil {
		return nil, fmt.Errorf("%w: Create - build insert query: %v", ErrBuildQuery, err)
	}

	var cardID int64
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query, args...).Scan(&cardID, &createdAt, &updatedAt)
	if err != nil {
		// Проверяем на duplicate key (UNIQUE constraint violation)
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == pqErrCodeUniqueViolation {
			return nil, ErrCardAlreadyExists
		}
		return nil, fmt.Errorf("%w: Create - insert card: %v", ErrExecQuery, err)
	}

	card.ID = cardID
	card.CreatedAt = createdAt.Time
	card.UpdatedAt = updatedAt.Time

	return card, nil
}

// Update обновляет карту лояльности
func (r *Repository) Update(ctx context.Context, input domain.UpdateLoyaltyCardInput) (*domain.LoyaltyCard, error) {
	// Строим WHERE clause в зависимости от того, что передано
	updateBuilder := psqlbuilder.Update("loyalty_cards")

	if input.CardID != nil {
		updateBuilder = updateBuilder.Where(squirrel.Eq{"id": *input.CardID})
	} else if input.UserID != nil && input.CompanyID != nil {
		updateBuilder = updateBuilder.Where(squirrel.Eq{"user_id": *input.UserID, "company_id": *input.CompanyID})
	} else {
		return nil, fmt.Errorf("%w: Update - either card_id or (user_id + company_id) must be provided", ErrBuildQuery)
	}

	// Добавляем только те поля, которые нужно обновить
	hasUpdates := false

	if input.CardType != nil {
		updateBuilder = updateBuilder.Set("card_type", string(*input.CardType))
		hasUpdates = true
	}

	if input.Status != nil {
		updateBuilder = updateBuilder.Set("status", string(*input.Status))
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
	updateBuilder = updateBuilder.Suffix("RETURNING id, user_id, company_id, card_type, status, discount_percentage, created_at, updated_at")

	query, args, err := updateBuilder.ToSql()
	if err != nil {
		return nil, fmt.Errorf("%w: Update - build update query: %v", ErrBuildQuery, err)
	}

	var card domain.LoyaltyCard
	var cardType, status string
	var createdAt, updatedAt sql.NullTime

	err = r.db.QueryRowContext(ctx, query, args...).Scan(
		&card.ID,
		&card.UserID,
		&card.CompanyID,
		&cardType,
		&status,
		&card.DiscountPercentage,
		&createdAt,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, ErrCardNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("%w: Update - scan updated card: %v", ErrScanRow, err)
	}

	card.CardType = domain.CardType(cardType)
	card.Status = domain.CardStatus(status)
	card.CreatedAt = createdAt.Time
	card.UpdatedAt = updatedAt.Time

	return &card, nil
}
