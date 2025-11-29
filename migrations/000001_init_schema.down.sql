-- Удаляем триггеры
DROP TRIGGER IF EXISTS update_loyalty_cards_updated_at ON loyalty_cards;
DROP TRIGGER IF EXISTS update_loyalty_configs_updated_at ON loyalty_configs;

-- Удаляем функцию
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Удаляем таблицы
DROP TABLE IF EXISTS loyalty_cards;
DROP TABLE IF EXISTS loyalty_configs;
