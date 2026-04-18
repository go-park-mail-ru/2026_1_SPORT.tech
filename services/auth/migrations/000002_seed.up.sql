INSERT INTO auth_user (user_id, email, username, password_hash, role, status)
VALUES
  (1001, 'anna.coach@sporttech.local', 'coach_anna', '$2a$10$wyfCWjNi8FUDZt10JYydtON/VIMO2IsbMaQGzkj5QsT47DIpHeW6W', 'trainer', 'active'),
  (1002, 'ivan.runner@sporttech.local', 'runner_ivan', '$2a$10$wyfCWjNi8FUDZt10JYydtON/VIMO2IsbMaQGzkj5QsT47DIpHeW6W', 'client', 'active'),
  (1003, 'mike.swim@sporttech.local', 'swim_mike', '$2a$10$wyfCWjNi8FUDZt10JYydtON/VIMO2IsbMaQGzkj5QsT47DIpHeW6W', 'trainer', 'active');

SELECT setval(
  pg_get_serial_sequence('auth_user', 'user_id'),
  (SELECT GREATEST(COALESCE(MAX(user_id), 1), 1003) FROM auth_user),
  true
);
