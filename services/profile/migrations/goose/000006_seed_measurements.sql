-- +goose Up

INSERT INTO measurement
  (measurement_id, user_id, measured_at,
   weight_kg, body_fat_pct, chest_cm, waist_cm, hips_cm, notes)
VALUES
  (2701, 1002, '2026-03-01', 78.50, 18.5, 96, 84, 96,
   'Начало тренировок. Цель — подготовка к полумарафону.'),
  (2702, 1002, '2026-04-01', 77.20, 17.8, 96, 82, 95,
   'Месяц по плану Анны. Бегаю три раза в неделю, самочувствие хорошее.'),
  (2703, 1002, '2026-05-01', 75.80, 17.0, 95, 80, 94,
   'Вес снижается, выносливость растёт. Длинная уже 16 км без остановок.'),

  (2704, 1008, '2026-03-15', 62.00, 24.0, 88, 72, 98,
   'Первый замер. Начинаю заниматься после долгого перерыва.'),
  (2705, 1008, '2026-04-15', 61.30, 23.2, 88, 71, 97,
   'Йога и домашние силовые. Спина болит меньше, стала спокойнее сплю.'),
  (2706, 1008, '2026-05-15', 60.50, 22.5, 87, 70, 96,
   'Продолжаю. Чувствую себя лучше, чем за последние два года.'),

  (2707, 1009, '2026-04-01', 85.00, 22.0, 102, 92, 102,
   'Начало. Хочу сбросить 5 кг и укрепить корпус.'),
  (2708, 1009, '2026-05-01', 83.70, 21.3, 101, 90, 101,
   'Месяц занятий с Сергеем. Минус 1.3 кг, плечи стали плотнее.')
ON CONFLICT (measurement_id) DO UPDATE
SET user_id      = EXCLUDED.user_id,
    measured_at  = EXCLUDED.measured_at,
    weight_kg    = EXCLUDED.weight_kg,
    body_fat_pct = EXCLUDED.body_fat_pct,
    chest_cm     = EXCLUDED.chest_cm,
    waist_cm     = EXCLUDED.waist_cm,
    hips_cm      = EXCLUDED.hips_cm,
    notes        = EXCLUDED.notes,
    updated_at   = now();

SELECT setval(
  pg_get_serial_sequence('measurement', 'measurement_id'),
  (SELECT GREATEST(COALESCE(MAX(measurement_id), 1), 2708) FROM measurement),
  true
);

INSERT INTO measurement_sharing (client_user_id, trainer_user_id)
VALUES
  (1002, 1001), 
  (1008, 1004) 
ON CONFLICT (client_user_id, trainer_user_id) DO NOTHING;

-- +goose Down
DELETE FROM measurement_sharing
WHERE (client_user_id = 1002 AND trainer_user_id = 1001)
   OR (client_user_id = 1008 AND trainer_user_id = 1004);

DELETE FROM measurement
WHERE measurement_id BETWEEN 2701 AND 2708;
