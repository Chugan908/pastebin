db_migrations_up: goose postgres "host=localhost user=postgres password=lobsteri228 dbname=text_urls port=5432 sslmode=disable" up
db_migrations_down: goose postgres "host=localhost user=postgres password=lobsteri228 dbname=text_urls port=5432 sslmode=disable" down
