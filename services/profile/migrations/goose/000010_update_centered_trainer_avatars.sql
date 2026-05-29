-- +goose Up
UPDATE profile
SET avatar_url = CASE user_id
    WHEN 1001 THEN 'https://images.pexels.com/photos/8454926/pexels-photo-8454926.jpeg?auto=compress&cs=tinysrgb&fit=crop&w=800&h=800&crop=faces'
    WHEN 1004 THEN 'https://images.pexels.com/photos/6697367/pexels-photo-6697367.jpeg?auto=compress&cs=tinysrgb&fit=crop&w=800&h=800&crop=faces'
    ELSE avatar_url
  END,
  updated_at = NOW()
WHERE is_trainer = TRUE
  AND (
    (user_id = 1001 AND avatar_url = 'https://images.pexels.com/photos/8455977/pexels-photo-8455977.jpeg?auto=compress&cs=tinysrgb&w=800')
    OR (user_id = 1004 AND avatar_url = 'https://images.pexels.com/photos/4056441/pexels-photo-4056441.jpeg?auto=compress&cs=tinysrgb&w=800')
  );

-- +goose Down
UPDATE profile
SET avatar_url = CASE user_id
    WHEN 1001 THEN 'https://images.pexels.com/photos/8455977/pexels-photo-8455977.jpeg?auto=compress&cs=tinysrgb&w=800'
    WHEN 1004 THEN 'https://images.pexels.com/photos/4056441/pexels-photo-4056441.jpeg?auto=compress&cs=tinysrgb&w=800'
    ELSE avatar_url
  END,
  updated_at = NOW()
WHERE is_trainer = TRUE
  AND (
    (user_id = 1001 AND avatar_url = 'https://images.pexels.com/photos/8454926/pexels-photo-8454926.jpeg?auto=compress&cs=tinysrgb&fit=crop&w=800&h=800&crop=faces')
    OR (user_id = 1004 AND avatar_url = 'https://images.pexels.com/photos/6697367/pexels-photo-6697367.jpeg?auto=compress&cs=tinysrgb&fit=crop&w=800&h=800&crop=faces')
  );
