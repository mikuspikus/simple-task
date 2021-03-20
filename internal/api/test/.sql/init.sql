CREATE TYPE Status AS ENUM ('В пути', 'На складе', 'Продан', 'Снят с продажи');

-- 1. Уникальный идентификатор (любой тип, общение с БД не является критерием чего-либо, можно сделать и in-memory хранилище на время жизни сервиса)
-- 2. Бренд автомобиля (текст)
-- 3. Модель автомобиля (текст)
-- 4. Цена автомобиля (целое, не может быть меньше 0)
-- 5. Статус автомобиля (В пути, На складе, Продан, Снят с продажи)
-- 6. Пробег (целое)
CREATE TABLE CarModel (
    id SERIAL PRIMARY KEY,
    brand VARCHAR(256) NOT NULL,
    model VARCHAR(256) NOT NULL,
    price bigint CHECK (price > 0),
    status Status,
    mileage bigint CHECK (mileage > 0)
);