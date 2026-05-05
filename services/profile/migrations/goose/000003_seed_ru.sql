-- +goose Up
UPDATE sport_type
SET name = CASE sport_type_id
  WHEN 3001 THEN 'Бег'
  WHEN 3002 THEN 'Плавание'
  WHEN 3003 THEN 'Йога'
  WHEN 3004 THEN 'Велоспорт'
  ELSE name
END,
updated_at = NOW()
WHERE sport_type_id IN (3001, 3002, 3003, 3004);

INSERT INTO sport_type (sport_type_id, name)
VALUES
  (3005, 'Бокс'),
  (3006, 'Силовые тренировки'),
  (3007, 'Функциональный тренинг'),
  (3008, 'Растяжка')
ON CONFLICT (sport_type_id) DO UPDATE
SET name = EXCLUDED.name,
    updated_at = NOW();

INSERT INTO profile (user_id, username, first_name, last_name, bio, avatar_url, is_trainer)
VALUES
  (1001, 'coach_anna', 'Анна', 'Петрова', 'Тренер по бегу и ОФП. Помогаю готовиться к забегам 10K, полумарафону и первому марафону без травм.', NULL, true),
  (1002, 'runner_ivan', 'Иван', 'Сидоров', 'Любитель бега. Ищу тренера и собираю программу подготовки к первому полумарафону.', NULL, false),
  (1003, 'swim_mike', 'Михаил', 'Волков', 'Тренер по плаванию. Работаю с техникой дыхания, постановкой гребка и выносливостью.', NULL, true),
  (1004, 'yoga_elena', 'Елена', 'Смирнова', 'Инструктор по йоге и мягкой растяжке. Подходит для новичков, восстановления и работы со спиной.', NULL, true),
  (1005, 'box_sergey', 'Сергей', 'Орлов', 'Тренер по боксу. Ставлю технику ударов, работу ног, защиту и безопасные спарринги.', NULL, true),
  (1006, 'fit_olga', 'Ольга', 'Кузнецова', 'Тренер по силовым и функциональным тренировкам. Делаю программы для дома и зала.', NULL, true),
  (1007, 'cycle_dima', 'Дмитрий', 'Морозов', 'Тренер по велоспорту. Помогаю развивать выносливость, каденс и готовиться к длинным заездам.', NULL, true),
  (1008, 'client_maria', 'Мария', 'Федорова', 'Хочу вернуться к регулярным тренировкам после перерыва.', NULL, false),
  (1009, 'client_pavel', 'Павел', 'Никитин', 'Начинаю заниматься спортом и выбираю тренера для системной работы.', NULL, false),
  (1010, 'admin_sporttech', 'Админ', 'SPORTtech', 'Технический аккаунт администратора.', NULL, false)
ON CONFLICT (user_id) DO UPDATE
SET username = EXCLUDED.username,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    bio = EXCLUDED.bio,
    avatar_url = EXCLUDED.avatar_url,
    is_trainer = EXCLUDED.is_trainer,
    updated_at = NOW();

INSERT INTO trainer_profile (user_id, education_degree, career_since_date)
VALUES
  (1001, 'Магистр физической культуры', DATE '2018-09-01'),
  (1003, 'КМС по плаванию', DATE '2017-05-15'),
  (1004, 'Сертифицированный инструктор хатха-йоги', DATE '2019-03-10'),
  (1005, 'МС по боксу', DATE '2016-02-01'),
  (1006, 'Специалист по физической подготовке', DATE '2020-01-20'),
  (1007, 'Тренер по циклическим видам спорта', DATE '2018-04-12')
ON CONFLICT (user_id) DO UPDATE
SET education_degree = EXCLUDED.education_degree,
    career_since_date = EXCLUDED.career_since_date,
    updated_at = NOW();

INSERT INTO trainer_sport (user_id, sport_type_id, experience_years, sports_rank)
VALUES
  (1001, 3001, 7, 'КМС'),
  (1001, 3004, 4, NULL),
  (1003, 3002, 9, 'КМС'),
  (1003, 3006, 5, NULL),
  (1004, 3003, 6, NULL),
  (1004, 3008, 5, NULL),
  (1005, 3005, 10, 'МС'),
  (1005, 3007, 6, NULL),
  (1006, 3006, 8, NULL),
  (1006, 3007, 7, NULL),
  (1007, 3004, 8, 'КМС'),
  (1007, 3001, 5, NULL)
ON CONFLICT (user_id, sport_type_id) DO UPDATE
SET experience_years = EXCLUDED.experience_years,
    sports_rank = EXCLUDED.sports_rank,
    updated_at = NOW();

SELECT setval(
  pg_get_serial_sequence('sport_type', 'sport_type_id'),
  (SELECT GREATEST(COALESCE(MAX(sport_type_id), 1), 3008) FROM sport_type),
  true
);

-- +goose Down
DELETE FROM trainer_sport
WHERE user_id BETWEEN 1004 AND 1007
   OR (user_id = 1003 AND sport_type_id = 3006)
   OR (user_id = 1007 AND sport_type_id = 3001);

DELETE FROM trainer_profile
WHERE user_id BETWEEN 1004 AND 1007;

DELETE FROM profile
WHERE user_id BETWEEN 1004 AND 1010;

DELETE FROM sport_type
WHERE sport_type_id BETWEEN 3005 AND 3008;

UPDATE sport_type
SET name = CASE sport_type_id
  WHEN 3001 THEN 'Running'
  WHEN 3002 THEN 'Swimming'
  WHEN 3003 THEN 'Yoga'
  WHEN 3004 THEN 'Cycling'
  ELSE name
END,
updated_at = NOW()
WHERE sport_type_id IN (3001, 3002, 3003, 3004);

UPDATE profile
SET first_name = CASE user_id
    WHEN 1001 THEN 'Anna'
    WHEN 1002 THEN 'Ivan'
    WHEN 1003 THEN 'Mikhail'
    ELSE first_name
  END,
  last_name = CASE user_id
    WHEN 1001 THEN 'Petrova'
    WHEN 1002 THEN 'Sidorov'
    WHEN 1003 THEN 'Volkov'
    ELSE last_name
  END,
  bio = CASE user_id
    WHEN 1001 THEN 'Тренер по бегу и ОФП. Помогаю готовиться к забегам 10K и полумарафону.'
    WHEN 1002 THEN 'Любитель бега. Ищу тренера и собираю программу подготовки к первому полумарафону.'
    WHEN 1003 THEN 'Тренер по плаванию. Работаю с техникой дыхания и выносливостью.'
    ELSE bio
  END,
  updated_at = NOW()
WHERE user_id IN (1001, 1002, 1003);
