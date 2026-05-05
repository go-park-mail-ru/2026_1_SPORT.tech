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
