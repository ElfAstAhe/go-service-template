-- Manual usage

-- 1. Разрешаем пользователю видеть саму схему
GRANT USAGE ON SCHEMA example_service TO test;

-- 2. Даем права на данные (SELECT, INSERT, UPDATE, DELETE)
GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA example_service TO test;

-- 3. ОБЯЗАТЕЛЬНО: Даем права на последовательности (нужно для SERIAL и IDENTITY колонок)
GRANT USAGE, SELECT ON ALL SEQUENCES IN SCHEMA example_service TO test;

-- 4. На будующее
ALTER DEFAULT PRIVILEGES IN SCHEMA example_service
    GRANT SELECT, INSERT, UPDATE, DELETE ON TABLES TO test;

ALTER DEFAULT PRIVILEGES IN SCHEMA example_service
    GRANT USAGE, SELECT ON SEQUENCES TO test;
