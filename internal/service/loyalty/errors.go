package loyalty

import "errors"

var (
	// ErrCardNotFound возвращается, когда карта лояльности не найдена
	ErrCardNotFound = errors.New("loyalty card not found")

	// ErrCardAlreadyExists возвращается, когда карта лояльности уже существует
	ErrCardAlreadyExists = errors.New("loyalty card already exists")

	// ErrConfigNotFound возвращается, когда программа лояльности не настроена для компании
	ErrConfigNotFound = errors.New("loyalty program not configured for this company")

	// ErrConfigDisabled возвращается, когда программа лояльности выключена
	ErrConfigDisabled = errors.New("loyalty program is disabled for this company")

	// ErrConfigAlreadyExists возвращается, когда программа лояльности уже настроена
	ErrConfigAlreadyExists = errors.New("loyalty program already configured for this company")

	// ErrAccessDenied возвращается, когда у пользователя нет прав доступа
	ErrAccessDenied = errors.New("access denied: user is not a manager of this company")

	// ErrSellerServiceUnavailable возвращается, когда SellerService недоступен
	ErrSellerServiceUnavailable = errors.New("seller service unavailable")

	// ErrInvalidInput возвращается при некорректных входных данных
	ErrInvalidInput = errors.New("invalid input data")

	// ErrInternal возвращается при внутренних ошибках сервиса
	ErrInternal = errors.New("service: internal error")
)
