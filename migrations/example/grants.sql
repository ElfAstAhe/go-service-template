-- Manual usage

-- 1. Позволяет создавать новые объекты (таблицы, индексы, последовательности)
GRANT CREATE ON SCHEMA example_service TO test;

-- 2. Позволяет видеть схему
GRANT USAGE ON SCHEMA example_service TO test;

-- 3. Если таблицы уже созданы другим юзером (например, postgres),
-- нужно сделать test их владельцем или дать полные права:
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA example_service TO test;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA example_service TO test;

-- 4. Чтобы будущие объекты тоже были под контролем:
ALTER DEFAULT PRIVILEGES IN SCHEMA example_service
    GRANT ALL PRIVILEGES ON TABLES TO test;
ALTER DEFAULT PRIVILEGES IN SCHEMA example_service
    GRANT ALL PRIVILEGES ON SEQUENCES TO test;