package router

import (
	"fmt"

	fileController "whatsapp_file_handling/controller"
	middlewares "whatsapp_file_handling/middlewares"

	"github.com/gin-gonic/gin"
)

func RouterInit(route *gin.Engine) {
	groupRoute := route.Group("/api/v1")

	groupRoute.GET("/health", middlewares.CheckAuth, func(c *gin.Context) {
		fmt.Println("Incoming Origin:", c.Request.Header.Get("Origin"))

		c.JSON(200, gin.H{"message": "healthy", "status": 200})
	})

	groupRoute.POST("/upload", fileController.UploadFileHandler)
}
