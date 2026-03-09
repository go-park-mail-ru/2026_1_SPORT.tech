INSERT INTO sport_type (sport_type_id, name)
VALUES
    (1, 'Бег'),
    (2, 'Плавание'),
    (3, 'Бокс'),
    (4, 'Теннис')
ON CONFLICT (sport_type_id) DO NOTHING;

INSERT INTO "user" (user_id, email, password_hash)
VALUES
    (1, 'trainer@example.com', '$2a$10$exampletrainerhash'),
    (2, 'client@example.com', '$2a$10$exampleclienthash'),
    (3, 'mamapapaya@example.com', '$2a$10$eDl.rzMw6ldNi/GbGiT2QuV.ED8Y44E5vhObrArjoKUbXur9gVU.i')
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO user_profile (user_id, username, first_name, last_name, bio, avatar_url)
VALUES
    (1, 'trainer_one', 'Иван', 'Тренеров', 'Тренер по бегу и плаванию', 'https://placehold.co/200x200'),
    (2, 'client_one', 'Петр', 'Клиентов', 'Люблю спорт', 'https://placehold.co/200x200'),
    (3, 'mamapapaya', 'Mama', 'Papaya', NULL, NULL)
ON CONFLICT (user_id) DO NOTHING;

INSERT INTO trainer_details (trainer_user_id, education_degree, career_since_date)
VALUES
    (1, 'Bachelor of Sports Science', '2020-01-01')
ON CONFLICT (trainer_user_id) DO NOTHING;

INSERT INTO trainer_to_sport_type (trainer_id, sport_type_id, experience_years, sports_rank)
VALUES
    (1, 1, 5, 'КМС'),
    (1, 2, 3, NULL)
ON CONFLICT (trainer_id, sport_type_id) DO NOTHING;

INSERT INTO subscription_tier (
    subscription_tier_id,
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
VALUES
    (1, 1, 'Базовый', 'Доступ к базовым материалам', 'RUB', 50000, 2, 1, false, NULL),
    (2, 1, 'Премиум', 'Доступ ко всем материалам', 'RUB', 150000, 2, 2, false, NULL)
ON CONFLICT (subscription_tier_id) DO NOTHING;

INSERT INTO post (
    post_id,
    trainer_id,
    min_tier_id,
    title,
    text_content
)
VALUES
    (1, 1, NULL, 'План тренировок на неделю', 'Три силовые и две кардио тренировки.'),
    (2, 1, 1, 'Закрытая программа для подписчиков', 'Подробная программа для подписчиков базового tier и выше.')
ON CONFLICT (post_id) DO NOTHING;

INSERT INTO post_attachment (post_attachment_id, post_id, file_url, kind)
VALUES
    (1, 1, 'https://placehold.co/600x400', 'image'),
    (2, 2, 'https://placehold.co/800x600', 'document')
ON CONFLICT (post_attachment_id) DO NOTHING;
