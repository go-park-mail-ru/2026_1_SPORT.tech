INSERT INTO content_post (post_id, author_user_id, title, required_subscription_level)
VALUES
  (2001, 1001, 'План подготовки к полумарафону', NULL),
  (2002, 1001, 'Закрытая тренировка по темпу', 2),
  (2003, 1003, 'Техника дыхания в бассейне', NULL);

INSERT INTO content_post_block (post_block_id, post_id, position, kind, text_content, file_url)
VALUES
  (2101, 2001, 0, 'text', 'Начинаем с 3 беговых тренировок в неделю: легкий бег, интервалы и длинная тренировка.', NULL),
  (2102, 2001, 1, 'image', NULL, 'https://images.unsplash.com/photo-1552674605-db6ffd4facb5'),
  (2103, 2002, 0, 'text', 'Закрытый материал для подписчиков: разбор темповых интервалов и недельной нагрузки.', NULL),
  (2104, 2002, 1, 'document', NULL, 'https://example.com/docs/tempo-workout-plan.pdf'),
  (2105, 2003, 0, 'text', 'Главная ошибка новичков в бассейне — задержка выдоха под водой. Ниже короткая памятка.', NULL),
  (2106, 2003, 1, 'video', NULL, 'https://example.com/videos/swim-breathing-drill.mp4');

INSERT INTO content_comment (comment_id, post_id, author_user_id, body)
VALUES
  (2201, 2001, 1002, 'Спасибо, как раз искал понятный стартовый план на 8 недель.'),
  (2202, 2003, 1001, 'Отличный разбор. Особенно полезен блок про ритм дыхания на развороте.');

INSERT INTO content_post_like (post_id, user_id)
VALUES
  (2001, 1002),
  (2003, 1001);

SELECT setval(
  pg_get_serial_sequence('content_post', 'post_id'),
  (SELECT GREATEST(COALESCE(MAX(post_id), 1), 2003) FROM content_post),
  true
);

SELECT setval(
  pg_get_serial_sequence('content_post_block', 'post_block_id'),
  (SELECT GREATEST(COALESCE(MAX(post_block_id), 1), 2106) FROM content_post_block),
  true
);

SELECT setval(
  pg_get_serial_sequence('content_comment', 'comment_id'),
  (SELECT GREATEST(COALESCE(MAX(comment_id), 1), 2202) FROM content_comment),
  true
);
