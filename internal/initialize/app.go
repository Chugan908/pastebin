package initialize

import (
	"context"
	"fmt"
	"os"
	"pastebin/internal/controllers"
	"pastebin/internal/services/cache"
	"pastebin/internal/services/object_storage"
	"pastebin/internal/services/storage"
)

func NewApp() *controllers.Handler {
	db, err := storage.New(os.Getenv("DATABASE_DSN"))
	if err != nil {
		panic(fmt.Errorf("couldn't connect to database:%w", err))
	}

	rdb, err := cache.New(0)
	if err != nil {
		panic(err)
	}

	objClient := object_storage.New()

	return controllers.New(context.Background(), db, rdb, objClient)
}
