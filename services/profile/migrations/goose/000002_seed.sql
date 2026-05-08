-- +goose Up
INSERT INTO sport_type (sport_type_id, name)
VALUES
  (3001, 'Running'),
  (3002, 'Swimming'),
  (3003, 'Yoga'),
  (3004, 'Cycling');

INSERT INTO profile (user_id, username, first_name, last_name, bio, avatar_url, is_trainer)
VALUES
  (1001, 'coach_anna', 'Anna', 'Petrova', 'Тренер по бегу и ОФП. Помогаю готовиться к забегам 10K и полумарафону.', NULL, true),
  (1002, 'runner_ivan', 'Ivan', 'Sidorov', 'Любитель бега. Ищу тренера и собираю программу подготовки к первому полумарафону.', NULL, false),
  (1003, 'swim_mike', 'Mikhail', 'Volkov', 'Тренер по плаванию. Работаю с техникой дыхания и выносливостью.', NULL, true);

INSERT INTO trainer_profile (user_id, education_degree, career_since_date)
VALUES
  (1001, 'Магистр физической культуры', DATE '2018-09-01'),
  (1003, 'КМС по плаванию', DATE '2017-05-15');

INSERT INTO trainer_sport (user_id, sport_type_id, experience_years, sports_rank)
VALUES
  (1001, 3001, 7, 'КМС'),
  (1001, 3004, 4, NULL),
  (1003, 3002, 9, 'КМС');

SELECT setval(
  pg_get_serial_sequence('sport_type', 'sport_type_id'),
  (SELECT GREATEST(COALESCE(MAX(sport_type_id), 1), 3004) FROM sport_type),
  true
);

-- +goose Down
DELETE FROM trainer_sport
WHERE user_id IN (1001, 1003);

DELETE FROM trainer_profile
WHERE user_id IN (1001, 1003);

DELETE FROM profile
WHERE user_id IN (1001, 1002, 1003);

DELETE FROM sport_type
WHERE sport_type_id IN (3001, 3002, 3003, 3004);
