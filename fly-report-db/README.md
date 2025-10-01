# БД Обработки полётов

Данные записываются в мастер. Создана потоковая репликация zeusdb.
На реплике данные доступны для чтения.

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

## Настройки

Убедиться в том, что sh-файлы в docker-entrypoint-initdb.d исполняемые, если нет, то выполнить

```
chmod +x master-init.sh
```

## pgAdmin

Расскоментировать нужный блок в docker-compose.
При запуске станет доступен pgAdmin.

Адрес pgAdmin:
http://localhost:5502/
http://127.0.0.1:5502/
http://0.0.0.0:5502/

# БД Мастер

## PSQL

```
sudo docker exec -it pg_fly_report_master_container bash
root@9cbdd4c14ed3:/#
export PGPASSWORD='extR3@der'; psql -U teleuser -d zeusdb
```

## IP адрес БД

```
sudo docker inspect \
 -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' pg_fly_report_master_container
```

# БД Реплика (для отчётов)

## PSQL

```
sudo docker exec -it pg_fly_report_replica_container bash
root@9cbdd4c14ed3:/#
export PGPASSWORD='extR3@der'; psql -U teleuser -d zeusdb
```

## IP адрес БД

```
sudo docker inspect \
 -f '{{range.NetworkSettings.Networks}}{{.IPAddress}}{{end}}' pg_fly_report_replica_container
```
