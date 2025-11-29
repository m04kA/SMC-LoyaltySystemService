package configure_loyalty

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"

	"github.com/m04kA/SMC-LoyaltySystemService/internal/api/handlers"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/api/middleware"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/service/loyalty"
	"github.com/m04kA/SMC-LoyaltySystemService/internal/service/loyalty/models"
)

const (
	msgMissingUserID      = "отсутствует заголовок X-User-ID"
	msgInvalidCompanyID   = "некорректный companyId"
	msgInvalidRequestBody = "некорректное тело запроса"
	msgAccessDenied       = "доступ запрещён: пользователь не является менеджером компании"
	msgInvalidInput       = "некорректные входные данные"
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

// Handle POST /api/v1/companies/{companyId}/loyalty-config
func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	// 1. Извлекаем user ID из контекста (установлен middleware.Auth)
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		h.logger.Warn("POST /companies/{companyId}/loyalty-config - Missing user ID in context")
		handlers.RespondUnauthorized(w, msgMissingUserID)
		return
	}

	// 2. Парсим companyId из URL
	vars := mux.Vars(r)
	companyIDStr := vars["companyId"]

	companyID, err := strconv.ParseInt(companyIDStr, 10, 64)
	if err != nil {
		h.logger.Warn("POST /companies/{companyId}/loyalty-config - Invalid companyId: %v", err)
		handlers.RespondBadRequest(w, msgInvalidCompanyID)
		return
	}

	// 3. Парсим request body
	var req models.ConfigureLoyaltyRequest
	if err := handlers.DecodeJSON(r, &req); err != nil {
		h.logger.Warn("POST /companies/{companyId}/loyalty-config - Invalid request body: %v", err)
		handlers.RespondBadRequest(w, msgInvalidRequestBody)
		return
	}

	// 4. Вызываем сервис
	config, err := h.service.ConfigureLoyalty(r.Context(), companyID, userID, &req)
	if err != nil {
		if errors.Is(err, loyalty.ErrAccessDenied) {
			h.logger.Warn("POST /companies/{companyId}/loyalty-config - Access denied: user_id=%d, company_id=%d", userID, companyID)
			handlers.RespondForbidden(w, msgAccessDenied)
			return
		}
		if errors.Is(err, loyalty.ErrInvalidInput) {
			h.logger.Warn("POST /companies/{companyId}/loyalty-config - Invalid input: company_id=%d, error=%v", companyID, err)
			handlers.RespondBadRequest(w, msgInvalidInput)
			return
		}
		if errors.Is(err, loyalty.ErrSellerServiceUnavailable) {
			h.logger.Error("POST /companies/{companyId}/loyalty-config - SellerService unavailable: company_id=%d, error=%v", companyID, err)
			handlers.RespondInternalError(w)
			return
		}
		h.logger.Error("POST /companies/{companyId}/loyalty-config - Failed to configure loyalty: user_id=%d, company_id=%d, error=%v", userID, companyID, err)
		handlers.RespondInternalError(w)
		return
	}

	// 5. Возвращаем успешный ответ
	h.logger.Info("POST /companies/{companyId}/loyalty-config - Loyalty configured successfully: user_id=%d, company_id=%d", userID, companyID)
	handlers.RespondJSON(w, http.StatusOK, config)
}
