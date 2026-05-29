-- +goose Up
UPDATE profile
SET avatar_url = CASE user_id
    WHEN 1001 THEN 'https://images.pexels.com/photos/8455977/pexels-photo-8455977.jpeg?auto=compress&cs=tinysrgb&w=800'
    WHEN 1003 THEN 'https://images.pexels.com/photos/18167936/pexels-photo-18167936.jpeg?auto=compress&cs=tinysrgb&w=800'
    WHEN 1004 THEN 'https://images.pexels.com/photos/4056441/pexels-photo-4056441.jpeg?auto=compress&cs=tinysrgb&w=800'
    WHEN 1005 THEN 'https://images.pexels.com/photos/9944859/pexels-photo-9944859.jpeg?auto=compress&cs=tinysrgb&w=800'
    WHEN 1006 THEN 'https://images.pexels.com/photos/6739931/pexels-photo-6739931.jpeg?auto=compress&cs=tinysrgb&w=800'
    WHEN 1007 THEN 'https://images.pexels.com/photos/18021124/pexels-photo-18021124.jpeg?auto=compress&cs=tinysrgb&w=800'
    ELSE avatar_url
  END,
  updated_at = NOW()
WHERE user_id IN (1001, 1003, 1004, 1005, 1006, 1007)
  AND is_trainer = TRUE
  AND avatar_url IN (
    'https://randomuser.me/api/portraits/women/44.jpg',
    'https://randomuser.me/api/portraits/men/75.jpg',
    'https://randomuser.me/api/portraits/women/65.jpg',
    'https://randomuser.me/api/portraits/men/32.jpg',
    'https://randomuser.me/api/portraits/women/68.jpg',
    'https://randomuser.me/api/portraits/men/46.jpg'
  );

-- +goose Down
SELECT 1;
