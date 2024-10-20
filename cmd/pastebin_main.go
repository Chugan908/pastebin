package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"pastebin/internal/initialize"
)

func main() {
	server := gin.Default()

	err := godotenv.Load()
	if err != nil {
		panic(fmt.Errorf("couldn't load the environment:%w", err))
	}

	handler := initialize.NewApp()

	initialize.SetupRoutes(handler, server)

	server.Run("localhost:8080")

	// TODO: bcrypt the address of file and put in the storage
	// send user the name of file
	// when the user wants to get the body, provided url is compared to the stored one
}
