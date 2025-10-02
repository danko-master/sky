# Информационная система обработки и статистики полётов

## Запуск всех компонентов одновременно

Переименовать файл **docker-compose.yml.example** в **docker-compose.yml**
далее выполнить:

```
sudo docker-compose up
```

Для запуска в фоновом режиме:

```
sudo docker-compose up -d
```

Удалить все компоненты (Network, Volume, Container):

```
sudo docker-compose down -v
```

## Запуск отдельных компонентов

В каждой папке содержится отдельный модуль системы, который содержит свои команды для локального запуска, если это необходимо.
Модули, которые содержат **docker-compose.yml.example** неообходимо переименовать в **docker-compose.yml**

Для пользователя и администратора предусмотрено свое автоматизированное рабочее место (АРМ).

## Документация

Описание архитектуры информационной системы, таблицы, состав полей:
https://github.com/danko-master/sky/tree/main/archi

АРМ вынесен в подмодуль:
https://github.com/danko-master/fly-app

БД для АРМа:
https://github.com/danko-master/sky/tree/main/fly-db

БД статистических данных полётов:
https://github.com/danko-master/sky/tree/main/fly-report-db

Обработчик файлов:
https://github.com/danko-master/sky/tree/main/worker-file-reader

Обработчик статусов:
https://github.com/danko-master/sky/tree/main/worker-status

Брокер Kafka:
https://github.com/danko-master/sky/tree/main/kafka

Файловое хранилище MinIO:
https://github.com/danko-master/sky/tree/main/minio
