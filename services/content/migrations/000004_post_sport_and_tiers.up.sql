ALTER TABLE content_post
ADD COLUMN sport_type_id BIGINT;

CREATE TABLE content_subscription_tier (
  trainer_user_id BIGINT NOT NULL,
  tier_id INTEGER NOT NULL CHECK (tier_id >= 1),
  name TEXT NOT NULL CHECK (char_length(name) BETWEEN 1 AND 80),
  price INTEGER NOT NULL CHECK (price >= 0),
  description TEXT CHECK (description IS NULL OR char_length(description) <= 500),
  created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

  PRIMARY KEY (trainer_user_id, tier_id)
);

INSERT INTO content_subscription_tier (trainer_user_id, tier_id, name, price, description)
VALUES
  (1001, 1, 'Базовый', 500, 'Доступ к открытым тренировочным материалам и базовым планам.'),
  (1001, 2, 'Продвинутый', 1500, 'Закрытые темповые тренировки, недельные планы и разбор нагрузки.'),
  (1003, 1, 'Базовый', 600, 'Материалы по технике плавания и дыханию.'),
  (1004, 1, 'Базовый', 400, 'Короткие комплексы йоги для ежедневной практики.'),
  (1004, 2, 'Глубокая практика', 1200, 'Закрытые комплексы растяжки и восстановления.'),
  (1005, 1, 'Базовый', 700, 'Базовая техника бокса и работа ног.'),
  (1005, 2, 'Спарринг', 1400, 'Расширенные упражнения для защиты и контратак.'),
  (1005, 3, 'Премиум', 2200, 'Подробные закрытые разборы защитных действий.'),
  (1006, 1, 'Базовый', 500, 'Домашние силовые тренировки без инвентаря.'),
  (1006, 2, 'Прогрессия', 1300, 'Закрытые программы прогрессии и контроля нагрузки.'),
  (1007, 1, 'Базовый', 600, 'Материалы по технике велотренировок.'),
  (1007, 2, 'План подготовки', 1500, 'Закрытые месячные планы велоподготовки.')
ON CONFLICT (trainer_user_id, tier_id) DO UPDATE
SET name = EXCLUDED.name,
    price = EXCLUDED.price,
    description = EXCLUDED.description,
    updated_at = now();

INSERT INTO content_subscription_tier (trainer_user_id, tier_id, name, price, description)
SELECT DISTINCT
  author_user_id,
  required_subscription_level,
  'Уровень ' || required_subscription_level,
  0,
  'Автоматически созданный уровень для существующих закрытых постов.'
FROM content_post
WHERE required_subscription_level IS NOT NULL
ON CONFLICT (trainer_user_id, tier_id) DO NOTHING;

UPDATE content_post
SET sport_type_id = CASE post_id
  WHEN 2001 THEN 3001
  WHEN 2002 THEN 3001
  WHEN 2003 THEN 3003
  WHEN 2004 THEN 3005
  WHEN 2005 THEN 3005
  WHEN 2006 THEN 3006
  WHEN 2007 THEN 3006
  WHEN 2008 THEN 3007
  WHEN 2009 THEN 3007
  WHEN 2010 THEN 3008
  WHEN 2011 THEN 3008
  ELSE sport_type_id
END
WHERE post_id BETWEEN 2001 AND 2011;

ALTER TABLE content_post
ADD CONSTRAINT content_post_required_subscription_tier_fkey
FOREIGN KEY (author_user_id, required_subscription_level)
REFERENCES content_subscription_tier(trainer_user_id, tier_id);

CREATE INDEX content_post_sport_type_id_idx ON content_post(sport_type_id);
