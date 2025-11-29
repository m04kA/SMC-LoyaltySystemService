-- Таблица конфигураций программ лояльности компаний
CREATE TABLE loyalty_configs (
    id BIGSERIAL PRIMARY KEY,
    company_id BIGINT NOT NULL UNIQUE,
    card_type VARCHAR(50) NOT NULL DEFAULT 'fixed_discount',
    is_enabled BOOLEAN NOT NULL DEFAULT true,
    discount_percentage DECIMAL(5,2) CHECK (discount_percentage >= 0 AND discount_percentage <= 100),
    progressive_config JSONB,
    points_config JSONB,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT loyalty_configs_valid_card_type CHECK (card_type IN ('fixed_discount', 'progressive_discount', 'points_based'))
);

-- Индексы для loyalty_configs
CREATE INDEX idx_loyalty_configs_company_id ON loyalty_configs(company_id);
CREATE INDEX idx_loyalty_configs_enabled ON loyalty_configs(is_enabled);

-- Таблица карт лояльности клиентов
CREATE TABLE loyalty_cards (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    company_id BIGINT NOT NULL,
    card_type VARCHAR(50) NOT NULL DEFAULT 'fixed_discount',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    discount_percentage DECIMAL(5,2) NOT NULL CHECK (discount_percentage >= 0 AND discount_percentage <= 100),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT loyalty_cards_unique_user_company UNIQUE (user_id, company_id)
);

-- Индексы для loyalty_cards
CREATE INDEX idx_loyalty_cards_user_id ON loyalty_cards(user_id);
CREATE INDEX idx_loyalty_cards_company_id ON loyalty_cards(company_id);
CREATE INDEX idx_loyalty_cards_user_company ON loyalty_cards(user_id, company_id);
CREATE INDEX idx_loyalty_cards_type ON loyalty_cards(card_type);
CREATE INDEX idx_loyalty_cards_status ON loyalty_cards(status);

-- Функция для автоматического обновления updated_at
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Триггеры для автоматического обновления updated_at
CREATE TRIGGER update_loyalty_configs_updated_at BEFORE UPDATE ON loyalty_configs
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_loyalty_cards_updated_at BEFORE UPDATE ON loyalty_cards
    FOR EACH ROW EXECUTE FUNCTION update_updated_at_column();
