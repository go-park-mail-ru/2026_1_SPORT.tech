INSERT INTO content_post (post_id, author_user_id, title, required_subscription_level)
VALUES
  (2004, 1004, 'Утренняя йога для спины', NULL),
  (2005, 1004, 'Закрытый комплекс глубокой растяжки', 2),
  (2006, 1005, 'Боксерская стойка и работа ног', NULL),
  (2007, 1005, 'Разбор защитных действий для подписчиков', 3),
  (2008, 1006, 'Силовая тренировка дома без инвентаря', NULL),
  (2009, 1006, 'Программа прогрессии в приседаниях', 2),
  (2010, 1007, 'Как держать ровный каденс на велосипеде', NULL),
  (2011, 1007, 'План велоподготовки на месяц', 2)
ON CONFLICT (post_id) DO UPDATE
SET author_user_id = EXCLUDED.author_user_id,
    title = EXCLUDED.title,
    required_subscription_level = EXCLUDED.required_subscription_level,
    updated_at = NOW();

INSERT INTO content_post_block (post_block_id, post_id, position, kind, text_content, file_url)
VALUES
  (2107, 2001, 2, 'text', 'Если пульс растет слишком быстро, снижаем темп и оставляем запас на последнюю треть дистанции.', NULL),
  (2108, 2002, 2, 'text', 'Главное правило: темповая работа не должна превращаться в гонку на каждой тренировке.', NULL),
  (2109, 2003, 2, 'text', 'Попробуйте выдыхать спокойно в воду на 3-4 счета, а вдох делать коротким и без подъема головы.', NULL),
  (2110, 2004, 0, 'text', 'Пять минут мягкой разминки помогают убрать скованность после сна и подготовить спину к дню.', NULL),
  (2111, 2004, 1, 'image', NULL, 'https://images.unsplash.com/photo-1544367567-0f2fcb009e0b'),
  (2112, 2004, 2, 'text', 'Двигайтесь без боли: кошка-корова, вытяжение в позе ребенка и легкие скручивания.', NULL),
  (2113, 2005, 0, 'text', 'Закрытый комплекс для подписчиков: глубокая растяжка задней поверхности бедра и грудного отдела.', NULL),
  (2114, 2005, 1, 'video', NULL, 'https://example.com/videos/deep-stretching-flow.mp4'),
  (2115, 2005, 2, 'text', 'Каждое положение удерживаем 40-60 секунд, дыхание спокойное, без рывков.', NULL),
  (2116, 2006, 0, 'text', 'Начинаем с базовой стойки: подбородок ниже, локти ближе к корпусу, вес распределен мягко.', NULL),
  (2117, 2006, 1, 'image', NULL, 'https://images.unsplash.com/photo-1549719386-74dfcbf7dbed'),
  (2118, 2006, 2, 'text', 'После стойки добавляем короткие шаги вперед-назад и в сторону, не скрещивая ноги.', NULL),
  (2119, 2007, 0, 'text', 'Премиум-разбор: уклоны, нырки и выходы из угла после атаки соперника.', NULL),
  (2120, 2007, 1, 'document', NULL, 'https://example.com/docs/boxing-defense-checklist.pdf'),
  (2121, 2007, 2, 'text', 'Отрабатывайте защиту сначала медленно, потом добавляйте скорость только без потери баланса.', NULL),
  (2122, 2008, 0, 'text', 'Эта тренировка подходит для дома: приседания, отжимания, ягодичный мост и планка.', NULL),
  (2123, 2008, 1, 'image', NULL, 'https://images.unsplash.com/photo-1518611012118-696072aa579a'),
  (2124, 2008, 2, 'text', 'Сделайте 3 круга по 40 секунд работы и 20 секунд отдыха между упражнениями.', NULL),
  (2125, 2009, 0, 'text', 'Закрытая программа: как увеличить объем приседаний без перегруза коленей и поясницы.', NULL),
  (2126, 2009, 1, 'document', NULL, 'https://example.com/docs/squat-progression.pdf'),
  (2127, 2009, 2, 'text', 'Добавляйте нагрузку только если техника остается стабильной на последних повторениях.', NULL),
  (2128, 2010, 0, 'text', 'Ровный каденс помогает экономить силы на длинной дистанции и не забивать ноги на подъемах.', NULL),
  (2129, 2010, 1, 'image', NULL, 'https://images.unsplash.com/photo-1485965120184-e220f721d03e'),
  (2130, 2010, 2, 'text', 'Начните с диапазона 80-90 оборотов в минуту и следите, чтобы дыхание оставалось контролируемым.', NULL),
  (2131, 2011, 0, 'text', 'План для подписчиков: четыре недели с постепенным ростом объема и одной восстановительной неделей.', NULL),
  (2132, 2011, 1, 'document', NULL, 'https://example.com/docs/cycling-month-plan.pdf'),
  (2133, 2011, 2, 'text', 'Не пропускайте легкие дни: именно на них организм адаптируется к нагрузке.', NULL)
ON CONFLICT (post_block_id) DO UPDATE
SET post_id = EXCLUDED.post_id,
    position = EXCLUDED.position,
    kind = EXCLUDED.kind,
    text_content = EXCLUDED.text_content,
    file_url = EXCLUDED.file_url,
    updated_at = NOW();

INSERT INTO content_comment (comment_id, post_id, author_user_id, body)
VALUES
  (2203, 2004, 1008, 'После такого комплекса реально легче сидеть за ноутбуком. Спасибо!'),
  (2204, 2006, 1009, 'Не думал, что работа ног настолько важна. Буду отрабатывать перед зеркалом.'),
  (2205, 2008, 1002, 'Круговая тренировка зашла. Можно потом вариант посложнее?'),
  (2206, 2010, 1008, 'Про каденс полезно, раньше всегда ехала слишком тяжело на высокой передаче.'),
  (2207, 2001, 1009, 'Анна, а этот план подойдет, если я бегаю только два раза в неделю?'),
  (2208, 2003, 1008, 'Поняла, почему быстро уставала в бассейне. Буду тренировать выдох.'),
  (2209, 2006, 1006, 'Хорошая база. Я бы еще добавила разминку плеч перед ударной работой.'),
  (2210, 2008, 1005, 'Отличный формат для новичков: коротко и без лишнего оборудования.')
ON CONFLICT (comment_id) DO UPDATE
SET post_id = EXCLUDED.post_id,
    author_user_id = EXCLUDED.author_user_id,
    body = EXCLUDED.body,
    updated_at = NOW();

INSERT INTO content_post_like (post_id, user_id)
VALUES
  (2001, 1008),
  (2003, 1008),
  (2004, 1002),
  (2004, 1008),
  (2006, 1009),
  (2008, 1002),
  (2008, 1005),
  (2010, 1008),
  (2010, 1009)
ON CONFLICT (post_id, user_id) DO NOTHING;

SELECT setval(
  pg_get_serial_sequence('content_post', 'post_id'),
  (SELECT GREATEST(COALESCE(MAX(post_id), 1), 2011) FROM content_post),
  true
);

SELECT setval(
  pg_get_serial_sequence('content_post_block', 'post_block_id'),
  (SELECT GREATEST(COALESCE(MAX(post_block_id), 1), 2133) FROM content_post_block),
  true
);

SELECT setval(
  pg_get_serial_sequence('content_comment', 'comment_id'),
  (SELECT GREATEST(COALESCE(MAX(comment_id), 1), 2210) FROM content_comment),
  true
);
