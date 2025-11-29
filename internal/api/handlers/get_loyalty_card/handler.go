package get_loyalty_card

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/api/handlers"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/service/loyalty"
)

const (
	msgInvalidUserID    = "некорректный или отсутствующий параметр userId"
	msgInvalidCompanyID = "некорректный или отсутствующий параметр companyId"
	msgCardNotFound     = "карта лояльности не найдена"
	msgConfigNotFound   = "программа лояльности не настроена для данной компании"
	msgConfigDisabled   = "программа лояльности отключена для данной компании"
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

// Handle GET /api/v1/loyalty-cards?userId={userId}&companyId={companyId}
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	// 1. Парсим query параметры
	userIDStr := r.URL.Query().Get("userId")
	companyIDStr := r.URL.Query().Get("companyId")

	if userIDStr == "" {
		h.logger.Warn("GET /loyalty-cards - Missing userId parameter")
		handlers.RespondBadRequest(w, msgInvalidUserID)
		return
	}

	if companyIDStr == "" {
		h.logger.Warn("GET /loyalty-cards - Missing companyId parameter")
		handlers.RespondBadRequest(w, msgInvalidCompanyID)
		return
	}

	userID, err := strconv.ParseInt(userIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("GET /loyalty-cards - Invalid userId: %v", err)
		handlers.RespondBadRequest(w, msgInvalidUserID)
		return
	}

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("GET /loyalty-cards - Invalid companyId: %v", err)
		handlers.RespondBadRequest(w, msgInvalidCompanyID)
		return
	}

	// 2. Вызываем сервис
	card, err := h.service.GetCard(r.Context(), userID, companyID)
	if err != nil {
		if errors.Is(err, loyalty.ErrCardNotFound) {
			h.logger.Warn("GET /loyalty-cards - Card not found: user_id=%d, company_id=%d", userID, companyID)
			handlers.RespondNotFound(w, msgCardNotFound)
			return
		}
		if errors.Is(err, loyalty.ErrConfigNotFound) {
			h.logger.Warn("GET /loyalty-cards - Config not found: company_id=%d", companyID)
			handlers.RespondNotFound(w, msgConfigNotFound)
			return
		}
		if errors.Is(err, loyalty.ErrConfigDisabled) {
			h.logger.Warn("GET /loyalty-cards - Config disabled: company_id=%d", companyID)
			handlers.RespondForbidden(w, msgConfigDisabled)
			return
		}
		h.logger.Error("GET /loyalty-cards - Failed to get card: user_id=%d, company_id=%d, error=%v", userID, companyID, err)
		handlers.RespondInternalError(w)
		return
	}

	// 3. Возвращаем успешный ответ
	h.logger.Info("GET /loyalty-cards - Card retrieved successfully: user_id=%d, company_id=%d, card_id=%d", userID, companyID, card.CardID)
	handlers.RespondJSON(w, http.StatusOK, card)
}
