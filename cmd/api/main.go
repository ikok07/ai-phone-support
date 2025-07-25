package main

import (
	"log"
	"os"

	"ai-phone-support/internal/routes"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env.local")
	if err != nil {
		log.Fatalln("Failed to load environment variables!")
	}
	var engine *gin.Engine = gin.Default()

	routes.GlobalRoutes(engine)

	engine.Run(os.Getenv("BASE_URL") + ":" + os.Getenv("PORT"))
}
