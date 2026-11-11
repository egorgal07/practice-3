-- Записи сайта. Добавляйте и меняйте строки в INSERT ниже.
CREATE TABLE IF NOT EXISTS items (
  id SERIAL PRIMARY KEY,
  name TEXT NOT NULL
);

-- Очищаем таблицу, чтобы после применения в ней было ровно то, что написано ниже
TRUNCATE TABLE items RESTART IDENTITY;

INSERT INTO items (name) VALUES
  ('first item'),
  ('second item');