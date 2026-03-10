package server

import (
	"net/http"
	"main/internal/database" 
	"github.com/gin-gonic/gin"
)



func SetupRouter() *gin.Engine {
	// Create a Gin router with default middleware (logger and recovery)
	r := gin.Default()


	r.POST("/api/websites", func(c *gin.Context) {
		name := c.PostForm("name")
		url := c.PostForm("url")
		dsc := c.PostForm("description")
	})
}