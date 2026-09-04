package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/Leongt1/url-shortener/internal/handler"
	"github.com/Leongt1/url-shortener/internal/repository"
	"github.com/Leongt1/url-shortener/internal/routes"
	"github.com/Leongt1/url-shortener/internal/service"
	"github.com/Leongt1/url-shortener/internal/worker"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	logger := log.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "healthy",
		})
	})

	repo, err := repository.NewRepository(":6379")
	if err != nil {
		logger.Fatal(err.Error())
	}

	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	routes.Register(r, h)

	// starting worker
	w := worker.NewWorker(repo)
	go w.Run(context.Background())

	fmt.Println("Server is running in port :8080")
	if err := r.Run(":8080"); err != nil {
		logger.Fatal("Server failed: ", err)
	}
}
