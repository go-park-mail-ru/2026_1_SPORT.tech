INSERT INTO auth_user (user_id, email, username, password_hash, role, status)
VALUES
  (1004, 'elena.yoga@sporttech.local', 'yoga_elena', '$2a$10$wyfCWjNi8FUDZt10JYydtON/VIMO2IsbMaQGzkj5QsT47DIpHeW6W', 'trainer', 'active'),
  (1005, 'sergey.box@sporttech.local', 'box_sergey', '$2a$10$wyfCWjNi8FUDZt10JYydtON/VIMO2IsbMaQGzkj5QsT47DIpHeW6W', 'trainer', 'active'),
  (1006, 'olga.fit@sporttech.local', 'fit_olga', '$2a$10$wyfCWjNi8FUDZt10JYydtON/VIMO2IsbMaQGzkj5QsT47DIpHeW6W', 'trainer', 'active'),
  (1007, 'dima.cycle@sporttech.local', 'cycle_dima', '$2a$10$wyfCWjNi8FUDZt10JYydtON/VIMO2IsbMaQGzkj5QsT47DIpHeW6W', 'trainer', 'active'),
  (1008, 'maria.client@sporttech.local', 'client_maria', '$2a$10$wyfCWjNi8FUDZt10JYydtON/VIMO2IsbMaQGzkj5QsT47DIpHeW6W', 'client', 'active'),
  (1009, 'pavel.client@sporttech.local', 'client_pavel', '$2a$10$wyfCWjNi8FUDZt10JYydtON/VIMO2IsbMaQGzkj5QsT47DIpHeW6W', 'client', 'active'),
  (1010, 'admin@sporttech.local', 'admin_sporttech', '$2a$10$wyfCWjNi8FUDZt10JYydtON/VIMO2IsbMaQGzkj5QsT47DIpHeW6W', 'admin', 'active')
ON CONFLICT (user_id) DO UPDATE
SET email = EXCLUDED.email,
    username = EXCLUDED.username,
    password_hash = EXCLUDED.password_hash,
    role = EXCLUDED.role,
    status = EXCLUDED.status,
    updated_at = NOW();

SELECT setval(
  pg_get_serial_sequence('auth_user', 'user_id'),
  (SELECT GREATEST(COALESCE(MAX(user_id), 1), 1010) FROM auth_user),
  true
);
