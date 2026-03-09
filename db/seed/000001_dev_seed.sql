INSERT INTO sport_type (name)
VALUES
    ('Бег'),
    ('Плавание'),
    ('Бокс'),
    ('Теннис')
ON CONFLICT (name) DO NOTHING;

INSERT INTO "user" (email, password_hash)
VALUES
    ('trainer@example.com', '$2a$10$exampletrainerhash'),
    ('client@example.com', '$2a$10$exampleclienthash'),
    ('mamapapaya@example.com', '$2a$10$eDl.rzMw6ldNi/GbGiT2QuV.ED8Y44E5vhObrArjoKUbXur9gVU.i')
ON CONFLICT (email) DO UPDATE
SET password_hash = EXCLUDED.password_hash;

INSERT INTO user_profile (user_id, username, first_name, last_name, bio, avatar_url)
SELECT
    u.user_id,
    'trainer_one',
    'Иван',
    'Тренеров',
    'Тренер по бегу и плаванию',
    'https://placehold.co/200x200'
FROM "user" u
WHERE u.email = 'trainer@example.com'
ON CONFLICT (user_id) DO UPDATE
SET
    username = EXCLUDED.username,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    bio = EXCLUDED.bio,
    avatar_url = EXCLUDED.avatar_url;

INSERT INTO user_profile (user_id, username, first_name, last_name, bio, avatar_url)
SELECT
    u.user_id,
    'client_one',
    'Петр',
    'Клиентов',
    'Люблю спорт',
    'https://placehold.co/200x200'
FROM "user" u
WHERE u.email = 'client@example.com'
ON CONFLICT (user_id) DO UPDATE
SET
    username = EXCLUDED.username,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    bio = EXCLUDED.bio,
    avatar_url = EXCLUDED.avatar_url;

INSERT INTO user_profile (user_id, username, first_name, last_name, bio, avatar_url)
SELECT
    u.user_id,
    'mamapapaya',
    'Mama',
    'Papaya',
    NULL,
    NULL
FROM "user" u
WHERE u.email = 'mamapapaya@example.com'
ON CONFLICT (user_id) DO UPDATE
SET
    username = EXCLUDED.username,
    first_name = EXCLUDED.first_name,
    last_name = EXCLUDED.last_name,
    bio = EXCLUDED.bio,
    avatar_url = EXCLUDED.avatar_url;

INSERT INTO trainer_details (trainer_user_id, education_degree, career_since_date)
SELECT
    u.user_id,
    'Bachelor of Sports Science',
    DATE '2020-01-01'
FROM "user" u
WHERE u.email = 'trainer@example.com'
ON CONFLICT (trainer_user_id) DO UPDATE
SET
    education_degree = EXCLUDED.education_degree,
    career_since_date = EXCLUDED.career_since_date;

INSERT INTO trainer_to_sport_type (trainer_id, sport_type_id, experience_years, sports_rank)
SELECT
    u.user_id,
    st.sport_type_id,
    5,
    'КМС'
FROM "user" u
JOIN sport_type st ON st.name = 'Бег'
WHERE u.email = 'trainer@example.com'
ON CONFLICT (trainer_id, sport_type_id) DO UPDATE
SET
    experience_years = EXCLUDED.experience_years,
    sports_rank = EXCLUDED.sports_rank;

INSERT INTO trainer_to_sport_type (trainer_id, sport_type_id, experience_years, sports_rank)
SELECT
    u.user_id,
    st.sport_type_id,
    3,
    NULL
FROM "user" u
JOIN sport_type st ON st.name = 'Плавание'
WHERE u.email = 'trainer@example.com'
ON CONFLICT (trainer_id, sport_type_id) DO UPDATE
SET
    experience_years = EXCLUDED.experience_years,
    sports_rank = EXCLUDED.sports_rank;

INSERT INTO subscription_tier (
    trainer_id,
    title,
    description,
    price_currency,
    price_value,
    price_exponent,
    level_rank,
    is_archived,
    archived_at
)
SELECT
    u.user_id,
    'Базовый',
    'Доступ к базовым материалам',
    'RUB',
    50000,
    2,
    1,
    false,
    NULL
FROM "user" u
WHERE u.email = 'trainer@example.com'
ON CONFLICT (trainer_id, level_rank) DO UPDATE
SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    price_currency = EXCLUDED.price_currency,
    price_value = EXCLUDED.price_value,
    price_exponent = EXCLUDED.price_exponent,
    is_archived = EXCLUDED.is_archived,
    archived_at = EXCLUDED.archived_at;

INSERT INTO subscription_tier (
    trainer_id,
    title,
    description,
    price_currency,
    price_value,
    price_exponent,
    level_rank,
    is_archived,
    archived_at
)
SELECT
    u.user_id,
    'Премиум',
    'Доступ ко всем материалам',
    'RUB',
    150000,
    2,
    2,
    false,
    NULL
FROM "user" u
WHERE u.email = 'trainer@example.com'
ON CONFLICT (trainer_id, level_rank) DO UPDATE
SET
    title = EXCLUDED.title,
    description = EXCLUDED.description,
    price_currency = EXCLUDED.price_currency,
    price_value = EXCLUDED.price_value,
    price_exponent = EXCLUDED.price_exponent,
    is_archived = EXCLUDED.is_archived,
    archived_at = EXCLUDED.archived_at;

INSERT INTO post (
    trainer_id,
    min_tier_id,
    title,
    text_content
)
SELECT
    u.user_id,
    NULL,
    'План тренировок на неделю',
    'Три силовые и две кардио тренировки.'
FROM "user" u
WHERE u.email = 'trainer@example.com'
AND NOT EXISTS (
    SELECT 1
    FROM post p
    WHERE p.trainer_id = u.user_id
      AND p.title = 'План тренировок на неделю'
);

INSERT INTO post (
    trainer_id,
    min_tier_id,
    title,
    text_content
)
SELECT
    u.user_id,
    st.subscription_tier_id,
    'Закрытая программа для подписчиков',
    'Подробная программа для подписчиков базового tier и выше.'
FROM "user" u
JOIN subscription_tier st
  ON st.trainer_id = u.user_id
 AND st.level_rank = 1
WHERE u.email = 'trainer@example.com'
AND NOT EXISTS (
    SELECT 1
    FROM post p
    WHERE p.trainer_id = u.user_id
      AND p.title = 'Закрытая программа для подписчиков'
);

INSERT INTO post_attachment (post_id, file_url, kind)
SELECT
    p.post_id,
    'https://placehold.co/600x400',
    'image'::attachment_kind
FROM post p
JOIN "user" u ON u.user_id = p.trainer_id
WHERE u.email = 'trainer@example.com'
  AND p.title = 'План тренировок на неделю'
  AND NOT EXISTS (
      SELECT 1
      FROM post_attachment pa
      WHERE pa.post_id = p.post_id
        AND pa.file_url = 'https://placehold.co/600x400'
  );

INSERT INTO post_attachment (post_id, file_url, kind)
SELECT
    p.post_id,
    'https://placehold.co/800x600',
    'document'::attachment_kind
FROM post p
JOIN "user" u ON u.user_id = p.trainer_id
WHERE u.email = 'trainer@example.com'
  AND p.title = 'Закрытая программа для подписчиков'
  AND NOT EXISTS (
      SELECT 1
      FROM post_attachment pa
      WHERE pa.post_id = p.post_id
        AND pa.file_url = 'https://placehold.co/800x600'
  );
