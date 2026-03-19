# URL Shortener (Ozon Testcase)

Сервис сокращения ссылок на Go с двумя реализациями хранилища (`postgres` и `in-memory`), HTTP API на `gin` и cron-job

## Возможности

- Создание короткой ссылки для оригинального URL.
- Получение оригинального URL по короткому коду (JSON).
- Редирект по короткому коду на оригинальный URL.
- Поддержка двух storage-режимов:
  - `postgres`
  - `memory`
- Ежедневная cleanup-задача (по умолчанию в `03:00`), удаляющая старые ссылки(по умолчанию те, что не использовались 3 года).

## Запуск
### Создайте .env в корне проекта и скопируйте туда .env/example
### Запустить Docker 
### Запустить команду в корне проекта 

```bash
docker compose up --build
```

## Команды Makefile

```bash
make test
make test-integration
make db-up
make docker-up
make docker-up-d
make docker-down
make docker-reset
make logs
```

- `make test` — запуск всех unit-тестов проекта.
- `make test-integration` — запуск integration-тестов репозитория с тегом `integration`. (требуется поднятая бд)
- `make db-up` — поднять только контейнер Postgres в фоне.
- `make docker-up` — собрать и запустить все сервисы `docker compose` в foreground.
- `make docker-up-d` — собрать и запустить все сервисы `docker compose` в фоне.
- `make docker-down` — остановить и удалить контейнеры/сеть (без удаления volume).
- `make docker-reset` — остановить и удалить контейнеры/сеть/volume (полный сброс данных БД).
- `make logs` — смотреть логи `app` и `db` в режиме follow.


## API

### 1) Создать короткую ссылку

`POST /shrt/links`

` POST http://localhost:8080/shrt/links`

Request:

```json
{
  "url": "https://example.com/path"
}
```

Response `201`:

```json
{
  "short_url": "http://localhost:8080/shrt/Abc123_abc"
}
```

### 2) Получить оригинальный URL по коду (JSON)

`GET /shrt/links/:code`

` GET http://localhost:8080/shrt/links/:code`

пример : `GET /shrt/links/Abc123_abc`

Response `200`:

```json
{
  "original_url": "https://example.com/path"
}
```

### 3) Редирект по коду

`GET /shrt/:code`

` GET http://localhost:8080/shrt/:code`

пример : `GET /shrt/Abc123_abc`

Response:
- `302 Found` + `Location: <original_url>`

## Архитектура и слои

Применена классическая слоистая архитектура


Ключевое архитектурное решение: сервис работает через интерфейс `URLRepository`, поэтому storage можно переключать параметром запуска без изменения бизнес-логики.

## Алгоритм генерации короткой ссылки

1. Вход: нормализованный URL + `nonce`.
2. Формируется строка: `normalizedURL + "|" + nonce`.
3. Считается `SHA-256`.
4. Полный хеш переводится в число (`big.Int`).
5. Число кодируется в `base63` алфавит:
   - `a-z`, `A-Z`, `0-9`, `_`
6. Берутся 10 символов (`ShortCodeLength = 10`).

Свойства:
- детерминированность для одинакового `URL + nonce`;
- ограниченный алфавит и фиксированная длина;
- при коллизии выполняются ретраи (`maxCreateAttempts = 10`).
- nonce всегда начинается с 0 и каждый ретрай увеличивается на 1 

## Почему ссылки уникальны

- `original_url` имеет `UNIQUE` ограничение.
- `short_url` имеет `UNIQUE` ограничение.
- При гонках и конфликте вставки сервис корректно обрабатывает `23505`:
  - для `original_url` -> возвращает уже существующую ссылку;
  - для `short_url` -> пробует следующий `nonce`.

## Cleanup (крон-задача)

Реализована через [`go-co-op/gocron`](https://github.com/go-co-op/gocron).

- Запуск по расписанию: ежедневно.
- Время по умолчанию: `03:00`.
- Удаляются ссылки:
  - с `last_used_at < cutoff`,
  - либо (если `last_used_at IS NULL`) с `created_at < cutoff`.
- Удаление выполняется батчами через `DeleteExpiredBatch`.

Настраивается через env:
- `CLEANUP_BATCH_SIZE`
- `CLEANUP_RETENTION_YEARS`
- `CLEANUP_HOUR`
- `CLEANUP_MINUTE`

## Стек

- Go 1.25
- HTTP: `gin`
- DB: PostgreSQL 18
- DB driver/pool: `pgx/v5` + `pgxpool`
- Scheduler: `gocron/v2`
- Logging: `zap`
- Tests: стандартный `testing` + `testify`
- Контейнеризация: `Docker`, `docker-compose`

## Конфигурация (env)

См. `.env.example`.

Основные параметры:

- `HTTP_ADDR` — адрес сервера, например `:8080`
- `BASE_URL` — публичный base URL для формирования `short_url`
- `STORAGE` — указывайте `postgres` или `memory` в зависимости от ваших потребностей


## Тесты

### Unit тесты

```bash
go test ./...
```

Покрываются:
- генерация short code;
- service-логика;
- handler-слой;
- domain-валидации;
- worker/cleanup.

### Integration тесты (Postgres repository)

```bash
go test -tags=integration ./internal/repository -v
```

Нужна доступная Postgres (через `DB_*` или `DATABASE_URL` в окружении теста).

перед интеграционными тестами поднять бд 
```bash
docker compose up -d db
```

## Нагрузочное тестирование

### Ознакомиться с результатами можно открыв два png файла в корне репозитория
Профиль коллекции:
- только `POST /shrt/links` и `GET /shrt/links/:code` (без редиректа),
- соотношение операций: `70% GET / 30% POST`,
- входные URL для POST сделаны максимально уникальными.