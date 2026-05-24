-- +goose Up

UPDATE content_post
SET sport_type_id = CASE post_id
  WHEN 2003 THEN 3002
  WHEN 2004 THEN 3003
  WHEN 2005 THEN 3003
  WHEN 2006 THEN 3005
  WHEN 2007 THEN 3005
  WHEN 2008 THEN 3006
  WHEN 2009 THEN 3006
  WHEN 2010 THEN 3004
  WHEN 2011 THEN 3004
  ELSE sport_type_id
END
WHERE post_id IN (2003, 2004, 2005, 2006, 2007, 2008, 2009, 2010, 2011);

UPDATE content_subscription_tier
SET chat_enabled = true
WHERE (trainer_user_id = 1001 AND tier_id = 2) 
   OR (trainer_user_id = 1004 AND tier_id = 2) 
   OR (trainer_user_id = 1005 AND tier_id = 3) 
   OR (trainer_user_id = 1006 AND tier_id = 2) 
   OR (trainer_user_id = 1007 AND tier_id = 2); 

INSERT INTO content_comment (comment_id, post_id, author_user_id, body)
VALUES
  (2201, 2001, 1002, 'Спасибо за план! Начал с трёх тренировок в неделю, уже через две чувствую прогресс.'),
  (2202, 2001, 1008, 'Подскажите, как часто делать длинную тренировку в самом начале?'),
  (2203, 2001, 1001, 'Мария, в первые две недели — раз в неделю, не больше 70–80% от целевого темпа. Главное — не торопиться.'),
  (2204, 2003, 1009, 'Попробовал методику дыхания — стало намного комфортнее в воде. Раньше задерживал дыхание и уставал быстро.'),
  (2205, 2003, 1001, 'Главное — не форсировать, тело само подстраивается за несколько тренировок.'),
  (2206, 2004, 1008, 'Утренний комплекс просто спасение для спины. Делаю уже неделю — чувствую ощутимую разницу.'),
  (2207, 2004, 1002, 'Попробовал кошку-корову после длинной пробежки — спина расслабляется лучше, чем от простой растяжки.'),
  (2208, 2006, 1009, 'Стойка стала намного устойчивее. Теперь понимаю, почему локти так важны — сразу уменьшилась нагрузка на плечи.'),
  (2209, 2008, 1002, 'Сделал три круга по схеме 40/20 — отличная нагрузка, инвентарь не нужен.'),
  (2210, 2010, 1008, 'Держать 85 оборотов в минуту оказалось сложнее, чем казалось. Буду тренироваться с метрономом.'),
  (2211, 2010, 1009, 'После третьей тренировки каденс стал ровнее, дыхание тоже улучшилось.')
ON CONFLICT (comment_id) DO UPDATE
SET post_id       = EXCLUDED.post_id,
    author_user_id = EXCLUDED.author_user_id,
    body          = EXCLUDED.body,
    updated_at    = now();

SELECT setval(
  pg_get_serial_sequence('content_comment', 'comment_id'),
  (SELECT GREATEST(COALESCE(MAX(comment_id), 1), 2211) FROM content_comment),
  true
);

INSERT INTO content_donation (donation_id, sender_user_id, recipient_user_id, amount_value, currency, message)
VALUES
  (2301, 1002, 1001, 300, 'RUB', 'Анна, отличные материалы по подготовке — спасибо!'),
  (2302, 1008, 1004, 250, 'RUB', 'Елена, комплекс йоги реально помог со спиной.'),
  (2303, 1009, 1005, 500, 'RUB', 'Сергей, лучшее объяснение стойки, что видел — спасибо!'),
  (2304, 1008, 1006, 200, 'RUB', 'Ольга, домашняя тренировка зашла на ура, уже третий раз делаю.')
ON CONFLICT (donation_id) DO UPDATE
SET sender_user_id    = EXCLUDED.sender_user_id,
    recipient_user_id = EXCLUDED.recipient_user_id,
    amount_value      = EXCLUDED.amount_value,
    currency          = EXCLUDED.currency,
    message           = EXCLUDED.message;

SELECT setval(
  pg_get_serial_sequence('content_donation', 'donation_id'),
  (SELECT GREATEST(COALESCE(MAX(donation_id), 1), 2304) FROM content_donation),
  true
);

INSERT INTO content_notification
  (notification_id, user_id, type, actor_user_id, title, body,
   post_id, comment_id, donation_id, subscription_id, read_at)
VALUES
  (2501, 1001, 'subscription', 1002, 'Новый подписчик',
   'Иван Сидоров оформил подписку «Продвинутый».',
   NULL, NULL, NULL, 2401, now() - INTERVAL '25 days'),
  (2502, 1004, 'subscription', 1008, 'Новый подписчик',
   'Мария Федорова оформила подписку «Глубокая практика».',
   NULL, NULL, NULL, 2402, now() - INTERVAL '20 days'),
  (2503, 1005, 'subscription', 1009, 'Новый подписчик',
   'Павел Никитин оформил подписку «Премиум».',
   NULL, NULL, NULL, 2403, now() - INTERVAL '15 days'),

  (2504, 1001, 'like', 1008, 'Новый лайк',
   'Мария Федорова оценила ваш пост «План подготовки к полумарафону».',
   2001, NULL, NULL, NULL, now() - INTERVAL '10 days'),
  (2505, 1003, 'like', 1008, 'Новый лайк',
   'Мария Федорова оценила ваш пост «Техника дыхания в бассейне».',
   2003, NULL, NULL, NULL, NULL),
  (2506, 1004, 'like', 1002, 'Новый лайк',
   'Иван Сидоров оценил ваш пост «Утренняя йога для спины».',
   2004, NULL, NULL, NULL, NULL),
  (2507, 1006, 'like', 1005, 'Новый лайк',
   'Сергей Орлов оценил ваш пост «Силовая тренировка дома без инвентаря».',
   2008, NULL, NULL, NULL, NULL),
  (2508, 1007, 'like', 1008, 'Новый лайк',
   'Мария Федорова оценила ваш пост «Как держать ровный каденс на велосипеде».',
   2010, NULL, NULL, NULL, NULL),
  (2509, 1007, 'like', 1009, 'Новый лайк',
   'Павел Никитин оценил ваш пост «Как держать ровный каденс на велосипеде».',
   2010, NULL, NULL, NULL, NULL),

  (2510, 1001, 'comment', 1008, 'Новый комментарий',
   'Мария Федорова спросила: «Как часто делать длинную тренировку в самом начале?»',
   2001, 2202, NULL, NULL, NULL),
  (2511, 1001, 'comment', 1002, 'Новый комментарий',
   'Иван Сидоров написал: «Спасибо за план! Начал с трёх тренировок, уже вижу прогресс.»',
   2001, 2201, NULL, NULL, now() - INTERVAL '5 days'),
  (2512, 1004, 'comment', 1008, 'Новый комментарий',
   'Мария Федорова написала: «Утренний комплекс просто спасение для спины.»',
   2004, 2206, NULL, NULL, now() - INTERVAL '3 days'),
  (2513, 1005, 'comment', 1009, 'Новый комментарий',
   'Павел Никитин написал: «Стойка стала намного устойчивее.»',
   2006, 2208, NULL, NULL, NULL),
  (2514, 1006, 'comment', 1002, 'Новый комментарий',
   'Иван Сидоров написал: «Сделал три круга по схеме 40/20 — отличная нагрузка.»',
   2008, 2209, NULL, NULL, NULL),

  (2515, 1001, 'donation', 1002, 'Новый донат',
   'Иван Сидоров отправил 300 ₽: «Анна, отличные материалы по подготовке — спасибо!»',
   NULL, NULL, 2301, NULL, now() - INTERVAL '8 days'),
  (2516, 1004, 'donation', 1008, 'Новый донат',
   'Мария Федорова отправила 250 ₽: «Елена, комплекс йоги реально помог со спиной.»',
   NULL, NULL, 2302, NULL, NULL),
  (2517, 1005, 'donation', 1009, 'Новый донат',
   'Павел Никитин отправил 500 ₽: «Сергей, лучшее объяснение стойки, что видел!»',
   NULL, NULL, 2303, NULL, NULL),
  (2518, 1006, 'donation', 1008, 'Новый донат',
   'Мария Федорова отправила 200 ₽: «Ольга, домашняя тренировка зашла на ура.»',
   NULL, NULL, 2304, NULL, NULL),

  (2519, 1002, 'post', 1001, 'Новый пост от тренера',
   'Анна Петрова опубликовала: «Закрытая тренировка по темпу».',
   2002, NULL, NULL, NULL, now() - INTERVAL '18 days'),
  (2520, 1008, 'post', 1004, 'Новый пост от тренера',
   'Елена Смирнова опубликовала: «Закрытый комплекс глубокой растяжки».',
   2005, NULL, NULL, NULL, NULL),
  (2521, 1009, 'post', 1005, 'Новый пост от тренера',
   'Сергей Орлов опубликовал: «Разбор защитных действий для подписчиков».',
   2007, NULL, NULL, NULL, NULL)
ON CONFLICT (notification_id) DO UPDATE
SET user_id         = EXCLUDED.user_id,
    type            = EXCLUDED.type,
    actor_user_id   = EXCLUDED.actor_user_id,
    title           = EXCLUDED.title,
    body            = EXCLUDED.body,
    post_id         = EXCLUDED.post_id,
    comment_id      = EXCLUDED.comment_id,
    donation_id     = EXCLUDED.donation_id,
    subscription_id = EXCLUDED.subscription_id,
    read_at         = EXCLUDED.read_at,
    updated_at      = now();

SELECT setval(
  pg_get_serial_sequence('content_notification', 'notification_id'),
  (SELECT GREATEST(COALESCE(MAX(notification_id), 1), 2521) FROM content_notification),
  true
);

INSERT INTO content_chat_message (message_id, sender_user_id, receiver_user_id, body, is_read)
VALUES
  (2601, 1002, 1001,
   'Анна, добрый день! Хотел спросить про темп на длинной тренировке — насколько можно разгоняться?',
   true),
  (2602, 1001, 1002,
   'Иван, привет! На длинных держите «разговорный» темп — должны уметь говорить короткими фразами. Ориентир — пульс 130–145 уд/мин.',
   true),
  (2603, 1002, 1001,
   'Понял, спасибо! Попробую в воскресенье. Пульсометра нет, ориентируюсь на дыхание.',
   true),
  (2604, 1001, 1002,
   'Дыхание — отличный ориентир. Если можете считать до 4 на выдох — темп правильный. Удачи в воскресенье!',
   false),

  (2605, 1008, 1004,
   'Елена, здравствуйте! Можно ли делать ваш утренний комплекс каждый день?',
   true),
  (2606, 1004, 1008,
   'Мария, привет! Да, утренний комплекс можно ежедневно — он мягкий. Главное, не через боль.',
   true),
  (2607, 1008, 1004,
   'Отлично, буду делать каждое утро. Спину уже меньше тянет после первой недели.',
   true),
  (2608, 1004, 1008,
   'Это хороший знак! Если будет дискомфорт в каком-то упражнении — пишите, подберём замену.',
   false),

  (2609, 1009, 1005,
   'Сергей, привет! Как часто можно отрабатывать защиту — каждый день или нужны дни отдыха?',
   true),
  (2610, 1005, 1009,
   'Павел, привет! Технические упражнения без контакта — можно каждый день. Спарринги — не чаще 2–3 раз в неделю.',
   true),
  (2611, 1009, 1005,
   'Понял. Пока без спаррингов, ставлю технику. На этой неделе отработал уклоны по вашему разбору.',
   true),
  (2612, 1005, 1009,
   'Отлично! Снимите видео на следующей тренировке — посмотрю и дам обратную связь.',
   false)
ON CONFLICT (message_id) DO UPDATE
SET sender_user_id   = EXCLUDED.sender_user_id,
    receiver_user_id = EXCLUDED.receiver_user_id,
    body             = EXCLUDED.body,
    is_read          = EXCLUDED.is_read;

SELECT setval(
  pg_get_serial_sequence('content_chat_message', 'message_id'),
  (SELECT GREATEST(COALESCE(MAX(message_id), 1), 2612) FROM content_chat_message),
  true
);

-- +goose Down
DELETE FROM content_chat_message
WHERE message_id BETWEEN 2601 AND 2612;

DELETE FROM content_notification
WHERE notification_id BETWEEN 2501 AND 2521;

DELETE FROM content_donation
WHERE donation_id BETWEEN 2301 AND 2304;

DELETE FROM content_comment
WHERE comment_id BETWEEN 2201 AND 2211;

UPDATE content_subscription_tier
SET chat_enabled = false
WHERE (trainer_user_id = 1001 AND tier_id = 2)
   OR (trainer_user_id = 1004 AND tier_id = 2)
   OR (trainer_user_id = 1005 AND tier_id = 3)
   OR (trainer_user_id = 1006 AND tier_id = 2)
   OR (trainer_user_id = 1007 AND tier_id = 2);

UPDATE content_post
SET sport_type_id = CASE post_id
  WHEN 2003 THEN 3003
  WHEN 2004 THEN 3005
  WHEN 2005 THEN 3005
  WHEN 2006 THEN 3006
  WHEN 2007 THEN 3006
  WHEN 2008 THEN 3007
  WHEN 2009 THEN 3007
  WHEN 2010 THEN 3008
  WHEN 2011 THEN 3008
  ELSE sport_type_id
END
WHERE post_id IN (2003, 2004, 2005, 2006, 2007, 2008, 2009, 2010, 2011);
