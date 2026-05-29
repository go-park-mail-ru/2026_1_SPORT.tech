# SportTech

> Backend — Go-микросервисы платформы **SPORT.tech**

## Команда
Верясов Михаил - backend
Клубков Максим - frontend

## Ссылки

- **Прод:** https://sporteon.ru
- **Swagger / API:** https://sporteon.ru/api/docs
- **Figma:** [тык](https://www.figma.com/design/cjFfrkwiX09KlK3CrhNOQR/Sporteon?node-id=0-1&p=f&t=KMaMwwt8V0ER43oc-0)
- **Frontend-репозиторий:** [тык](https://github.com/frontend-park-mail-ru/2026_1_SPORT.tech)

## О проекте

Бэкенд платформы спортивного контента **SPORT.tech**. Четыре Go-сервиса
общаются между собой по gRPC, наружу выставлен HTTP REST через grpc-gateway.

```
services/
  api-gateway/   — единственный публичный сервис: HTTP REST → gRPC (порт 8080)
  auth/          — аутентификация и сессии (gRPC :9091, HTTP :8081)
  profile/       — профили, аватары, виды спорта (gRPC :9092, HTTP :8082)
  content/       — посты, подписки, донаты, платежи (gRPC :9093, HTTP :8083)
grpc/
  proto/         — исходные .proto
  gen/           — сгенерированный код (НЕ редактировать руками)
```

## Быстрый старт

```bash
make compose-up      # собрать и поднять весь стек (нужен .env)
make compose-down
```

## Разработка

```bash
make lint            # gofmt + go vet
make test            # go test ./...
make coverage-check  # минимум 60% покрытия (запускается в CI)
make ci              # lint + test + coverage-check
make generate        # перегенерировать proto + easyjson
```
