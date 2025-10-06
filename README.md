# Practice with Postgres full text search

This repo gives you a starting point for exploring the full text search capabilities of Postgres.

## Getting started

You must have Go and Docker installed.

- `go mod install`
- `docker compose up -d`
  - This will start Postgres and run the `init.sql` file
- `go run ./cmd/... -dsn "postgres://postgres:postgres@localhost:5432/postgres" -count 1000`
  - You can change the value of `count` to be whatever you need
  - Alternatively use `./dev.sh`
    - May require `chmod +x ./dev.sh`

## Full text search

The column `products.product_search_vector` is already of type `tsvector` so you are ready to start exploring full text search.

Here's a basic example where we query for products with the search term "computer".

```sql
SELECT
    *
FROM
    products
WHERE
    product_search_vector @@ to_tsquery('english', 'computer')
```

Here's a query that allows you to see where in the text a search term appears.

```sql
SELECT
    ts_headline(
		description,
		to_tsquery('english', 'wood'),
		'StartSel = <, StopSel = />, MinWords=5, MaxWords=7, MaxFragments=1'
	)
FROM
    products
WHERE
    product_search_vector @@ to_tsquery('english', 'wood')
```

Finally, it can be interesting to see how much space on disk the `product_search_vector` takes up.

```sql
WITH cte_sum AS (
	SELECT
		count(*),
		sum(pg_column_size(product_search_vector)) AS column_size_bytes
	FROM products
)
SELECT
	"count" AS num_rows,
	pg_size_pretty(column_size_bytes / "count") AS avg_column_size_bytes,
	pg_size_pretty(column_size_bytes) AS total_column_size
FROM
	cte_sum;
```
