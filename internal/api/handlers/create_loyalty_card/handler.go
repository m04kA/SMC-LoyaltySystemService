package create_loyalty_card

import (
	"errors"
	"net/http"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/api/handlers"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/service/loyalty"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/service/loyalty/models"
)

const (
	msgInvalidRequestBody = "некорректное тело запроса"
	msgConfigNotFound     = "программа лояльности не настроена для данной компании"
	msgConfigDisabled     = "программа лояльности отключена для данной компании"
	msgCardAlreadyExists  = "карта лояльности уже существует"
)

type Handler struct {
	service LoyaltyService
	logger  Logger
}

func NewHandler(service LoyaltyService, logger Logger) *Handler {
	return &Handler{
		service: service,
		logger:  logger,
	}
}

// Handle POST /api/v1/loyalty-cards
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим request body
	var req models.CreateLoyaltyCardRequest
	if err := handlers.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("POST /loyalty-cards - Invalid request body: %v", err)
		handlers.RespondBadRequest(w, msgInvalidRequestBody)
		return
	}

	// 2. Вызываем сервис
	card, err := h.service.CreateCard(r.Context(), &req)
	if err != nil {
		if errors.Is(err, loyalty.ErrConfigNotFound) {
			h.logger.Warn("POST /loyalty-cards - Config not found: company_id=%d", req.CompanyID)
			handlers.RespondNotFound(w, msgConfigNotFound)
			return
		}
		if errors.Is(err, loyalty.ErrConfigDisabled) {
			h.logger.Warn("POST /loyalty-cards - Config disabled: company_id=%d", req.CompanyID)
			handlers.RespondNotFound(w, msgConfigDisabled)
			return
		}
		if errors.Is(err, loyalty.ErrCardAlreadyExists) {
			h.logger.Warn("POST /loyalty-cards - Card already exists: user_id=%d, company_id=%d", req.UserID, req.CompanyID)
			handlers.RespondConflict(w, msgCardAlreadyExists)
			return
		}
		h.logger.Error("POST /loyalty-cards - Failed to create card: user_id=%d, company_id=%d, error=%v", req.UserID, req.CompanyID, err)
		handlers.RespondInternalError(w)
		return
	}

	// 3. Возвращаем успешный ответ
	h.logger.Info("POST /loyalty-cards - Card created successfully: user_id=%d, company_id=%d, card_id=%d", req.UserID, req.CompanyID, card.CardID)
	handlers.RespondJSON(w, http.StatusCreated, card)
}
