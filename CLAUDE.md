# SMC-LoyaltySystemService

Микросервис для управления программами лояльности компаний (автомоек) в экосистеме SMC. Предоставляет API для создания виртуальных карт лояльности, настройки программ лояльности и генерации данных для QR-кодов.

## Место в экосистеме SMC

SMC-LoyaltySystemService является частью микросервисной архитектуры SMC (Slot My Car) и взаимодействует с:

- **SellerService (8081)** - валидация прав менеджеров компаний при настройке программ лояльности
- **UserService (8080)** - получение данных о пользователях (планируется)
- **BookingService (8082)** - получение истории посещений для прогрессивных скидок (планируется)

**Порт сервиса:** 8084
**База данных:** PostgreSQL на порту 5439

## Функциональность

### Текущая реализация (MVP)

**Фиксированная скидка:**
- Создание виртуальных карт лояльности для клиентов
- Настройка фиксированного процента скидки для всех клиентов компании
- Генерация данных карты для QR-кодов (card_id, user_id, company_id, discount_percentage)

### Будущие возможности (архитектура готова)

**Прогрессивная скидка:**
- Скидка зависит от истории обслуживания (количество посещений)
- Конфигурация через JSONB поле `progressive_config`
- Требует интеграцию с BookingService

**Накопительная система баллов:**
- Начисление и списание баллов
- Конфигурация через JSONB поле `points_config`
- Требует новую таблицу `loyalty_transactions`

## Архитектура

Проект следует принципам Clean Architecture и разделён на слои:

```
├── cmd/                    # Точки входа приложения
├── internal/              # Внутренняя бизнес-логика
│   ├── api/              # HTTP handlers и middleware
│   │   ├── handlers/
│   │   │   ├── get_loyalty_card/         # GET /api/v1/loyalty-cards
│   │   │   ├── create_loyalty_card/      # POST /api/v1/loyalty-cards
│   │   │   └── configure_loyalty/        # POST /api/v1/companies/{id}/loyalty-config
│   │   └── middleware/                   # Auth, Metrics
│   ├── service/          # Бизнес-логика
│   │   └── loyalty/                      # GetCard, CreateCard, ConfigureLoyalty
│   ├── infra/storage/    # Репозитории для работы с БД
│   │   ├── loyalty_card/                 # CRUD для карт
│   │   └── loyalty_config/               # CRUD для конфигураций
│   ├── domain/           # Доменные модели
│   │   ├── enums.go                      # CardType, CardStatus
│   │   ├── loyalty_card.go               # Модель карты лояльности
│   │   └── loyalty_config.go             # Модель конфигурации
│   ├── integrations/     # Внешние сервисы
│   │   └── sellerservice/                # Клиент для валидации прав
│   └── config/           # Конфигурация приложения
├── pkg/                   # Переиспользуемые пакеты
│   ├── metrics/          # Клиент метрик Prometheus
│   ├── dbmetrics/        # Обёртка над database/sql с метриками
│   ├── logger/           # Структурированное логирование
│   └── psqlbuilder/      # SQL query builder
├── migrations/           # Миграции базы данных
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── schemas/             # OpenAPI/Swagger спецификации
│   ├── schema.yaml                       # API спецификация LoyaltyService
│   └── clients/
│       └── smc-sellerservice.yaml        # API спецификация SellerService
└── test_data/           # Тестовые данные и примеры API-запросов
```

## Схема базы данных

### loyalty_cards

Таблица для хранения виртуальных карт лояльности клиентов:

```sql
CREATE TABLE loyalty_cards (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,              -- Telegram user ID клиента
    company_id BIGINT NOT NULL,            -- ID компании из SellerService
    card_type VARCHAR(50) NOT NULL DEFAULT 'fixed_discount',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    discount_percentage DECIMAL(5,2) NOT NULL CHECK (discount_percentage >= 0 AND discount_percentage <= 100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT loyalty_cards_unique_user_company UNIQUE (user_id, company_id)
);

-- Индексы для быстрого поиска
CREATE INDEX idx_loyalty_cards_user_id ON loyalty_cards(user_id);
CREATE INDEX idx_loyalty_cards_company_id ON loyalty_cards(company_id);
CREATE INDEX idx_loyalty_cards_user_company ON loyalty_cards(user_id, company_id);
CREATE INDEX idx_loyalty_cards_type ON loyalty_cards(card_type);
CREATE INDEX idx_loyalty_cards_status ON loyalty_cards(status);
```

**Типы карт (card_type):**
- `fixed_discount` - фиксированная скидка (текущая реализация)
- `progressive_discount` - прогрессивная скидка (будущее)
- `points_based` - накопительная система (будущее)

**Статусы карт (status):**
- `active` - активная карта
- `disabled` - выключенная карта
- `suspended` - приостановленная карта
- `expired` - истёкшая карта

### loyalty_configs

Таблица для хранения настроек программ лояльности компаний:

```sql
CREATE TABLE loyalty_configs (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL UNIQUE,     -- ID компании из SellerService
    card_type VARCHAR(50) NOT NULL DEFAULT 'fixed_discount',
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    discount_percentage DECIMAL(5,2) CHECK (discount_percentage >= 0 AND discount_percentage <= 100),
    progressive_config JSONB,              -- Для будущего: настройки прогрессивных скидок
    points_config JSONB,                   -- Для будущего: настройки накопительной системы
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT loyalty_configs_valid_card_type CHECK (card_type IN ('fixed_discount', 'progressive_discount', 'points_based'))
);

CREATE INDEX idx_loyalty_configs_company_id ON loyalty_configs(company_id);
CREATE INDEX idx_loyalty_configs_enabled ON loyalty_configs(is_enabled);
```

**Расширяемость через JSONB:**

Прогрессивная скидка (пример):
```json
{
  "progressive_config": {
    "tiers": [
      {"min_visits": 0, "discount": 5},
      {"min_visits": 5, "discount": 10},
      {"min_visits": 10, "discount": 15}
    ]
  }
}
```

Накопительная система (пример):
```json
{
  "points_config": {
    "points_per_ruble": 1,
    "redemption_rate": 0.01
  }
}
```

## API Endpoints

### 1. GET /api/v1/loyalty-cards

**Публичный endpoint** - получение карты лояльности клиента в компании.

**Query параметры:**
- `userId` (int64, required) - Telegram user ID клиента
- `companyId` (int64, required) - ID компании

**Response 200 OK:**
```json
{
  "card_id": 123,
  "user_id": 987654321,
  "company_id": 1,
  "card_type": "fixed_discount",
  "status": "active",
  "discount_percentage": 10.0,
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

**Возможные ошибки:**
- `400 Bad Request` - невалидные query параметры
- `404 Not Found` - карта не найдена
- `500 Internal Server Error` - внутренняя ошибка сервера

**Использование:**
Этот endpoint используется для генерации QR-кодов. Возвращаемые данные достаточны для MVP (в будущем планируется добавить HMAC подпись для защиты от подделки).

### 2. POST /api/v1/loyalty-cards

**Публичный endpoint** - создание новой карты лояльности для клиента.

**Request Body:**
```json
{
  "user_id": 987654321,
  "company_id": 1
}
```

**Response 201 Created:**
```json
{
  "card_id": 123,
  "user_id": 987654321,
  "company_id": 1,
  "card_type": "fixed_discount",
  "status": "active",
  "discount_percentage": 10.0,
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

**Условия создания карты:**
1. Программа лояльности настроена для компании (`loyalty_configs` существует)
2. Программа лояльности включена (`is_enabled = true`)
3. У клиента ещё нет карты в этой компании (проверка через UNIQUE constraint)

**Возможные ошибки:**
- `400 Bad Request` - невалидные данные в request body
- `404 Not Found` - программа лояльности не настроена для компании
- `409 Conflict` - карта лояльности уже существует
- `500 Internal Server Error` - внутренняя ошибка сервера

### 3. POST /api/v1/companies/{companyId}/loyalty-config

**Protected endpoint** - настройка программы лояльности компании (требует аутентификацию).

**Headers:**
- `X-User-ID` (int64, required) - Telegram user ID текущего пользователя

**Path параметры:**
- `companyId` (int64, required) - ID компании

**Request Body:**
```json
{
  "discount_percentage": 15.0
}
```

**Response 200 OK:**
```json
{
  "company_id": 1,
  "card_type": "fixed_discount",
  "is_enabled": true,
  "discount_percentage": 15.0,
  "created_at": "2025-01-15T10:00:00Z",
  "updated_at": "2025-01-15T10:00:00Z"
}
```

**Проверка прав доступа:**

Сервис проверяет права через интеграцию с SellerService:

1. Вызывается `GET /api/v1/companies/{companyId}` в SellerService
2. Проверяется наличие `X-User-ID` в массиве `manager_ids` компании
3. Если пользователь не является менеджером компании → `403 Forbidden`

**Возможные ошибки:**
- `400 Bad Request` - невалидные данные (например, процент скидки вне диапазона 0-100)
- `401 Unauthorized` - отсутствует или невалидный заголовок `X-User-ID`
- `403 Forbidden` - пользователь не является менеджером компании
- `500 Internal Server Error` - внутренняя ошибка сервера
- `503 Service Unavailable` - SellerService недоступен (не удалось проверить права)

### 4. GET /health

**Публичный endpoint** - проверка работоспособности сервиса.

**Response 200 OK:**
```json
{
  "status": "ok",
  "timestamp": "2025-01-15T10:00:00Z"
}
```

### 5. GET /metrics

**Публичный endpoint** - метрики Prometheus для мониторинга.

## Интеграции с внешними сервисами

### SellerService

Клиент для валидации прав менеджеров реализован в `internal/integrations/sellerservice/`.

**Используемые endpoints:**
- `GET /api/v1/companies/{id}` - получение данных компании и списка менеджеров

**Схема API:** `schemas/clients/smc-sellerservice.yaml`

**Пример использования:**

```go
// Получение данных компании для проверки прав
company, err := sellerClient.GetCompanyByID(ctx, companyID)
if err != nil {
    return nil, fmt.Errorf("failed to get company from SellerService: %w", err)
}

// Проверка, является ли пользователь менеджером
isManager := false
for _, managerID := range company.ManagerIDs {
    if managerID == userID {
        isManager = true
        break
    }
}

if !isManager {
    return nil, ErrAccessDenied
}
```

## Ключевые особенности

### 1. Система метрик Prometheus

Проект включает полноценную интеграцию с Prometheus для мониторинга производительности:

#### HTTP метрики
- `http_requests_total` - общее количество HTTP запросов
- `http_request_duration_seconds` - длительность HTTP запросов (histogram)
- `http_errors_total` - количество HTTP ошибок с категоризацией

#### Database метрики
- `db_queries_total` - общее количество запросов к БД
- `db_query_duration_seconds` - длительность запросов к БД (histogram)
- `db_errors_total` - количество ошибок БД с категоризацией
- `db_connections_active` - количество активных соединений
- `db_connections_idle` - количество простаивающих соединений
- `db_connections_max` - максимальное количество соединений

Все метрики доступны через endpoint `/metrics`.

#### Обёртки для клиентов

Пакет [pkg/dbmetrics](pkg/dbmetrics) предоставляет **обёртки над стандартными database/sql клиентами**, которые автоматически собирают метрики для всех операций с БД:

- `dbmetrics.DB` - обёртка над `*sql.DB`
- `dbmetrics.Tx` - обёртка над `*sql.Tx`

Обёртки перехватывают все операции (`QueryContext`, `ExecContext`, `BeginTx`, `Commit`, `Rollback`) и автоматически:
- Измеряют время выполнения запросов
- Парсят SQL запросы для извлечения типа операции и таблицы
- Категоризируют ошибки БД (duplicate key, foreign key violations, timeouts, и т.д.)
- Отправляют метрики в Prometheus

Пример использования:

```go
// Создаём обёрнутое соединение с автоматическим сбором метрик
wrappedDB := dbmetrics.WrapWithDefault(db, metricsCollector, serviceName, stopCh)

// Используем как обычный *sql.DB - метрики собираются автоматически
rows, err := wrappedDB.QueryContext(ctx, "SELECT * FROM loyalty_cards WHERE user_id = $1", userID)
```

Пакет [pkg/metrics](pkg/metrics) предоставляет централизованную структуру для управления всеми метриками приложения и методы для их записи.

Middleware [internal/api/middleware/metrics.go](internal/api/middleware/metrics.go) автоматически собирает HTTP метрики для всех endpoint'ов.

**Важно:** Все метрики автоматически собираются через обёртки над клиентами БД и HTTP middleware, не требуя дополнительного кода в handlers.

### 2. Конфигурация

Гибкая система конфигурации через TOML файлы с поддержкой переопределения через переменные окружения:

```toml
[logs]
level = "info"
file = "./logs/app.log"

[server]
http_port = 8084
read_timeout = 15
write_timeout = 15
idle_timeout = 60
shutdown_timeout = 10

[database]
host = "localhost"
port = 5439
user = "postgres"
password = "postgres"
dbname = "loyaltysystem"
sslmode = "disable"
max_open_conns = 25
max_idle_conns = 5
conn_max_lifetime = 300

[metrics]
enabled = true
path = "/metrics"
service_name = "loyaltysystemservice"

[sellerservice]
base_url = "http://localhost:8081"
timeout = 10
```

**Переменные окружения** автоматически переопределяют значения из конфига:
- `DB_HOST`, `DB_PORT`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_SSLMODE`
- `HTTP_PORT`
- `LOG_LEVEL`, `LOG_FILE`
- `METRICS_ENABLED`, `METRICS_PATH`, `METRICS_SERVICE_NAME`
- `SELLERSERVICE_BASE_URL`, `SELLERSERVICE_TIMEOUT`

См. [.env.example](.env.example) для полного списка переменных окружения.

### 3. Graceful Shutdown

Сервер поддерживает корректное завершение работы:
- Перехват сигналов `SIGINT` и `SIGTERM`
- Завершение обработки текущих запросов
- Остановка сборщиков метрик
- Закрытие соединений с БД

### 4. Middleware

- **Auth** - проверка заголовка `X-User-ID` (используется только для protected endpoints)
- **Metrics** - автоматический сбор HTTP метрик

**Важно:** Сервис использует только `X-User-ID` header. Валидация ролей происходит через интеграцию с SellerService (проверка `manager_ids`).

### 5. Структурированное логирование

Кастомный логгер ([pkg/logger](pkg/logger)) с поддержкой уровней (INFO, ERROR, FATAL) и записью в файл.

### 6. Query Builder

Пакет [pkg/psqlbuilder](pkg/psqlbuilder) предоставляет удобный интерфейс для построения SQL запросов с использованием библиотеки `squirrel`.

## Быстрый старт

### Локальный запуск

```bash
# Установить зависимости
GOPROXY=https://proxy.golang.org,direct go mod download

# Создать конфигурацию
cp .env.example .env

# Запустить PostgreSQL
docker-compose up -d postgres

# Применить миграции
make migrate-up

# Запустить приложение локально
make run
```

**Важно:** При первом запуске используйте `GOPROXY=https://proxy.golang.org,direct` для загрузки зависимостей.

### Docker

```bash
# Собрать и запустить все сервисы (app + postgres + migrations)
make docker-up

# Посмотреть логи
make docker-logs

# Остановить сервисы
make docker-down
```

## Тестирование API

### Автоматизированные тесты

Проект включает набор автоматизированных тестов для всех API endpoints.

**Файлы:**
- [test_data/README.md](test_data/README.md) - документация по тестированию
- [test_data/api_requests.sh](test_data/api_requests.sh) - исполняемый скрипт с тестами

**Структура тестов:**
- **Group 1:** Get Loyalty Card (6 тестов)
- **Group 2:** Create Loyalty Card (7 тестов)
- **Group 3:** Configure Loyalty (9 тестов)
- **Smoke Tests:** 4 основных сценария

**Запуск тестов:**

```bash
# Интерактивное меню
chmod +x test_data/api_requests.sh
./test_data/api_requests.sh

# Smoke тесты
./test_data/api_requests.sh
# Выбрать опцию 's'

# Отдельные функции
source test_data/api_requests.sh
tc_1_1         # Получение карты
tc_2_1         # Создание карты
tc_3_1         # Настройка конфигурации
run_group_1    # Вся группа тестов
```

**Требования:**
- `curl` - для HTTP запросов
- `jq` - для форматирования JSON (`brew install jq`)

### Примеры использования API

### Сценарий 1: Настройка программы лояльности менеджером

```bash
# 1. Менеджер настраивает программу лояльности для своей компании
curl -X POST http://localhost:8084/api/v1/companies/1/loyalty-config \
  -H "Content-Type: application/json" \
  -H "X-User-ID: 123456789" \
  -d '{
    "discount_percentage": 15.0
  }'
```

### Сценарий 2: Клиент создаёт карту лояльности

```bash
# 2. Клиент создаёт карту лояльности в компании
curl -X POST http://localhost:8084/api/v1/loyalty-cards \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 987654321,
    "company_id": 1
  }'
```

### Сценарий 3: Получение карты для QR-кода

```bash
# 3. Получение данных карты для генерации QR-кода
curl "http://localhost:8084/api/v1/loyalty-cards?userId=987654321&companyId=1"
```

## Паттерны и архитектурные решения

### Структура Handler'ов

Каждый handler организован как отдельный пакет с чёткой структурой:

```
internal/api/handlers/create_loyalty_card/
├── handler.go    # Основная логика обработки запроса
└── contract.go   # Интерфейсы зависимостей (Service, Logger)
```

#### Пример handler'а

```go
type Handler struct {
    service LoyaltyService  // Инжектируется через интерфейс из contract.go
    logger  Logger          // Инжектируется через интерфейс из contract.go
}

func NewHandler(service LoyaltyService, logger Logger) *Handler {
    return &Handler{
        service: service,
        logger:  logger,
    }
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
    // 1. Парсинг request body
    var req models.CreateLoyaltyCardRequest
    if err := handlers.DecodeJSON(r, &req); err != nil {
        handlers.RespondBadRequest(w, msgInvalidRequestBody)
        return
    }

    // 2. Вызов бизнес-логики
    card, err := h.service.CreateCard(r.Context(), &req)

    // 3. Обработка ошибок через errors.Is()
    if errors.Is(err, loyalty.ErrConfigNotFound) {
        handlers.RespondNotFound(w, msgConfigNotFound)
        return
    }
    if errors.Is(err, loyalty.ErrCardAlreadyExists) {
        handlers.RespondConflict(w, msgCardAlreadyExists)
        return
    }

    // 4. Возврат успешного результата
    handlers.RespondJSON(w, http.StatusCreated, card)
}
```

**Ключевые особенности:**

1. **Константы сообщений** - все user-facing сообщения вынесены в константы в начале файла
2. **Dependency injection** через интерфейсы из `contract.go`
3. **Переиспользуемые helper'ы** из [handlers/utils.go](internal/api/handlers/utils.go)
4. **Структурированное логирование** с контекстом запроса

### Роутинг в main.go

Роутинг настраивается в [cmd/main.go](cmd/main.go) с использованием **Gorilla Mux**:

```go
r := mux.NewRouter()

// 1. Глобальные middleware
if cfg.Metrics.Enabled {
    r.Use(middleware.MetricsMiddleware(metricsCollector, cfg.Metrics.ServiceName))
}

// 2. Публичные endpoints (без аутентификации)
r.Handle("/metrics", promhttp.Handler()).Methods(http.MethodGet)
r.HandleFunc("/health", healthHandler.Handle).Methods(http.MethodGet)

// 3. API v1 prefix
api := r.PathPrefix("/api/v1").Subrouter()

// 4. Публичные API endpoints
api.HandleFunc("/loyalty-cards", getLoyaltyCardHandler.Handle).Methods(http.MethodGet)
api.HandleFunc("/loyalty-cards", createLoyaltyCardHandler.Handle).Methods(http.MethodPost)

// 5. Protected routes (с middleware аутентификации)
protected := api.PathPrefix("").Subrouter()
protected.Use(middleware.Auth)  // Проверяет X-User-ID

// 6. Protected API endpoints
protected.HandleFunc("/companies/{companyId}/loyalty-config",
    configureLoyaltyHandler.Handle).Methods(http.MethodPost)
```

**Архитектурный подход:**
- Middleware применяется иерархически (глобальные → групповые → endpoint-specific)
- Чёткое разделение публичных и защищённых routes
- Version prefix для API (`/api/v1`)
- Только один protected endpoint (настройка конфигурации)

### Обёртывание ошибок между слоями

Проект использует **typed errors** с обёрткой через `fmt.Errorf` и `%w`:

#### 1. Repository Layer

```go
// internal/infra/storage/loyalty_card/errors.go
var (
    ErrCardNotFound = errors.New("repository: loyalty card not found")
    ErrCardAlreadyExists = errors.New("repository: loyalty card already exists")
    ErrBuildQuery = errors.New("repository: failed to build SQL query")
    ErrExecQuery = errors.New("repository: failed to execute SQL query")
)
```

Репозиторий возвращает свои типизированные ошибки:

```go
func (r *Repository) GetByUserAndCompany(ctx context.Context, userID, companyID int64) (*domain.LoyaltyCard, error) {
    // ...
    if err == sql.ErrNoRows {
        return nil, ErrCardNotFound
    }
    return nil, fmt.Errorf("%w: %v", ErrExecQuery, err)
}
```

#### 2. Service Layer

```go
// internal/service/loyalty/errors.go
var (
    ErrCardNotFound = errors.New("loyalty card not found")
    ErrCardAlreadyExists = errors.New("loyalty card already exists")
    ErrConfigNotFound = errors.New("loyalty program not configured for this company")
    ErrConfigDisabled = errors.New("loyalty program is disabled")
    ErrAccessDenied = errors.New("access denied: user is not a manager of this company")
    ErrSellerServiceUnavailable = errors.New("seller service unavailable")
    ErrInternal = errors.New("service: internal error")
)
```

Сервисный слой **перехватывает и преобразует** ошибки репозитория:

```go
func (s *Service) GetCard(ctx context.Context, userID, companyID int64) (*models.LoyaltyCardResponse, error) {
    card, err := s.cardRepo.GetByUserAndCompany(ctx, userID, companyID)
    if err != nil {
        if errors.Is(err, cardRepo.ErrCardNotFound) {
            return nil, ErrCardNotFound
        }
        return nil, fmt.Errorf("%w: GetCard - repository error: %v", ErrInternal, err)
    }

    // Получаем конфигурацию для добавления discount_percentage
    config, err := s.configRepo.GetByCompanyID(ctx, companyID)
    if err != nil {
        return nil, fmt.Errorf("%w: GetCard - failed to get config: %v", ErrInternal, err)
    }

    return models.FromDomainLoyaltyCard(card, config.DiscountPercentage), nil
}
```

#### 3. Handler Layer

Handler **проверяет типизированные ошибки** и преобразует в HTTP responses:

```go
card, err := h.service.CreateCard(r.Context(), &req)
if err != nil {
    if errors.Is(err, loyalty.ErrConfigNotFound) {
        h.logger.Warn("Config not found: company_id=%d", req.CompanyID)
        handlers.RespondNotFound(w, msgConfigNotFound)  // 404
        return
    }
    if errors.Is(err, loyalty.ErrCardAlreadyExists) {
        handlers.RespondConflict(w, msgCardAlreadyExists)  // 409
        return
    }
    h.logger.Error("Failed to create loyalty card: %v", err)
    handlers.RespondInternalError(w)  // 500
    return
}
```

**Преимущества подхода:**
- Каждый слой определяет свои ошибки
- Ошибки нижних слоёв проверяются через `errors.Is()` и преобразуются
- Полный error trace через `fmt.Errorf("%w", err)`
- Handler не зависит от ошибок repository напрямую
- Логирование с контекстом на каждом уровне

### Вынос констант

Константы выносятся на разные уровни в зависимости от области применения:

#### 1. Handler-level константы

User-facing сообщения и настройки конкретного handler'а:

```go
// internal/api/handlers/create_loyalty_card/handler.go
const (
    msgInvalidRequestBody = "invalid request body"
    msgConfigNotFound = "loyalty program not configured for this company"
    msgCardAlreadyExists = "loyalty card already exists"
)
```

#### 2. Domain-level константы

Enum значения для типов и статусов:

```go
// internal/domain/enums.go
const (
    CardTypeFixedDiscount = "fixed_discount"
    CardTypeProgressiveDiscount = "progressive_discount"
    CardTypePointsBased = "points_based"

    CardStatusActive = "active"
    CardStatusSuspended = "suspended"
    CardStatusExpired = "expired"
)
```

### Contract файлы (contract.go)

Каждый handler определяет **минимальные интерфейсы** для своих зависимостей:

```go
// internal/api/handlers/create_loyalty_card/contract.go
type LoyaltyService interface {
    CreateCard(ctx context.Context,
               req *models.CreateLoyaltyCardRequest) (*models.LoyaltyCardResponse, error)
}

type Logger interface {
    Info(format string, v ...interface{})
    Warn(format string, v ...interface{})
    Error(format string, v ...interface{})
}
```

**Преимущества:**
- **Инверсия зависимостей** - handler зависит от интерфейса, а не от конкретной реализации
- **Лёгкое тестирование** - можно легко создавать моки
- **Минимальные зависимости** - handler видит только нужные ему методы
- **Явные границы слоёв** - чётко видно, что handler требует от service layer

### Поток данных в приложении

```
HTTP Request
    ↓
Middleware (Auth для protected, Metrics для всех)
    ↓
Handler (парсинг query/body, валидация)
    ↓
Service (бизнес-логика, интеграции с SellerService)
    ↓
Repository (работа с БД через dbmetrics.DB)
    ↓
PostgreSQL (loyalty_cards, loyalty_configs)
    ↓
Domain models → Service models → API response
    ↓
HTTP Response
```

Каждый слой использует свои модели данных с преобразованием на границах слоёв через методы `ToDomain()` и `FromDomain()`.

## Расширение функциональности

### Добавление прогрессивной скидки

1. Создать JSONB структуру в `progressive_config`:
```json
{
  "tiers": [
    {"min_visits": 0, "discount": 5},
    {"min_visits": 5, "discount": 10},
    {"min_visits": 10, "discount": 15}
  ]
}
```

2. Добавить интеграцию с BookingService для получения количества посещений
3. Реализовать логику расчёта скидки в `internal/service/loyalty/progressive.go`
4. Обновить API endpoint для настройки `progressive_config`

### Добавление накопительной системы баллов

1. Создать таблицу `loyalty_transactions` для хранения операций с баллами
2. Создать JSONB структуру в `points_config`:
```json
{
  "points_per_ruble": 1,
  "redemption_rate": 0.01
}
```

3. Реализовать endpoints:
   - `POST /api/v1/loyalty-cards/{id}/transactions` - начисление/списание баллов
   - `GET /api/v1/loyalty-cards/{id}/balance` - получение баланса баллов

4. Добавить репозиторий и сервис для работы с транзакциями

## Зависимости

- **Gorilla Mux** - HTTP роутинг
- **lib/pq** - PostgreSQL драйвер
- **Prometheus client** - метрики
- **Squirrel** - SQL query builder
- **TOML parser** - конфигурация

## Лицензия

MIT
