package backend

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
)

func DBConnect() bool {
	pgConnString := fmt.Sprintf("postgres://%s:%s@%s:5432/%s",
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_HOST"),
		os.Getenv("DB_DATABASE"),
	)

	conn, err := pgx.Connect(context.Background(), pgConnString)

	defer conn.Close(context.Background())

	if err != nil {
		log.Fatal(err)
		return false
	}

	return true
}
