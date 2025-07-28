package main

import (
	"log"
	"os"

	"ai-phone-support/internal/db"
	"ai-phone-support/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {

	// Load env variables
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatalln("Failed to load environment variables!")
	}

	// Connect to database
	dbService := db.NewDBService(
		os.Getenv("DB_HOST"),
		os.Getenv("DB_USER"),
		os.Getenv("DB_PASSWORD"),
		os.Getenv("DB_NAME"),
		nil,
		true,
		os.Getenv("DB_TIMEZONE"),
	)
	if err := dbService.Connect(); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
		return
	}

	// Migrate database
	if err := dbService.Migrate(); err != nil {
		log.Fatalf("Failed to migrate database: %v", err)
		return
	}

	// Start HTTP Server
	var engine *gin.Engine = gin.Default()

	routes.GlobalRoutes(engine)

	engine.Run(os.Getenv("BASE_URL") + ":" + os.Getenv("PORT"))
}
