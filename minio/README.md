# Файловое хранилище

## Настройки и доступ

Администратор:
minioadmin

Пароль администратора:
minioadmin

### Порты

MinIO API port: 9000
Console port: 9001

Пример:
API: http://172.26.0.2:9000
http://127.0.0.1:9000

WebUI: http://172.26.0.2:9001
http://127.0.0.1:9001

### Bucket

Наименование автоматически созданного bucket:
fly-telegraph

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

## IP адрес

```
sudo docker inspect \
 -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' minio
```
