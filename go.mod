module github.com/francisco/gridironmind

// +heroku install ./cmd/server

go 1.21

require (
	github.com/go-redis/redis/v8 v8.11.5
	github.com/google/uuid v1.6.0
	github.com/jackc/pgx/v5 v5.5.5
	github.com/joho/godotenv v1.5.1
)
