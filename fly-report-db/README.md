# БД Обработки полётов

Данные записываются в мастер. Создана потоковая репликация zeusdb.
На реплике данные доступны для чтения.

## Таблицы

### Субъекты РФ: regions

| Наименование     | Тип поля | Описание                                  |
| ---------------- | -------- | ----------------------------------------- |
| id               | INT      | Уникальный идентификатор в данной таблице |
| rus_code         | INT      | Код субъекта                              |
| rus_name         | VARCHAR  | Наименование субъекта                     |
| geometry_geojson | JSONB    | Границы субъекта в формате geojson        |
| geometry_postgis | GEOMETRY | Границы субъекта в формате geometry       |

### ОрВД: orvd

| Наименование | Тип поля | Описание                                  |
| ------------ | -------- | ----------------------------------------- |
| id           | INT      | Уникальный идентификатор в данной таблице |
| name         | VARCHAR  | Наименование                              |

### Полёты: flights

| Наименование   | Тип поля  | Описание                                                                             |
| -------------- | --------- | ------------------------------------------------------------------------------------ |
| id             | INT       | Уникальный идентификатор в данной таблице                                            |
| region_id      | INT       | Идентификатор субъекта                                                               |
| orvd_id        | INT       | Идентификатор ОрВД                                                                   |
| coords_postgis | GEOMETRY  | Координата полёта (используются координаты взлёта)                                   |
| event_date     | DATE      | Дата полёта                                                                          |
| xlsxfile_uuid  | UUID      | Уникальный идентификатор исходного xlsx файла                                        |
| row_num        | INT       | Строка xlsx файла                                                                    |
| details        | JSONB     | Набор полей ORVD, SHR, DEP, ARR в формате json                                       |
| details_SHA256 | VARCHAR   | Контрольная сумма содержания набора полей ORVD, SHR, DEP, ARR телеграфного сообщения |
| createdAt      | TIMESTAMP | Дата загрузки файла в систему                                                        |
| updatedAt      | TIMESTAMP | Дата обновления информации                                                           |

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

# \*БД Реплика (для отчётов)

\* БД для целевой архитектуры, при масштабировании системы (в существующей версии не испоьзуются)

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
