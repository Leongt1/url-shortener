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
	"github.com/Leongt1/url-shortener/internal/metrics"
	"github.com/Leongt1/url-shortener/internal/middleware"
	"github.com/Leongt1/url-shortener/internal/repository"
	"github.com/Leongt1/url-shortener/internal/routes"
	"github.com/Leongt1/url-shortener/internal/service"
	"github.com/Leongt1/url-shortener/internal/worker"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	var wg sync.WaitGroup
	logger := log.Default()

	if err := godotenv.Load(); err != nil {
		log.Println("no .env file found, reading config from environment")
	}

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	r := gin.Default()
	r.Use(middleware.RequestMetrics())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	repo, err := repository.NewRepository(redisAddr)
	if err != nil {
		logger.Fatal(err.Error())
	}

	svc := service.NewService(repo)
	h := handler.NewHandler(svc)

	r.GET("/health", func(c *gin.Context) {
		if err := repo.Ping(c.Request.Context()); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"error": err.Error(),
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message": "healthy",
		})
	})

	metrics.RegisterQueueDepth(func() float64 {
		queueCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)

		queueLen, err := repo.QueueDepth(queueCtx)
		cancel()

		if err != nil {
			log.Println("error processing queue depth:", err)
			return -1
		}

		return float64(queueLen)
	})

	// Prometheus
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	routes.Register(r, h)

	// starting worker
	w := worker.NewWorker(repo)
	wg.Add(1)
	go func() {
		defer wg.Done()
		w.Run(ctx)
	}()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	srv := &http.Server{Addr: ":" + port, Handler: r}
	go func() {
		fmt.Println("Server started in port: " + port)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal("error starting server:", err)
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
