# USDT Rates Service

GRPC сервис для получения курса USDT с биржи Grinex с сохранением в PostgreSQL базе данных.

## Описание

Сервис реализует:
- GRPC метод `GetRates` для получения курса USDT (ask и bid цены + метка времени)
- Автоматическое сохранение курса в PostgreSQL при каждом вызове `GetRates`
- GRPC метод `Healthcheck` для проверки работоспособности сервиса
- Graceful shutdown
- Мониторинг с помощью Prometheus
- Трассировка с помощью OpenTelemetry
- Структурированное логирование с помощью zap

## Требования

- Go 1.23+
- PostgreSQL 15+
- Docker и Docker Compose
- Protocol Buffers compiler (protoc)

## Быстрый старт

1Соберите приложение:
```bash
make build
```

2Запустите с помощью Docker Compose:
```bash
docker-compose up -d
```

3Запустите приложение:
```bash
docker-compose run --rm app ./app
```

## Makefile команды

- `make build` - сборка приложения
- `make test` - запуск unit-тестов
- `make docker-build` - сборка Docker-образа
- `make run` - запуск приложения локально
- `make lint` - запуск линтера golangci-lint
- `make generate` - генерация protobuf файлов
- `make docker-run` - запуск сервисов через docker-compose
- `make docker-stop` - остановка docker-compose сервисов
- `make help` - список всех доступных команд

## Конфигурация

Приложение поддерживает конфигурацию через флаги командной строки и переменные окружения.

### Переменные окружения

Все переменные окружения имеют префикс `USDT_`:

#### База данных
- `USDT_DATABASE_HOST` - хост базы данных (по умолчанию: `localhost`)
- `USDT_DATABASE_PORT` - порт базы данных (по умолчанию: `5432`)
- `USDT_DATABASE_USER` - пользователь БД (по умолчанию: `postgres`)
- `USDT_DATABASE_PASSWORD` - пароль БД (по умолчанию: `postgres`)
- `USDT_DATABASE_DATABASE` - имя базы данных (по умолчанию: `usdt_rates`)
- `USDT_DATABASE_SSL_MODE` - режим SSL (по умолчанию: `disable`)
- `USDT_DATABASE_MAX_OPEN_CONNS` - максимальное количество открытых соединений (по умолчанию: `25`)
- `USDT_DATABASE_MAX_IDLE_CONNS` - максимальное количество idle соединений (по умолчанию: `25`)
- `USDT_DATABASE_CONN_MAX_LIFETIME` - время жизни соединения (по умолчанию: `5m`)

#### Сервер
- `USDT_SERVER_PORT` - порт GRPC сервера (по умолчанию: `8080`)
- `USDT_SERVER_GRACEFUL_TIMEOUT` - таймаут graceful shutdown (по умолчанию: `30s`)
- `USDT_SERVER_READ_TIMEOUT` - таймаут чтения (по умолчанию: `10s`)
- `USDT_SERVER_WRITE_TIMEOUT` - таймаут записи (по умолчанию: `10s`)
- `USDT_SERVER_MAX_CONNECTION_IDLE` - время idle соединения (по умолчанию: `2m`)

#### Grinex API
- `USDT_GRINEX_BASE_URL` - базовый URL API (по умолчанию: `https://grinex.io`)
- `USDT_GRINEX_TIMEOUT` - таймаут запросов к API (по умолчанию: `10s`)
- `USDT_GRINEX_MARKET` - торговая пара (по умолчанию: `usdtrub`)

#### Логирование
- `USDT_LOGGING_LEVEL` - уровень логирования: `debug`, `info`, `warn`, `error` (по умолчанию: `info`)
- `USDT_LOGGING_FORMAT` - формат логов: `json`, `console` (по умолчанию: `json`)

#### Трассировка
- `USDT_TRACING_ENABLED` - включить трассировку (по умолчанию: `false`)
- `USDT_TRACING_JAEGER_URL` - URL Jaeger коллектора (по умолчанию: `http://localhost:14268/api/traces`)
- `USDT_TRACING_SERVICE_NAME` - имя сервиса для трассировки (по умолчанию: `usdt-rates-service`)

#### Метрики
- `USDT_METRICS_ENABLED` - включить метрики (по умолчанию: `true`)
- `USDT_METRICS_PATH` - путь для метрик (по умолчанию: `/metrics`)
- `USDT_METRICS_PORT` - порт сервера метрик (по умолчанию: `9090`)

### Флаги командной строки

Все переменные окружения имеют соответствующие флаги командной строки. Например:
- `--database.host` соответствует `USDT_DATABASE_HOST`
- `--server.port` соответствует `USDT_SERVER_PORT`
- `--grinex.base-url` соответствует `USDT_GRINEX_BASE_URL`

Полный список флагов можно получить командой:
```bash
./app --help
```

## API

### GRPC сервис

Сервис предоставляет следующие методы:

#### GetRates
Получение текущих курсов валют.

**Запрос:**
```protobuf
message GetRatesRequest {
  string market = 1; // Торговая пара, например "usdtrub"
}
```

**Ответ:**
```protobuf
message GetRatesResponse {
  string ask = 1;                        // Цена продажи
  string bid = 2;                        // Цена покупки
  google.protobuf.Timestamp timestamp = 3; // Время получения курса
  string market = 4;                     // Торговая пара
}
```

#### Healthcheck
Проверка состояния сервиса.

**Запрос:**
```protobuf
message HealthcheckRequest {}
```

**Ответ:**
```protobuf
message HealthcheckResponse {
  string status = 1;                     // Статус: "healthy" или "unhealthy"
  string version = 2;                    // Версия сервиса
  google.protobuf.Timestamp timestamp = 3; // Время проверки
}
```

### Примеры использования

#### grpcurl
```bash
# Получить курсы
grpcurl -plaintext -d '{"market":"usdtrub"}' localhost:8080 rates.RatesService/GetRates

# Проверить здоровье сервиса
grpcurl -plaintext localhost:8080 rates.RatesService/Healthcheck
```

#### Метрики Prometheus
Метрики доступны по адресу `http://localhost:9090/metrics`

## Архитектура

### Структура проекта
```
TestForWork/
├── cmd/server/           # Главное приложение
├── internal/
│   ├── api/grpc/        # GRPC сервер и хэндлеры
│   ├── client/          # HTTP клиент для Grinex API
│   ├── config/          # Управление конфигурацией
│   ├── service/         # Бизнес-логика
│   └── storage/
│       ├── postgres/    # PostgreSQL репозиторий
│       └── migrations/  # Миграции базы данных
├── pkg/logger/          # Логирование
├── proto/rates/         # Protobuf определения и сгенерированные файлы
├── tests/               # Unit-тесты
├── Dockerfile
├── docker-compose.yml
├── Makefile
└── README.md
```

### Компоненты

1. **GRPC сервер** - обрабатывает входящие запросы
2. **Сервисный слой** - содержит бизнес-логику
3. **HTTP клиент** - взаимодействие с API Grinex
4. **PostgreSQL репозиторий** - хранение данных
5. **Конфигурационный слой** - управление настройками

### База данных

#### Схема таблицы rates
```sql
CREATE TABLE rates (
    id SERIAL PRIMARY KEY,
    market VARCHAR(20) NOT NULL,
    ask DECIMAL(20, 8) NOT NULL,
    bid DECIMAL(20, 8) NOT NULL,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_rates_market ON rates(market);
CREATE INDEX idx_rates_timestamp ON rates(timestamp);
CREATE INDEX idx_rates_created_at ON rates(created_at);
```

## Разработка

### Установка инструментов разработки
```bash
make install-tools
```

### Генерация protobuf файлов
```bash
make generate
```

### Запуск тестов
```bash
make test
```

### Запуск линтера
```bash
make lint
```

### Форматирование кода
```bash
make fmt
```

## Мониторинг и наблюдаемость

### Логирование
Сервис использует структурированное логирование с помощью zap. Поддерживаются форматы JSON и console.

### Метрики
Доступны стандартные метрики gRPC:
- Количество запросов
- Длительность запросов
- Коды ошибок

### Трассировка
Поддерживается трассировка с помощью OpenTelemetry и экспорт в Jaeger.

## Troubleshooting

### Проблемы с подключением к базе данных
1. Убедитесь, что PostgreSQL запущен
2. Проверьте параметры подключения
3. Убедитесь, что база данных создана

### Проблемы с Grinex API
1. Проверьте доступность API: `curl https://grinex.io/api/v2/depth?market=usdtrub`
2. Убедитесь в корректности URL и таймаутов

### Проблемы с Docker
1. Убедитесь, что Docker и Docker Compose установлены
2. Проверьте права доступа к Docker socket
3. Убедитесь в наличии свободных портов 8080 и 9090