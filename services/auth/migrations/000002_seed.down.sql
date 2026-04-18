DELETE FROM auth_session
WHERE user_id IN (1001, 1002, 1003);

DELETE FROM auth_user
WHERE user_id IN (1001, 1002, 1003);
