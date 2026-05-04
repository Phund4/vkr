# Trace Data Service

Минимальный скелет сервиса для работы с данными трейсов, созданный на основе trace-pipelines-service.

## Описание

Сервис предоставляет базовые gRPC и HTTP API endpoints для health checks и готовности к работе. Содержит минимальную функциональность для деплоя и дальнейшего развития.

## API Endpoints

### HTTP API
- `GET /health` - Health check
- `GET /probe/ready` - Готовность сервиса  
- `GET /probe/health` - Здоровье сервиса

### gRPC API
- `ProbeService.Ready` - Готовность сервиса
- `ProbeService.Health` - Здоровье сервиса

## Структура проекта

```
.
├── cmd/                           # Точка входа приложения
│   └── main.go
├── internal/
│   ├── app/                       # Основная логика приложения
│   │   ├── app.go                 # Структура App
│   │   ├── deps.go                # Зависимости
│   │   ├── handlers.go            # gRPC handlers  
│   │   └── run.go                 # Логика запуска
│   ├── config/                    # Конфигурация
│   │   └── config.go
│   └── adapters/
│       └── grpc/                  # gRPC адаптеры
│           ├── probe_handler.go   # Probe endpoints
│           └── interceptor.go     # Logging interceptor
├── api/proto/                     # Proto файлы
│   ├── probe.proto
│   └── trace.proto
├── pkg/                          
│   ├── api/proto/                 # Сгенерированные Go файлы
│   └── svcinfo/                   # Информация о версии
├── docker/                        # Docker конфигурация
│   └── Dockerfile
├── deploy/                        # Конфигурация деплоя
│   └── common/
│       └── base-values.yml
├── third_party/                   # Third party proto файлы
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Локальная разработка

### Требования

- Go 1.23.3+
- Docker
- protoc (для генерации proto файлов)

### Установка и запуск

```bash
# Скачать зависимости
go mod tidy

# Установить инструменты для proto (если нужно)
make install-tools

# Сгенерировать proto файлы (если изменились)
make generate-proto

# Собрать приложение
make build

# Запустить локально
make run
```

### Переменные окружения

- `ENV` - окружение (local/prod), по умолчанию "local"
- `LISTENER_NETWORK` - сеть для gRPC сервера, по умолчанию "tcp"
- `LISTENER_PORT` - порт gRPC сервера, по умолчанию ":8001"
- `HTTP_SERVER_PORT` - порт HTTP сервера, по умолчанию ":8000"

### Проверка работы

```bash
# Health check
curl http://localhost:8000/health

# Probe endpoints
curl http://localhost:8000/probe/ready
curl http://localhost:8000/probe/health
```

## Сборка и деплой

### Сборка

```bash
# Локальная сборка
make build

# Docker образ
make docker-build
```

### Деплой

Конфигурация деплоя находится в `deploy/common/base-values.yml`.

Основные настройки:
- Реплик: 1
- Ресурсы: 6Gi RAM, 3000m CPU (запрос), 15Gi RAM, 3000m CPU (лимит)
- Порты: 8000 (HTTP), 8081 (метрики)
- Health check: `/health`

## Команды Makefile

```bash
make help          # Показать все доступные команды
make build         # Собрать приложение
make run           # Запустить приложение
make test          # Запустить тесты
make clean         # Очистить временные файлы
make install-tools # Установить инструменты для proto
make generate-proto # Сгенерировать Go код из proto
make docker-build  # Собрать Docker образ
make docker-run    # Запустить Docker контейнер
```

## Статус

✅ **ГОТОВ К ДЕПЛОЮ**

Минимальный скелет сервиса готов:
- Структура приложения создана
- Proto файлы настроены
- Health checks работают
- Docker и Kubernetes конфигурация готова
- Все команды сборки и запуска работают

Можно добавлять бизнес-логику и деплоить!
