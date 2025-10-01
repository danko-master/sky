# Kafka для проекта

## Запуск

```
sudo docker-compose up
```

В фоновом режиме

```
sudo docker-compose up -d
```

Удалить все компоненты (Network, Volume, Container):

```
sudo docker-compose down -v
```

## Kafka UI

http://localhost:8090/

При запуске автоматически создается кластер с наименованием:
local-kafka-cluster

Кластер также возможно добавить по адресу:
http://localhost:8090/ui/clusters/create-new-cluster

## Настройки для внешних приложений

Для внешних приложений использовать следующие параметры.

Пример переменных окружения для приложения, которое будет взаимодействовать с брокером.

Для отправки сообщения:

```
KAFKA_BROKER=localhost:9092
KAFKA_TOPIC=topic-fly-status
```

Для чтения сообщений:

```
KAFKA_BROKER=localhost:9092
KAFKA_TOPIC=topic-fly-status
KAFKA_GROUP=topic-group-fly-status
```

## Топики

| Наименование           | Описание                                   | Формат                                                                   |
| ---------------------- | ------------------------------------------ | ------------------------------------------------------------------------ |
| topic-fly-status       | Передача статусов                          | { uuid: UUID, status: statusCode }                                       |
| topic-fly-file         | Передача наименований файлов для обработки | { uuid: UUID }                                                           |
| \* topic-fly-file-line | Передача строк файла для обработки         | { uuid: UUID, rownum: rownum, orvd: orvd, shr: shr, dep: dep, arr: arr } |

\* топик для целевой архитектуры, при масштабировании системы (в существующей версии не испоьзуется)

uuid - сквозной идентификатор (см. поле таблицы xlsxfiles в fly-db)
statusCode - статусы xlsxfiles (см. fly-db)
rownum - порядковый номер строки в xlsx
orvd, shr, dep, arr - колонки в xlsx

## IP адрес

```
sudo docker inspect \
 -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' fly-line-kafka
```
