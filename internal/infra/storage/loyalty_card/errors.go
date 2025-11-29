package loyalty_card

import "errors"

const (
	// pqErrCodeUniqueViolation PostgreSQL код ошибки нарушения UNIQUE constraint
	pqErrCodeUniqueViolation = "23505"
)

var (
	// ErrCardNotFound возвращается, когда карта лояльности не найдена в БД
	ErrCardNotFound = errors.New("repository.loyalty_card: card not found")

	// ErrCardAlreadyExists возвращается, когда карта лояльности уже существует
	ErrCardAlreadyExists = errors.New("repository.loyalty_card: card already exists")

	// ErrBuildQuery возвращается при ошибке построения SQL запроса
	ErrBuildQuery = errors.New("repository.loyalty_card: failed to build SQL query")

	// ErrExecQuery возвращается при ошибке выполнения SQL запроса
	ErrExecQuery = errors.New("repository.loyalty_card: failed to execute SQL query")

	// ErrScanRow возвращается при ошибке сканирования строки из БД
	ErrScanRow = errors.New("repository.loyalty_card: failed to scan row")
)
