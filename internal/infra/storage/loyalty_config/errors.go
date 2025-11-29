package loyalty_config

import "errors"

const (
	// pqErrCodeUniqueViolation PostgreSQL код ошибки нарушения UNIQUE constraint
	pqErrCodeUniqueViolation = "23505"
)

var (
	// ErrConfigNotFound возвращается, когда конфигурация лояльности не найдена в БД
	ErrConfigNotFound = errors.New("repository.loyalty_config: config not found")

	// ErrConfigAlreadyExists возвращается, когда конфигурация уже существует
	ErrConfigAlreadyExists = errors.New("repository.loyalty_config: config already exists")

	// ErrBuildQuery возвращается при ошибке построения SQL запроса
	ErrBuildQuery = errors.New("repository.loyalty_config: failed to build SQL query")

	// ErrExecQuery возвращается при ошибке выполнения SQL запроса
	ErrExecQuery = errors.New("repository.loyalty_config: failed to execute SQL query")

	// ErrScanRow возвращается при ошибке сканирования строки из БД
	ErrScanRow = errors.New("repository.loyalty_config: failed to scan row")
)
