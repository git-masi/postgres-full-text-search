package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"

	"github.com/brianvoe/gofakeit/v7"
	"github.com/jackc/pgx/v5"
)

func main() {
	// "postgres://username:password@localhost:5432/database_name"
	dsn := flag.String("dsn", "", "A Postgresql DSN to connect to")
	count := flag.Int("count", 0, "The number of products to create in the DB")
	flag.Parse()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{}))

	conn, err := pgx.Connect(context.Background(), *dsn)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}
	defer conn.Close(context.Background())

	logger.Info("Creating new products in DB", slog.Int("count", *count))

	var products [][]any

	for i := range *count {
		name := gofakeit.ProductName()
		description := fmt.Sprintf("%s - %s", name, gofakeit.ProductDescription())
		price := fmt.Sprint(gofakeit.Price(9.99, 199.99))

		products = append(products, []any{
			name,
			description,
			price,
		})

		if i%10_000 == 0 {
			logger.Info(fmt.Sprintf("Progress: %d/%d", i, *count))
		}
	}

	logger.Info("Attempting to insert values")

	toInsert, err := Chunk(products, 100_000)
	if err != nil {
		logger.Error(err.Error())
		os.Exit(1)
	}

	for i, p := range toInsert {
		numInserted, err := conn.CopyFrom(
			context.TODO(),
			pgx.Identifier{"products"},
			[]string{"name", "description", "price"},
			pgx.CopyFromRows(p),
		)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}

		logger.Info(fmt.Sprintf("Chunk %d/%d finished", i+1, len(toInsert)), slog.Int64("num_inserted", numInserted))
	}

	logger.Info("Finished inserting rows")
}

func Chunk[T any](slice []T, size int) ([][]T, error) {
	if size < 1 {
		return nil, errors.New("error message")
	}
	if len(slice) < size {
		return nil, errors.New("error message")
	}

	batches := make([][]T, 0, ((len(slice)-1)/size)+1)

	for size < len(slice) {
		slice, batches = slice[size:], append(batches, slice[0:size:size])
	}
	batches = append(batches, slice)
	return batches, nil
}
