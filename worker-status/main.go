package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/lpernett/godotenv"
	"github.com/segmentio/kafka-go"
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
	fmt.Println("Run worker status")
	KAFKA_BROKER := env("KAFKA_BROKER")
	KAFKA_TOPIC := env("KAFKA_TOPIC")
	KAFKA_GROUP := env("KAFKA_GROUP")

	log.Println("kafka.NewReader")
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{KAFKA_BROKER}, // Replace with your Kafka broker addresses
		Topic:   KAFKA_TOPIC,
		GroupID: KAFKA_GROUP,
		// MinBytes:  10e3, // 10KB
		// MaxBytes:  10e6, // 10MB
		Partition: 0, // Or use GroupID for consumer group management across partitions
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

	go worker(ctx, reader, conn)

	// Keep the main goroutine alive, e.g., by waiting for a signal
	select {}
}

func worker(ctx context.Context, reader *kafka.Reader, conn *pgx.Conn) {
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

			var data BrokerMessageStatusStruct
			errJson := json.Unmarshal(m.Value, &data)
			if errJson != nil {
				log.Fatal("Fail unmarshaling JSON:", errJson)
			}
			uuid := data.UUID
			log.Println(uuid)
			statusCode := data.StatusCode
			log.Println(statusCode)

			// SQL Update statement
			sql := `UPDATE xlsxfiles
					SET "statusCode" = ($1)
					WHERE uuid = ($2);`

			// Execute the INSERT statement
			commandTag, err := conn.Exec(ctx, sql, statusCode, uuid)
			if err != nil {
				log.Fatalf("Error inserting Name: %v\n", err)
			}
			log.Printf("Updated %d row(s). Command Tag: %s\n", commandTag.RowsAffected(), commandTag.String())

		}
	}
}

type BrokerMessageStatusStruct struct {
	UUID       string `json:"uuid"`
	StatusCode string `json:"statusCode"`
}
