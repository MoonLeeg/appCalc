# Calculator Service

Проект реализует систему для асинхронного вычисления математических выражений с авторизацией через JWT и хранением данных в SQLite. Состав проекта:

- `cmd/orchestrator` — HTTP и gRPC сервис (Оркестратор)
- `cmd/agent` — gRPC клиент (Агент), выполняющий вычислительные задачи
- `internal/database` — работа с SQLite (модели, миграции)
- `internal/orchestrator` — логика HTTP-гайндлеров, парсера, планировщика задач и gRPC сервера
- `internal/agent` — gRPC-воркер, выполняющий задачи
- `pkg/grpc/calculator` — protobuf определение и сгенерированный код

## Требования

- Go 1.24+
- SQLite3
- protoc (если нужно перекомпилировать `.proto`)

## Установка и запуск

1. Клонировать репозиторий:
   ```bash
   git clone https://github.com/yourusername/calculator.git
   cd calculator
   ```

2. Установить зависимости:
   ```bash
   go mod tidy
   ```

3. Запустить Оркестратор:
   ```bash
   go run ./cmd/orchestrator
   ```

4. Запустить одного или нескольких Агентов:
   ```bash
   go run ./cmd/agent
   # для увеличения числа воркеров:
   COMPUTING_POWER=4 go run ./cmd/agent
   ```

## API HTTP (Оркестратор)

### Регистрация пользователя

```bash
curl -i -X POST http://localhost:8080/api/v1/register \
  -H 'Content-Type: application/json' \
  -d '{"login":"user1","password":"pass123"}'
```

- Успех: `201 Created`
- Если логин уже существует: `409 Conflict`

### Вход (JWT)

```bash
curl -i -X POST http://localhost:8080/api/v1/login \
  -H 'Content-Type: application/json' \
  -d '{"login":"user1","password":"pass123"}'
```

- Успех: `200 OK` и JSON `{ "token": "..." }`
- Ошибка авторизации: `401 Unauthorized`

### Вычисление выражения

```bash
curl -i -X POST http://localhost:8080/api/v1/calculate \
  -H 'Content-Type: application/json' \
  -H 'Authorization: Bearer <JWT>' \
  -d '{"expression":"2+2*2"}'
```

- Успех: `201 Created`, возвращает `id` задачи и первоначальный статус

### Получение статуса/результата

```bash
curl -i -X GET http://localhost:8080/api/v1/expressions/<id> \
  -H 'Authorization: Bearer <JWT>'
```

- Успех: `200 OK`, возвращает JSON с полями `status`, `result`, `steps`

## Примеры использования

1. Регистрация и вход:

    ```bash
    curl -s -X POST http://localhost:8080/api/v1/register \
      -H 'Content-Type: application/json' \
      -d '{"login":"demo","password":"secret"}'

    curl -s -X POST http://localhost:8080/api/v1/login \
      -H 'Content-Type: application/json' \
      -d '{"login":"demo","password":"secret"}' \
      | jq -r '.token'
    ```

2. Вычисление:

    ```bash
    TOKEN=$(curl -s -X POST http://localhost:8080/api/v1/login \
      -H 'Content-Type: application/json' \
      -d '{"login":"demo","password":"secret"}' \
      | jq -r '.token')

    curl -s -X POST http://localhost:8080/api/v1/calculate \
      -H 'Content-Type: application/json' \
      -H "Authorization: Bearer $TOKEN" \
      -d '{"expression":"(2+3)*4"}' \
      | jq

    curl -s -X GET http://localhost:8080/api/v1/expressions/1 \
      -H "Authorization: Bearer $TOKEN" \
      | jq
    ```

## Тесты

- Unit-тесты: `go test ./internal/orchestrator` (парсер, handlers)
- Интеграционные тесты: `go test ./internal/orchestrator` (API)

```bash
go test ./...
```