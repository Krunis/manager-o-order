package common

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/redis/go-redis/v9"
)

type DeliveryInformation struct {
	Table int32 `json:"table"`
}

type Employee struct {
	ID         string `json:"id"`
	Department string `json:"department"`
}

type Item struct {
	ID               string `json:"id"`
	Name             string `json:"name"`
	Count            uint32 `json:"count"`
	ConfirmationType string `json:"confirmation_type"`
}

type Order struct {
	ID                     string
	EmployeeID             string               `json:"employee_id"`
	DepartmentID           string               `json:"department_id"`
	Items                  []*Item              `json:"items"`
	ReserveID              string               `json:"reserve_id"`
	ConfirmationId         string               `json:"confirmation_id"`
	Delivery               *DeliveryInformation `json:"delivery_information"`
	ConfirmationEmployeeID string               `json:"confirmation_employee_id"`
	IdempotencyKey         string               `json:"idempotency_key"`
}

type Lifecycle struct {
	Ctx    context.Context
	Cancel context.CancelFunc
}

func GetDBConnectionString() string {
	var missingEnvVars []string

	checkEnvVar := func(envVar, envVarName string) {
		if envVar == "" {
			missingEnvVars = append(missingEnvVars, envVarName)
		}
	}

	dbName := os.Getenv("POSTGRES_DB")
	checkEnvVar(dbName, "POSTGRES_DB")

	dbUser := os.Getenv("POSTGRES_USER")
	checkEnvVar(dbUser, "POSTGRES_USER")

	dbHost := os.Getenv("POSTGRES_HOST")
	if dbHost == "" {
		dbHost = "localhost"
	}

	dbPort := os.Getenv("POSTGRES_PORT")
	checkEnvVar(dbPort, "POSTGRES_PORT")

	dbPassword := os.Getenv("POSTGRES_PASSWORD")
	checkEnvVar(dbPassword, "POSTGRES_PASSWORD")

	if len(missingEnvVars) > 0 {
		log.Fatalf("Required environment variables are not set: %s",
			strings.Join(missingEnvVars, ","))
	}

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", dbUser, dbPassword, dbHost, dbPort, dbName)
}

func ConnectToDB(ctx context.Context, dbConnectionString string) (*pgxpool.Pool, error) {
	var err error

	timer := time.NewTimer(time.Second * 26)
	defer timer.Stop()

	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			dbPool, err := pgxpool.Connect(ctx, dbConnectionString)
			if err == nil {
				log.Println("Connected to DB")
				return dbPool, nil
			}

			log.Printf("Failed to connect to DB: %s. Retrying...\n", err)

		case <-timer.C:
			return nil, fmt.Errorf("db connection timeout (25s): %v", err)
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

func ConnectToRedis(ctx context.Context) (*redis.Client, error) {
	redisPort := os.Getenv("REDIS_PORT")

	redisDB := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("redis:%s", redisPort),
		DB:   0,
	})

	if err := redisDB.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect redis db: %s", err)
	}

	log.Println("Connected to Redis")

	return redisDB, nil
}
