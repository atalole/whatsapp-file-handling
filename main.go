package main

import (
	"fmt"
	"os"
	initializers "whatsapp_file_handling/initializers"
	"whatsapp_file_handling/router"

	helmet "github.com/danielkov/gin-helmet"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

var app = gin.New()

func main() {

	godotenv.Load()
	initializers.ConnectDB()

	var port string = os.Getenv("PORT")
	if port == "" {
		port = "4002"
	}

	app.Use(gin.Logger())
	app.Use(gin.Recovery()) // optional

	app.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:4001",
			"https://ws-meta-whatsapp-uat.quickhub.ai",
			"https://ws-meta-whatsapp-staging.quickhub.ai"},
		AllowMethods:  []string{"GET", "POST"},
		AllowHeaders:  []string{"Origin", "Content-Type", "Authorization", "host"},
		AllowWildcard: true,
	}))

	app.Use(helmet.Default())

	// 🔥 Initialize routes
	router.RouterInit(app)
	fmt.Printf("🔥 Server is running on port %s\n", port)
	app.Run(":" + port)
}
