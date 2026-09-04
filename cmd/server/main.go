package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/Leongt1/url-shortener/internal/handler"
	"github.com/Leongt1/url-shortener/internal/repository"
	"github.com/Leongt1/url-shortener/internal/routes"
	"github.com/Leongt1/url-shortener/internal/service"
	"github.com/Leongt1/url-shortener/internal/worker"
	"github.com/gin-gonic/gin"
)

func main() {
	var wg sync.WaitGroup
	r := gin.Default()
	logger := log.Default()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo, err := repository.NewRepository(":6379")
	if err != nil {
		logger.Fatal(err.Error())
	}

	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	r.GET("/health", func(c *gin.Context) {
		if err := repo.Ping(ctx); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "healthy",
		})
	})

	routes.Register(r, h)

	// starting worker
	w := worker.NewWorker(repo)
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Run(ctx)
	}()

	srv := &http.Server{Addr: ":8080", Handler: r}
	go func() {
		fmt.Println("Server started in port: 8080")
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("error starting server:")
		}
	}()

	<-ctx.Done()

	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Println("server shutdown error:", err)
	}

	wg.Wait()
}
