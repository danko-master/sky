package main

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/lpernett/godotenv"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/segmentio/kafka-go"
	"github.com/xuri/excelize/v2"
)

// init is invoked before main()
func init() {
	fmt.Println("Run init")
	// loads values from .env into the system
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}
	godotenv.Load(".env." + env)
}

// Приоритет внешним переменным окружения
func env(name string) string {
	extEnv := os.Getenv(name)
	if extEnv == "" {
		newEnv, _ := os.LookupEnv(name)
		return newEnv
	} else {
		return extEnv
	}
}

func main() {
	fmt.Println("Run worker file reader")
	KAFKA_BROKER := env("KAFKA_BROKER")
	KAFKA_TOPIC := env("KAFKA_TOPIC")
	KAFKA_GROUP := env("KAFKA_GROUP")
	KAFKA_TOPIC_STATUS := env("KAFKA_TOPIC_STATUS")

	log.Println("kafka.NewReader")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{KAFKA_BROKER}, // Replace with your Kafka broker addresses
		Topic:   KAFKA_TOPIC,
		GroupID: KAFKA_GROUP,
		// MinBytes:  10e3, // 10KB
		// MaxBytes:  10e6, // 10MB
		Partition: 0, // Or use GroupID for consumer group management across partitions
	})

	log.Println("kafka.NewWriter")
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  []string{KAFKA_BROKER}, // Replace with your Kafka broker addresses
		Topic:    KAFKA_TOPIC_STATUS,
		Balancer: &kafka.LeastBytes{},
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Initialize reader and db as shown above
	fmt.Println("Connect to DB ...")
	DB_HOST := env("DB_HOST")
	DB_PORT := env("DB_PORT")
	DB_USER := env("DB_USER")
	DB_PASSWORD := env("DB_PASSWORD")
	DB_NAME := env("DB_NAME")

	// "postgres://username:password@localhost:5432/database_name"
	DATABASE_URL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		DB_USER, DB_PASSWORD, DB_HOST, DB_PORT, DB_NAME)
	conn, err := pgx.Connect(ctx, DATABASE_URL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		log.Fatal(err)
		os.Exit(1)
	} else {
		log.Printf("Successful connection to the database %v", DB_NAME)
	}

	// Initialize reader and writer as shown above

	go worker(ctx, reader, writer, conn)

	// Keep the main goroutine alive, e.g., by waiting for a signal
	select {}
}

func worker(ctx context.Context, reader *kafka.Reader, writer *kafka.Writer, conn *pgx.Conn) {
	log.Println("go worker")

	for {
		select {
		case <-ctx.Done():
			log.Println("Worker shutting down.")
			return
		default:
			// Read message
			log.Println("Read message")

			m, err := reader.FetchMessage(ctx)
			if err != nil {
				log.Printf("Error reading message: %v", err)
				continue
			}

			log.Printf("Received message from topic %s, partition %d, offset %d: %s\n",
				m.Topic, m.Partition, m.Offset, string(m.Value))

			var data BrokerMessageStruct
			errJson := json.Unmarshal(m.Value, &data)
			if errJson != nil {
				log.Fatal("Fail unmarshaling JSON:", errJson)
			}
			uuid := data.UUID
			log.Println(uuid)

			// Status file reader
			msgStatus := BrokerMessageStatusStruct{
				UUID:       uuid,
				StatusCode: "reading",
			}
			sendMsgStatus(ctx, writer, &msgStatus)

			// MinIO
			endpoint := env("MINIO_ENDPOINT") + ":" + env("MINIO_PORT")
			accessKeyID := env("MINIO_ROOT_USER")
			secretAccessKey := env("MINIO_ROOT_PASSWORD")
			useSSL := false
			bucketName := env("MINIO_BUCKET_NAME")

			minioClient, errMinio := minio.New(endpoint, &minio.Options{
				Creds:  credentials.NewStaticV4(accessKeyID, secretAccessKey, ""),
				Secure: useSSL,
			})
			if errMinio != nil {
				log.Fatalln(errMinio)
			}
			object, errMinioRead := minioClient.GetObject(context.Background(), bucketName, uuid, minio.GetObjectOptions{})
			if errMinioRead != nil {
				log.Fatal("Failed to get object:", errMinioRead)
			}
			defer object.Close()

			f, err := excelize.OpenReader(object)
			if err != nil {
				log.Fatalln(err)
			}
			defer func() {
				if err := f.Close(); err != nil {
					log.Fatalln(err)
				}
			}()

			// Get the list of all sheet names
			sheetList := f.GetSheetList()
			if len(sheetList) == 0 {
				log.Println("No sheets found in the Excel file.")
				return
			}

			firstSheetName := sheetList[0]
			fmt.Printf("The first sheet name is: '%s'\n", firstSheetName)

			// Get all the rows in the first sheet
			rows, err := f.GetRows(firstSheetName)
			if err != nil {
				log.Fatalln(err)
			}

			// Status
			msgStatus = BrokerMessageStatusStruct{
				UUID:       uuid,
				StatusCode: "parsing",
			}
			sendMsgStatus(ctx, writer, &msgStatus)

			// Для парсинга даты и координат
			patternDate := `(?:-ADD\s)([0-9][0-9][0-1][0-9][0-9][0-9])(?:\n)`
			reDate := regexp.MustCompile(patternDate)

			patternCoords := `(?:-ADEPZ\s)(\d+)(N|S|С|Ю)(\d+)(E|W|В|З)(?:\n)`
			reCoords := regexp.MustCompile(patternCoords)

			// SQL Select ORVD
			rowsOrvd, errSql := conn.Query(ctx, `Select id, name from orvd;`)
			if errSql != nil {
				log.Fatalf("Error selecting ORVD: %v\n", errSql)
			}

			orvdByName := make(map[string]int)

			for rowsOrvd.Next() {
				var record OrvdRecord
				err := rowsOrvd.Scan(&record.ID, &record.Name)
				if err != nil {
					log.Fatalf("Failed to scan row ORVD: %v\n", err)
				}
				orvdByName[record.Name] = record.ID
			}

			for indexRow, row := range rows {
				log.Println("row")
				log.Println(indexRow)

				msgNext := BrokerMessageLineStruct{
					UUID:   uuid,
					ROWNUM: indexRow,
					ORVD:   "",
					SHR:    "",
					DEP:    "",
					ARR:    "",
				}

				orvdID := -1
				year := ""
				day := ""
				month := ""
				lat := ""
				long := ""
				pointGeojson := ""
				regionID := -1

				for indexCell, colCell := range row {
					log.Println("colCell")
					log.Println(indexCell)

					switch indexCell {
					case 0:
						msgNext.ORVD = colCell
						orvdID = orvdByName[colCell]
					case 1:
						msgNext.SHR = colCell
					case 2:
						msgNext.DEP = colCell
						// Проверка наличия
						fmt.Println(reDate.MatchString(msgNext.DEP))
						if reDate.MatchString(msgNext.DEP) {

							// Первое совпадение находим так
							dtstr := reDate.FindString(msgNext.DEP)

							year = "20" + dtstr[5:7]
							fmt.Println(year)
							month = dtstr[7:9]
							fmt.Println(month)
							day = dtstr[9:11]
							fmt.Println(day)
						}

						fmt.Println(reCoords.MatchString(msgNext.DEP))

						if reCoords.MatchString(msgNext.DEP) {
							coordstr := reCoords.FindString(msgNext.DEP)

							fmt.Println(coordstr)

							subCoordStr := strings.Split(coordstr, "-ADEPZ ")[1]

							patternLatLong := `(\d+).(\d+)`
							reLatLong := regexp.MustCompile(patternLatLong)

							matches := reLatLong.FindStringSubmatch(subCoordStr)
							fmt.Println(matches)
							fmt.Println(matches[1])
							fmt.Println(matches[2])

							lat = matches[1][0:2] + "." + matches[1][2:]
							long = matches[2][0:3] + "." + matches[2][3:]

							fmt.Println(lat)
							fmt.Println(long)

							pointGeojson = `{"type":"Point","coordinates":[` + long + `,` + lat + `]}`
							fmt.Println(pointGeojson)

							sqlCheckCoords := `SELECT
									r.id,
									r.rus_name
								FROM
									regions AS r
								WHERE
									ST_Contains(
										r.geometry_postgis,
										ST_SetSRID(ST_GeomFromGeoJSON('` + pointGeojson + `'), 4326)
    							)  LIMIT 1;`
							fmt.Println(sqlCheckCoords)

							// Declare variables to store the result
							var name string

							// Execute the query using QueryRow and scan the result into variables
							errCheckCoords := conn.QueryRow(ctx, sqlCheckCoords).Scan(&regionID, &name)
							if errCheckCoords != nil {
								if errCheckCoords == pgx.ErrNoRows {
									fmt.Println("No rows were found.")
								} else {
									log.Fatalf("Unable to scan row: %v\n", errCheckCoords)
								}
							} else {
								fmt.Printf("First item: ID=%d, Name=%s\n", regionID, name)
							}

						}

						// fmt.Println(reDate.FindAllString(msgNext.DEP, 2))
						// Используем поле msgNext.ORVD для получения ORVD
						// И поле msgNext.DEP - для выяснения координат (ADARRZ) и даты (ADA)
						fmt.Print(colCell, "\t")
					case 3:
						msgNext.ARR = colCell
					}

					fmt.Println()

				}

				fmt.Println("orvdID")
				fmt.Println(orvdID)

				fmt.Println("date")
				fmt.Println(year)
				fmt.Println(month)
				fmt.Println(day)

				// Отправляем в БД только распознанные записи
				if regionID > 0 {
					jsonDataNext, errJsonNext := json.Marshal(msgNext)
					if errJsonNext != nil {
						fmt.Println("Error marshaling to JSON:", errJsonNext)
					}

					// SQL Insert statement
					// Update xlsxfile_uuid for exists rows
					// При повторных загрузках одного и того же файла делаем перепривязку уже созданных и обработанных строк
					sql := `INSERT into flights (region_id, orvd_id, coords_postgis, event_date, xlsxfile_uuid, row_num, details, details_SHA256)
			values($1, $2, ST_SetSRID(ST_GeomFromGeoJSON('` + pointGeojson + `'), 4326), MAKE_DATE($3, $4, $5), $6, $7, $8, $9)
			ON CONFLICT (event_date, details_SHA256) DO NOTHING;`

					fmt.Println(sql)

					// Execute the INSERT statement
					hash := sha256.Sum256(jsonDataNext)
					hashString := hex.EncodeToString(hash[:])

					// fmt.Println(jsonDataNext)
					// fmt.Println(string(jsonDataNext))
					// fmt.Println(hashString)

					commandTag, err := conn.Exec(ctx, sql, regionID, orvdID, year, month, day, uuid, indexRow, jsonDataNext, hashString)
					if err != nil {
						log.Fatalf("Error inserting flights: %v\n", err)
					}
					log.Printf("INSERTed %d row(s). Command Tag: %s\n", commandTag.RowsAffected(), commandTag.String())

				}

				// fmt.Println(msgNext)

				fmt.Println()
			}

			// Коммитим вручную после обработки
			errCommit := reader.CommitMessages(ctx, m)
			if errCommit != nil {
				log.Fatal("Ошибка коммита сообщения:", errCommit)
			}

			// Status
			msgStatus = BrokerMessageStatusStruct{
				UUID:       uuid,
				StatusCode: "saved",
			}
			sendMsgStatus(ctx, writer, &msgStatus)

		}
	}
}

type BrokerMessageStruct struct {
	UUID string `json:"uuid"`
}

type BrokerMessageStatusStruct struct {
	UUID       string `json:"uuid"`
	StatusCode string `json:"statusCode"`
}

type BrokerMessageLineStruct struct {
	UUID   string `json:"uuid"`
	ROWNUM int    `json:"rownum"`
	ORVD   string `json:"orvd"`
	SHR    string `json:"shr"`
	DEP    string `json:"dep"`
	ARR    string `json:"arr"`
}

type OrvdRecord struct {
	ID   int
	Name string
}

// Send status to Kafka
func sendMsgStatus(ctx context.Context, writer *kafka.Writer, msg *BrokerMessageStatusStruct) {
	jsonDataNext, errJsonNext := json.Marshal(msg)
	if errJsonNext != nil {
		fmt.Println("Error marshaling to JSON:", errJsonNext)
	}

	kfkMsgNext := kafka.Message{
		Value: []byte(string(jsonDataNext)),
	}

	errWriteNext := writer.WriteMessages(ctx, kfkMsgNext)
	if errWriteNext != nil {
		log.Fatal("Send message error:", errWriteNext)
	}
}
