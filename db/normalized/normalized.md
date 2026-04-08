# ДЗ 1: Проектирование БД - SPORT.tech (Аналог Patreon для фитнес-тренеров)

## 1. Краткое словесное описание таблиц и функциональные зависимости

Ниже представлен список всех отношений (таблиц), их назначение в рамках бизнес-логики платформы SPORT.tech и функциональные зависимости атрибутов.

### 1. Базовая авторизация (user)

Назначение: хранит учетные данные для входа и служебные поля. Общая таблица для всех пользователей.
- user_id (PK): Идентификатор пользователя
- email (UK): Электронная почта
- password_hash: Хэш пароля
- created_at: Дата и время создания записи
- updated_at: Дата и время последнего обновления записи

**Функциональные зависимости:**

{user_id} -> email, password_hash, created_at, updated_at
{email} -> user_id, password_hash, created_at, updated_at

### 2. Публичный профиль пользователя (user_profile)

Назначение: общие данные профиля, которые нужны и тренеру, и клиенту (имя, фамилия, аватар, био).
- user_id (PK, FK -> user.user_id): Идентификатор пользователя
- username - псевдоним
- first_name - имя
- last_name - фамилия
- bio - био
- avatar_url - ссылка на аватарку

**Функциональные зависимости:**

{user_id} -> username, first_name, last_name, bio, avatar_url

{username} -> user_id, first_name, last_name, bio, avatar_url

### 3. Профили, являющиеся админами (admin_profile)

Назначение: хранит факт, что пользователь является администратором платформы.
- admin_id (PK, FK -> user.user_id)
- created_at

**Функциональные зависимости:**

{admin_id} -> created_at

### 4. Детали тренера (trainer_details)

Назначение: специфичные для тренера поля (опыт, разряд, образование). Существует только для пользователей с ролью trainer.
- trainer_user_id (PK, FK -> user.user_id): идентификатор тренера
- education_degree: образование
- career_since_date: общий стаж работы

**Функциональные зависимости:**

{trainer_user_id} -> education_degree, career_since_date

### 5. Детали клиента (client_details)

Назначение: специфичные для клиента поля (фитнес-цель). Существует только для пользователей с ролью client.
- client_id (PK, FK -> user.user_id): идентификатор клиента
- fitness_goal: цель

**Функциональные зависимости:**

{client_id} -> fitness_goal

### 6. Справочник видов спорта (sport_type)

Назначение: категории спорта для фильтрации тренеров в каталоге.
- sport_type_id (PK): Идентификатор категории
- name (UK): Название вида спорта

**Функциональные зависимости:**

{sport_type_id} -> name
{name} -> sport_type_id

### 7. Связь тренера и категорий (trainer_to_sport_type)

Назначение: связь многие-ко-многим (какие виды спорта ведёт тренер).
- trainer_id (PK, FK -> user.user_id): Идентификатор тренера
- sport_type_id (PK, FK -> sport_type.sport_type_id): Идентификатор категории
- experience_years: стаж по конкретному виду спорта
- sports_rank: разряд по конкретному виду спорта
PK(trainer_id, sport_type_id)

**Функциональные зависимости:**

{trainer_id, sport_type_id} -> experience_years, sports_rank


### 8. Уровни подписки (subscription_tier)

Назначение: тарифы, которые создаёт тренер. level_rank задаёт порядок уровней доступа и уникален внутри одного тренера.
- subscription_tier_id (PK): Идентификатор тарифа
- trainer_id: Идентификатор тренера-создателя тарифа
- title: Название тарифа
- description: Описание тарифа
- price_currency: Валюта (например, RUB)
- price_value: мантисса
- price_exponent: порядок (по основанию 10)
- level_rank: Ранг уровня доступа
- is_archived: Флаг. Мягкое удаление
- archived_at: Дата и время архивации

**Функциональные зависимости:**

{subscription_tier_id} -> trainer_id, title, description, price_currency, price_value, price_exponent,
level_rank, is_archived, archived_at

{trainer_id, level_rank} -> subscription_tier_id, title, description, price_currency, price_value,
price_exponent, is_archived, archived_at

### 9. Справочник функций платформы (feature_dictionary)

Назначение: словарь "фич", которые могут входить в тариф (например, доступ к чату и т.п.).
- feature_id (PK): Идентификатор функции
- code_name (UK): Кодовое имя функции
- description: Описание функции

**Функциональные зависимости:**

{feature_id} -> code_name, description
{code_name} -> feature_id, description

### 10. Наполнение тарифов (subscription_tier_feature)

Назначение: связь многие-ко-многим "тариф включает фичу".
- subscription_tier_id (PK, FK -> subscription_tier.subscription_tier_id): Идентификатор
тарифа
- feature_id (PK, FK -> feature_dictionary.feature_id): Идентификатор функции

**Функциональные зависимости:**

{subscription_tier_id , feature_id} -> -


### 11. Оформленные подписки (user_subscription)

Назначение: записи о покупке тарифа клиентом и сроке действия.
- user_subscription_id (PK): Идентификатор подписки
- subscriber_user_id (FK -> user.user_id): Идентификатор подписчика
- subscription_tier_id (FK -> subscription_tier.subscription_tier_id): Идентификатор
тарифа
- started_at: Дата и время начала действия подписки
- expires_at: Дата и время окончания действия подписки

**Функциональные зависимости:**

{user_subscription_id } -> subscriber_user_id , subscription_tier_id, started_at, expires_at

### 12. Лента публикаций (post)

Назначение: посты тренера. Доступ регулируется ссылкой на конкретный тариф (min_tier_id), а не числом min_tier_level. Если min_tier_id IS NULL - пост бесплатный.
- post_id (PK): Идентификатор поста
- trainer_id (FK -> user.user_id): Идентификатор тренера (автора)
- min_tier_id (NULLABLE FK -> subscription_tier.subscription_tier_id): Минимальный
тариф (уровень подписки), необходимый для доступа к посту
- title: Заголовок
- text_content: Текстовое содержимое
- created_at: Дата создания
- updated_at: Дата обновления

**Функциональные зависимости:**

{post_id} -> trainer_id, min_tier_id, title, text_content, created_at, updated_at

### 13. Вложения к постам (post_attachment)

Назначение: файлы постов (фото/видео/документы). Вынесено отдельно для 1НФ.
- post_attachment_id (PK): Идентификатор вложения
- post_id (FK -> post.post_id): Идентификатор поста
- file_url: Ссылка на файл
- kind: Тип (image/video/document)

**Функциональные зависимости:**

{post_attachment_id} -> post_id, file_url, kind


### 14. Чат (chat)

Назначение: уникальный чат между конкретным клиентом и тренером. Прочитанность моделируется через "последнее прочитанное сообщение" отдельно для клиента и тренера.
- chat_id (PK): Идентификатор чата
- client_id (FK -> user.user_id): Идентификатор клиента
- trainer_id (FK -> user.user_id): Идентификатор тренера
- client_last_read_message_id (NULL FK -> message.message_id): Идентификатор
последнего прочитанного сообщения клиентом
- trainer_last_read_message_id (NULL FK -> message.message_id): Идентификатор
последнего прочитанного сообщения тренером
- created_at: Дата и время создания чата
UNIQUE(client_id, trainer_id)

**Функциональные зависимости:**

{chat_id} -> client_id, trainer_id, client_last_read_message_id, trainer_last_read_message_id,
created_at

{client_id, trainer_id} -> chat_id, client_last_read_message_id, trainer_last_read_message_id,
created_at

### 15. Сообщения (message)

Назначение: сообщения внутри чата. sender_id указывает, кто отправил сообщение (клиент или тренер). text_content может быть NULL (например, сообщение только с видео), но тогда должно быть вложение.
- message_id (PK): Идентификатор сообщения
- chat_id (FK -> chat.chat_id): Идентификатор чата
- sender_id (FK -> user.user_id): Идентификатор пользователя, который
отправил сообщение
- text_content (NULLABLE): Содержимое сообщения
- created_at, updated_at: Дата и время создания/обновления

**Функциональные зависимости:**

{message_id} -> chat_id, sender_id, text_content, created_at, updated_at

### 16. Вложения к сообщениям (message_attachment)

Назначение: файлы к сообщениям (в т.ч. видео).
- message_attachment_id (PK): Идентификатор вложения
- message_id (FK -> message.message_id): Идентификатор сообщения
- file_url: Ссылка на файл
- kind (image/video/document): Тип файла

**Функциональные зависимости:**

{message_attachment_id} -> message_id, file_url, kind

### 17. Комментарии к постам (comment)

Назначение: текстовые комментарии пользователей к постам.
- comment_id (PK): Идентификатор комментария
- post_id (FK -> post.post_id) - К какому посту комментарий
- author_id (FK -> user.user_id) - Кто написал
- parent_comment_id (NULL FK -> comment.comment_id) - Ответы (ветки
комментариев)
- text_content - TEXT
- created_at: Дата создания
- updated_at: Дата обновления

**Функциональные зависимости:**

{comment_id} -> post_id, author_id, parent_comment_id, text_content, created_at,
updated_at

### 18. Лайки постов (post_like)

Назначение: лайк = факт "пользователь лайкнул пост". Один пользователь может лайкнуть пост максимум один раз.
- post_id (PK, FK -> post.post_id): Идентификатор поста
- user_id (PK, FK -> user.user_id): Идентификатор пользователя
- created_at: Дата создания лайка

**Функциональные зависимости:**

{post_id, user_id} -> created_at

### 19. Пожертвования (donation)

Назначение: добровольное пожертвование от одного пользователя другому.
- donation_id (PK): Идентификатор пожертвования
- sender_user_id (FK -> user.user_id): Кто отправил пожертвование
- recipient_user_id (FK -> user.user_id): Кто получил пожертвование
- amount_mantissa: Мантисса суммы пожертвования, `>= 1`
- amount_scale: Порядок суммы пожертвования по основанию 10, `>= 0`
- currency: Код валюты, например `RUB`
- message: Необязательное сообщение к пожертвованию
- created_at: Дата создания пожертвования
- updated_at: Дата последнего обновления пожертвования

**Функциональные зависимости:**

{donation_id} -> sender_user_id, recipient_user_id, amount_mantissa, amount_scale, currency, message, created_at, updated_at

### 20. Уведомления (notification)

- notification_id (PK): Идентификатор уведомления
- user_id (FK -> user.user_id): Идентификатор пользователя (кому)
- is_read (boolean): Флаг прочитано ли уведомление
- created_at: Дата и время уведомления
- payload(jsonb): Содержимое уведомления
**payload хранит:**
1. type (строка, например "new_comment")
2. ссылки на объекты (post_id, comment_id, from_user_id)
3. данные для UI:
```json
"title": "Рикардо Милос подписался на вас"
"body": "..."
"avatar_url": "..."
```
**Пример:**
new_comment
```json
{
"type": "new_comment",
"post_id": 123,
"comment_id": 555,
"from_user_id": 42,
"title": "Новый комментарий",
"body": "комментарий",
"avatar_url": "..."
}
```
new_message
```json
{

"type": "new_message",
"chat_id": 77,
"message_id": 9001,
"from_user_id": 42,
"title": "Новое сообщение",
"body": "Скинул видео техники, посмотри",
"avatar_url": "..."
}
```

**Функциональные зависимости:**

{notification_id} -> user_id, payload, is_read, created_at

### 20. Сессии авторизации (session)

Назначение: хранит серверные сессии для авторизации на cookie-based sessions. В cookie клиенту выдаётся токен `sid` (HttpOnly), а в БД хранится только его хэш.

- session_id_hash (PK): хэш токена сессии
- user_id (FK -> user.user_id): пользователь, которому принадлежит сессия
- created_at: дата и время создания сессии
- expires_at: дата и время истечения сессии
- last_seen_at: дата и время последней активности
- revoked_at (NULLABLE): дата отзыва сессии (logout); если `NULL`, то сессия не отозвана
- ip (NULLABLE): IP-адрес, с которого создана/используется сессия
- user_agent (NULLABLE): user-agent клиента

**Функциональные зависимости:**

{session_id_hash} -> user_id, created_at, expires_at, last_seen_at, revoked_at, ip, user_agent

## 2. Доказательство нормализации

Схема данных строго соответствует требованиям 1, 2, 3 нормальных форм и Нормальной Форме Бойса-Кодда (НФБК).

### 1НФ (Первая нормальная форма):

Все отношения представлены в виде таблиц, где каждая строка описывает один факт и идентифицируется ключом. В каждой колонке хранится одно значение (скаляр), в строках нет повторяющихся групп атрибутов и многозначных полей. Там, где по смыслу возникает набор значений (например, несколько вложений у поста/сообщения, множество лайков, множество комментариев), он моделируется отдельными строками в связанных таблицах (например, post_attachment, message_attachment, comment, post_like). Исключением является таблица notification, в которой используется поле типа JSONB для хранения содержимого с изменчивой схемой.

### 2НФ (Вторая нормальная форма):

Все отношения находятся в 1НФ. В таблицах с простым первичным ключом частичных зависимостей не бывает по определению (ключ не составной).
**Таблицы с составным первичным ключом в схеме:**
1. trainer_to_sport_type (trainer_id, sport_type_id) - неключевые атрибуты
experience_years и sports_rank зависят от всей пары (trainer_id, sport_type_id), а не от одного trainer_id или одного sport_type_id.

2. subscription_tier_feature (subscription_tier_id, feature_id) - не имеет
неключевых атрибутов, следовательно частичных зависимостей нет.
3. post_like (post_id, user_id) - атрибут created_at зависит от пары (post_id,
user_id). Схема удовлетворяет 2НФ.

### 3НФ (Третья нормальная форма):

Схема находится во 2НФ. Для любой нетривиальной функциональной зависимости X -> A либо X однозначно идентифицирует строку отношения (является ключом), либо A - ключевой атрибут. В приведённых функциональных зависимостях детерминантами выступают ключи отношения (первичные или уникальные), поэтому нарушений 3НФ нет.

### НФБК (Нормальная форма Бойса-Кодда):

Выполняется: в каждом отношении любой детерминант является ключом. user: детерминанты user_id и email (уникален). subscription_tier_feature: детерминанты subscription_tier_id и (trainer_id, level_rank) (уникален внутри тренера). chat: детерминанты chat_id и (client_id, trainer_id) (уникальная пара участников). trainer_to_sport_type: ключ (trainer_id, sport_type_id); неключевые атрибуты (если есть) зависят от ключа целиком. feature_to_subscription_tier: ключ (subscription_tier_id, feature_id), неключевых атрибутов нет. post_like: ключ (post_id, user_id); created_at зависит от ключа.
