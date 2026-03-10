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

INSERT INTO user_subscription (
    subscriber_user_id,
    subscription_tier_id,
    started_at,
    expires_at
)
SELECT
    subscriber.user_id,
    tier.subscription_tier_id,
    now() - interval '7 days',
    now() + interval '30 days'
FROM "user" subscriber
JOIN "user" trainer ON trainer.email = 'trainer@example.com'
JOIN subscription_tier tier
  ON tier.trainer_id = trainer.user_id
 AND tier.level_rank = 1
WHERE subscriber.email = 'client@example.com'
AND NOT EXISTS (
    SELECT 1
    FROM user_subscription us
    WHERE us.subscriber_user_id = subscriber.user_id
      AND us.subscription_tier_id = tier.subscription_tier_id
      AND us.expires_at > now()
);

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

INSERT INTO post (
    trainer_id,
    min_tier_id,
    title,
    text_content,
    created_at,
    updated_at
)
SELECT
    u.user_id,
    NULL,
    'Разбор техники бега на 5 км',
    'Показываю, как держать корпус, куда ставить стопу и как не закисляться на первых километрах.',
    now() - interval '6 days',
    now() - interval '6 days'
FROM "user" u
WHERE u.email = 'trainer@example.com'
AND NOT EXISTS (
    SELECT 1
    FROM post p
    WHERE p.trainer_id = u.user_id
      AND p.title = 'Разбор техники бега на 5 км'
);

INSERT INTO post (
    trainer_id,
    min_tier_id,
    title,
    text_content,
    created_at,
    updated_at
)
SELECT
    u.user_id,
    NULL,
    'Утренняя мобилизация перед бассейном',
    'Короткий комплекс на плечи, грудной отдел и голеностоп перед плавательной тренировкой.',
    now() - interval '5 days',
    now() - interval '5 days'
FROM "user" u
WHERE u.email = 'trainer@example.com'
AND NOT EXISTS (
    SELECT 1
    FROM post p
    WHERE p.trainer_id = u.user_id
      AND p.title = 'Утренняя мобилизация перед бассейном'
);

INSERT INTO post (
    trainer_id,
    min_tier_id,
    title,
    text_content,
    created_at,
    updated_at
)
SELECT
    u.user_id,
    st.subscription_tier_id,
    'Тренировка корпуса и дыхания',
    'Силовой блок на стабилизацию корпуса плюс дыхательный протокол для длинных серий.',
    now() - interval '4 days',
    now() - interval '4 days'
FROM "user" u
JOIN subscription_tier st
  ON st.trainer_id = u.user_id
 AND st.level_rank = 1
WHERE u.email = 'trainer@example.com'
AND NOT EXISTS (
    SELECT 1
    FROM post p
    WHERE p.trainer_id = u.user_id
      AND p.title = 'Тренировка корпуса и дыхания'
);

INSERT INTO post (
    trainer_id,
    min_tier_id,
    title,
    text_content,
    created_at,
    updated_at
)
SELECT
    u.user_id,
    st.subscription_tier_id,
    'Премиум: микроцикл перед стартом',
    'Пошаговый план на последние 7 дней перед стартом: объем, интенсивность, сон и питание.',
    now() - interval '3 days',
    now() - interval '3 days'
FROM "user" u
JOIN subscription_tier st
  ON st.trainer_id = u.user_id
 AND st.level_rank = 2
WHERE u.email = 'trainer@example.com'
AND NOT EXISTS (
    SELECT 1
    FROM post p
    WHERE p.trainer_id = u.user_id
      AND p.title = 'Премиум: микроцикл перед стартом'
);

INSERT INTO post (
    trainer_id,
    min_tier_id,
    title,
    text_content,
    created_at,
    updated_at
)
SELECT
    u.user_id,
    NULL,
    'Ошибки восстановления после интервальных сессий',
    'Разбираю, почему после тяжелых интервалов нельзя сразу добивать себя объемом и как восстановиться быстрее.',
    now() - interval '2 days',
    now() - interval '2 days'
FROM "user" u
WHERE u.email = 'trainer@example.com'
AND NOT EXISTS (
    SELECT 1
    FROM post p
    WHERE p.trainer_id = u.user_id
      AND p.title = 'Ошибки восстановления после интервальных сессий'
);

INSERT INTO post (
    trainer_id,
    min_tier_id,
    title,
    text_content,
    created_at,
    updated_at
)
SELECT
    u.user_id,
    st.subscription_tier_id,
    'Чек-лист питания в день старта',
    'Готовый чек-лист по воде, углеводам и таймингу последнего приема пищи перед забегом.',
    now() - interval '1 day',
    now() - interval '1 day'
FROM "user" u
JOIN subscription_tier st
  ON st.trainer_id = u.user_id
 AND st.level_rank = 1
WHERE u.email = 'trainer@example.com'
AND NOT EXISTS (
    SELECT 1
    FROM post p
    WHERE p.trainer_id = u.user_id
      AND p.title = 'Чек-лист питания в день старта'
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

INSERT INTO post_attachment (post_id, file_url, kind)
SELECT
    p.post_id,
    'https://placehold.co/1200x800',
    'image'::attachment_kind
FROM post p
JOIN "user" u ON u.user_id = p.trainer_id
WHERE u.email = 'trainer@example.com'
  AND p.title = 'Разбор техники бега на 5 км'
  AND NOT EXISTS (
      SELECT 1
      FROM post_attachment pa
      WHERE pa.post_id = p.post_id
        AND pa.file_url = 'https://placehold.co/1200x800'
  );

INSERT INTO post_attachment (post_id, file_url, kind)
SELECT
    p.post_id,
    'https://placehold.co/1280x720',
    'image'::attachment_kind
FROM post p
JOIN "user" u ON u.user_id = p.trainer_id
WHERE u.email = 'trainer@example.com'
  AND p.title = 'Утренняя мобилизация перед бассейном'
  AND NOT EXISTS (
      SELECT 1
      FROM post_attachment pa
      WHERE pa.post_id = p.post_id
        AND pa.file_url = 'https://placehold.co/1280x720'
  );

INSERT INTO post_attachment (post_id, file_url, kind)
SELECT
    p.post_id,
    'https://placehold.co/1000x1400',
    'image'::attachment_kind
FROM post p
JOIN "user" u ON u.user_id = p.trainer_id
WHERE u.email = 'trainer@example.com'
  AND p.title = 'Тренировка корпуса и дыхания'
  AND NOT EXISTS (
      SELECT 1
      FROM post_attachment pa
      WHERE pa.post_id = p.post_id
        AND pa.file_url = 'https://placehold.co/1000x1400'
  );

INSERT INTO post_attachment (post_id, file_url, kind)
SELECT
    p.post_id,
    'https://placehold.co/1600x900',
    'image'::attachment_kind
FROM post p
JOIN "user" u ON u.user_id = p.trainer_id
WHERE u.email = 'trainer@example.com'
  AND p.title = 'Премиум: микроцикл перед стартом'
  AND NOT EXISTS (
      SELECT 1
      FROM post_attachment pa
      WHERE pa.post_id = p.post_id
        AND pa.file_url = 'https://placehold.co/1600x900'
  );

INSERT INTO post_attachment (post_id, file_url, kind)
SELECT
    p.post_id,
    'https://placehold.co/1080x1080',
    'image'::attachment_kind
FROM post p
JOIN "user" u ON u.user_id = p.trainer_id
WHERE u.email = 'trainer@example.com'
  AND p.title = 'Ошибки восстановления после интервальных сессий'
  AND NOT EXISTS (
      SELECT 1
      FROM post_attachment pa
      WHERE pa.post_id = p.post_id
        AND pa.file_url = 'https://placehold.co/1080x1080'
  );

INSERT INTO post_attachment (post_id, file_url, kind)
SELECT
    p.post_id,
    'https://placehold.co/900x1200',
    'image'::attachment_kind
FROM post p
JOIN "user" u ON u.user_id = p.trainer_id
WHERE u.email = 'trainer@example.com'
  AND p.title = 'Чек-лист питания в день старта'
  AND NOT EXISTS (
      SELECT 1
      FROM post_attachment pa
      WHERE pa.post_id = p.post_id
        AND pa.file_url = 'https://placehold.co/900x1200'
  );
