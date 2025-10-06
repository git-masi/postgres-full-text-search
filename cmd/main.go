package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"math"
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

	current := *count
	chunkSize := 100_000
	numChunks := int(math.Ceil(float64(*count) / float64(chunkSize)))
	chunkCount := 1

	for current > 0 {
		size := min(current, chunkSize)

		logger.Info(
			"Creating products",
			slog.String("chunk", fmt.Sprintf("%d of %d", chunkCount, numChunks)),
			slog.Int("num_products", size),
		)

		var products [][]any

		for range size {
			name := gofakeit.ProductName()
			description := fmt.Sprintf("%s - %s", name, gofakeit.ProductDescription())
			price := fmt.Sprint(gofakeit.Price(9.99, 199.99))

			products = append(products, []any{
				name,
				description,
				price,
			})
		}

		logger.Info("Inserting products into DB")

		numInserted, err := conn.CopyFrom(
			context.TODO(),
			pgx.Identifier{"products"},
			[]string{"name", "description", "price"},
			pgx.CopyFromRows(products),
		)
		if err != nil {
			logger.Error(err.Error())
			os.Exit(1)
		}

		logger.Info("Chunk inserted", slog.Int64("num_inserted", numInserted))

		current = current - chunkSize
		chunkCount++
	}

	logger.Info("Finished")
}
