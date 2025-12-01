# План реализации SMC-LoyaltySystemService

## 1. Обзор

**Цель:** Создание микросервиса для управления программами лояльности компаний (автомоек).

**Текущий scope:** Карты с фиксированной скидкой + архитектура для расширения (прогрессивные скидки, баллы).

**Технические требования:**
- Порт: `8084`
- БД: PostgreSQL на порту `5439`
- Clean Architecture
- Интеграции: SellerService (8081), UserService (8080)

---

## 2. Ключевые архитектурные решения

### 2.1 Типы карт: Enum + JSONB

- Поле `card_type` (enum) в `loyalty_cards`
- Текущая реализация: только `discount_percentage`
- JSONB поля для будущих типов (`progressive_config`, `points_config`)

### 2.2 Создание карт: Явное создание через POST endpoint

- Карта создаётся через `POST /api/v1/loyalty-cards` (клиент явно запрашивает создание)
- Только если конфигурация существует и включена
- Проверка, что карта ещё не существует

### 2.3 Данные для QR-кода

Возвращать: `card_id`, `user_id`, `company_id`, `discount_percentage`, `created_at`, `updated_at`

**Достаточно для MVP.** Будущее: HMAC подпись для защиты от подделки.

### 2.4 Валидация прав: Через SellerService (для managers)

- Вызов API SellerService через СУЩЕСТВУЮЩИЙ клиент в `internal/integrations/sellerservice`
- Проверка `manager_ids` в ответе от `/api/v1/companies/{id}`
- **ВАЖНО:** Используем только `X-User-ID` header (без `X-User-Role`)

---

## 3. Схема БД

### loyalty_cards
```sql
CREATE TABLE loyalty_cards (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    company_id BIGINT NOT NULL,
    card_type VARCHAR(50) NOT NULL DEFAULT 'fixed_discount',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT loyalty_cards_unique_user_company UNIQUE (user_id, company_id)
);

-- Индексы
CREATE INDEX idx_loyalty_cards_user_id ON loyalty_cards(user_id);
CREATE INDEX idx_loyalty_cards_company_id ON loyalty_cards(company_id);
CREATE INDEX idx_loyalty_cards_user_company ON loyalty_cards(user_id, company_id);
CREATE INDEX idx_loyalty_cards_type ON loyalty_cards(card_type);
CREATE INDEX idx_loyalty_cards_status ON loyalty_cards(status);
```

### loyalty_configs
```sql
CREATE TABLE loyalty_configs (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL UNIQUE,
    card_type VARCHAR(50) NOT NULL DEFAULT 'fixed_discount',
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    discount_percentage DECIMAL(5,2) CHECK (discount_percentage >= 0 AND discount_percentage <= 100),
    progressive_config JSONB,
    points_config JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    CONSTRAINT loyalty_configs_valid_card_type CHECK (card_type IN ('fixed_discount', 'progressive_discount', 'points_based'))
);

CREATE INDEX idx_loyalty_configs_company_id ON loyalty_configs(company_id);
CREATE INDEX idx_loyalty_configs_enabled ON loyalty_configs(is_enabled);
```

---

## 4. API Endpoints

### GET /api/v1/loyalty-cards (Public)

**Query:** `?userId={id}&companyId={id}`

**Response (200):**
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

**Ошибки:** 400, 404 (карта не найдена)

### POST /api/v1/loyalty-cards (Public)

**Request:**
```json
{
  "user_id": 987654321,
  "company_id": 1
}
```

**Response (201):**
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

**Ошибки:**
- 400 - Невалидные данные
- 404 - Конфигурация не найдена
- 409 - Карта уже существует

### POST /api/v1/companies/{companyId}/loyalty-config (Protected)

**Headers:** `X-User-ID`

**Request:**
```json
{
  "discount_percentage": 15.0
}
```

**Response (200):**
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

**Ошибки:** 400, 401, 403, 503 (SellerService недоступен)

---

## 5. Структура кода

### Domain (internal/domain/)
- `enums.go` - CardType, CardStatus
- `loyalty_card.go` - LoyaltyCard + Input модели
- `loyalty_config.go` - LoyaltyConfig + Input модели

### Repository (internal/infra/storage/)
- `loyalty_card/` - GetByUserAndCompany, Create
- `loyalty_config/` - GetByCompanyID, Create, Update

### Service (internal/service/)
- `loyalty/service.go`:
  - `GetCard()` - получение существующей карты
  - `CreateCard()` - создание новой карты
  - `ConfigureLoyalty()` - с проверкой прав через SellerService

### Integrations (internal/integrations/)
- `sellerservice/` - **УЖЕ СУЩЕСТВУЕТ** - использовать готовый клиент

### Handlers (internal/api/handlers/)
- `get_loyalty_card/` - GET /api/v1/loyalty-cards
- `create_loyalty_card/` - POST /api/v1/loyalty-cards
- `configure_loyalty/` - POST /api/v1/companies/{companyId}/loyalty-config

---

## 6. Порядок реализации (ОБНОВЛЁННЫЙ)

### Этап 1: Документация и спецификация
1. **OpenAPI schema** (schemas/schema.yaml) - полная спецификация API
2. **CLAUDE.md** - обновление документации проекта

### Этап 2: Подготовка
3. **Удалить шаблонные файлы:** company, address, working_hours (domain, service, repo, handlers, миграции)

### Этап 3: Реализация
4. **Миграции** (000001_create_loyalty_tables.up/down.sql) - ПЕРЕЗАПИСАТЬ миграцию 000001
5. **Domain модели** (enums.go, loyalty_card.go, loyalty_config.go)
6. **Repositories** (loyalty_card, loyalty_config)
7. **Service layer** (loyalty/service.go) - использовать СУЩЕСТВУЮЩИЙ SellerService клиент
8. **Handlers** (get_loyalty_card, create_loyalty_card, configure_loyalty)
9. **Config updates** (если нужно)
10. **Main wiring** (cmd/main.go)

### Важно:
- **НЕ создавать** новый external/seller_client.go - использовать существующий!
- **Убрать X-User-Role** из всех handlers - только X-User-ID
- **Удалить** все примеры из шаблона (companies, addresses, etc.)

---

## 7. Критические файлы

### Для удаления (шаблонные примеры):
- `internal/domain/company.go`, `address.go`, `working_hours.go`
- `internal/service/companies/` (весь пакет)
- `internal/infra/storage/company/` (весь пакет)
- `internal/api/handlers/create_company/`, `get_company/`
- `migrations/000001_init_schema.up.sql` (заменить на loyalty tables)

### Для создания:
1. `schemas/schema.yaml` - OpenAPI спецификация ✅ ПЕРВЫЙ ПРИОРИТЕТ
2. `CLAUDE.md` - Документация ✅ ПЕРВЫЙ ПРИОРИТЕТ
3. `migrations/000001_create_loyalty_tables.up.sql` - БД schema
4. `internal/domain/enums.go` - CardType, CardStatus
5. `internal/domain/loyalty_card.go` - LoyaltyCard модель
6. `internal/domain/loyalty_config.go` - LoyaltyConfig модель
7. `internal/infra/storage/loyalty_card/repository.go` - CRUD для карт
8. `internal/infra/storage/loyalty_config/repository.go` - CRUD для конфигов
9. `internal/service/loyalty/service.go` - Бизнес-логика (GetCard, CreateCard, ConfigureLoyalty)
10. `internal/api/handlers/get_loyalty_card/handler.go` - GET endpoint
11. `internal/api/handlers/create_loyalty_card/handler.go` - POST endpoint (создание карты)
12. `internal/api/handlers/configure_loyalty/handler.go` - POST endpoint (конфигурация)
13. `cmd/main.go` - Обновить wiring

---

## 8. Расширяемость

**Прогрессивная скидка:**
```json
{
  "progressive_config": {
    "tiers": [
      {"min_visits": 0, "discount": 5},
      {"min_visits": 5, "discount": 10}
    ]
  }
}
```

Требует: интеграция с BookingService для подсчёта посещений.

**Накопительная система:**
```json
{
  "points_config": {
    "points_per_ruble": 1,
    "redemption_rate": 0.01
  }
}
```

Требует: новая таблица `loyalty_transactions`.

---

## 9. Ответы на вопросы

### QR-код
Текущих полей достаточно. Дополнительные данные не требуются.

### Шаблон
Весь код из шаблона удалить - он нужен только как референс по стилю кода.

### Авторизация
**Убрать X-User-Role полностью.** Использовать только X-User-ID.
Проверка прав через SellerService (manager_ids).

### Название модуля
`github.com/m04kA/SMC-LoyaltySystemService` - **правильное** (SMC, не smc).

### SellerService клиент
**УЖЕ РЕАЛИЗОВАН** в `internal/integrations/sellerservice/` - использовать его!
Schema: `schemas/clients/smc-sellerservice.yaml`

---

## План готов к реализации ✅

**Начинаем с:**
1. OpenAPI schema
2. CLAUDE.md
3. Код реализации
