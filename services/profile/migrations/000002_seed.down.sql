DELETE FROM trainer_sport
WHERE user_id IN (1001, 1003);

DELETE FROM trainer_profile
WHERE user_id IN (1001, 1003);

DELETE FROM profile
WHERE user_id IN (1001, 1002, 1003);

DELETE FROM sport_type
WHERE sport_type_id IN (3001, 3002, 3003, 3004);
