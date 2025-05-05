# Распределённый калькулятор

> Сервис для асинхронного вычисления математических выражений в распределённой среде с поддержкой многопользовательского режима и персистентностью.

## 📁 Структура проекта

- `cmd/orchestrator` — HTTP и gRPC сервис (Оркестратор)
- `cmd/agent` — gRPC клиент (Агент), выполняющий вычислительные задачи
- `internal/database` — работа с SQLite (модели, миграции, CRUD)
- `internal/orchestrator` — логика HTTP-обработчиков, парсер выражений, планировщик задач, gRPC сервер
- `internal/agent` — gRPC-воркер, выполняющий вычисления
- `pkg/grpc/calculator` — protobuf-описание и сгенерированный код

## ⚙️ Требования

- Go 1.24 или выше
- SQLite3
- protoc + protoc-gen-go (при необходимости перекомпиляции `.proto`)
- (опционально) `jq` для удобного форматирования JSON в примерах

## 🚀 Быстрый старт

1. Клонируем репозиторий:
   ```bash
   git clone https://github.com/<ваш_логин>/appCalc.git
   cd appCalc
   ```

2. Устанавливаем зависимости:
   ```bash
   go mod tidy
   ```

3. Настраиваем переменные окружения:
   ```bash
   export JWT_SECRET="your_jwt_secret_here"
   export COMPUTING_POWER=4      # (опционально) число воркеров у агента, по умолчанию 1
   ```

4. Запускаем Оркестратор:
   ```bash
   go run ./cmd/orchestrator
   ```

5. В новом терминале запускаем Агентов (можно несколько экземпляров):
   ```bash
   go run ./cmd/agent
   ```

6. Открываем приложение в браузере по адресу:
   ```
   http://localhost:8080
   ```

> Убедитесь, что статика (`index.html`) лежит в `web/static/index.html` или в корне, как указано в коде.

## 📡 API HTTP (Оркестратор)

Базовый URL: `http://localhost:8080/api/v1`

### 1. Регистрация пользователя

- **POST** `/register`
  ```bash
  curl -s -X POST http://localhost:8080/api/v1/register \
    -H "Content-Type: application/json" \
    -d '{"login":"user1","password":"pass123"}'
  ```

- **Коды ответа**:
  - `201 Created` — успешно
  - `400 Bad Request` — неверный формат или пустые поля
  - `409 Conflict` — логин уже занят

### 2. Вход (JWT)

- **POST** `/login`
  ```bash
  curl -s -X POST http://localhost:8080/api/v1/login \
    -H "Content-Type: application/json" \
    -d '{"login":"user1","password":"pass123"}'
  ```

- **Коды ответа**:
  - `200 OK` и JSON:
    ```json
    { "token": "<JWT_TOKEN>" }
    ```
  - `400 Bad Request` — неверный формат
  - `401 Unauthorized` — неверные логин/пароль

### 3. Отправка выражения на вычисление

- **POST** `/calculate`
  ```bash
  curl -s -X POST http://localhost:8080/api/v1/calculate \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <JWT_TOKEN>" \
    -d '{"expression":"(2+3)*4"}'
  ```

- **Коды ответа**:
  - `201 Created` и JSON:
    ```json
    {
      "id": 1,
      "expression": "(2+3)*4",
      "status": "pending"
    }
    ```
  - `400 Bad Request` — пустое или некорректное выражение
  - `401 Unauthorized` — отсутствует или неверный токен

### 4. Получение статуса и результата

- **GET** `/expressions` — список всех ваших выражений
- **GET** `/expressions/<id>` — конкретное выражение по ID

```bash
curl -s -X GET http://localhost:8080/api/v1/expressions/<id> \
  -H "Authorization: Bearer <JWT_TOKEN>"
```

- **Коды ответа**:
  - `200 OK` и JSON-объект/массив:
    ```json
    [
      {
        "id": 1,
        "user_id": 1,
        "expression": "(2+3)*4",
        "status": "done",
        "result": 20,
        "steps": ["Result: 5","Result: 20"],
        "created_at": "...",
        "updated_at": "..."
      }
    ]
    ```

## 🧪 Тестирование

- Запуск всех тестов:
  ```bash
  go test ./...
  ```

- Юнит-тесты:
  ```bash
  go test ./internal/orchestrator/parser.go
  go test ./internal/agent/worker.go
  go test ./internal/database
  ```

- Интеграционные тесты:
  ```bash
  go test ./internal/orchestrator
  go test ./internal/orchestrator/grpc_server_test.go
  ```

## 🔧 Генерация protobuf

Если нужно обновить код из `.proto`:

```bash
protoc --go_out=. --go-grpc_out=. pkg/grpc/calculator/calculator.proto
```
---