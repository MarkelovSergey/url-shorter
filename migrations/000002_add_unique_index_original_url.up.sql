-- Создаем уникальный индекс для original_url
CREATE UNIQUE INDEX IF NOT EXISTS idx_unique_original_url ON urls(original_url);
