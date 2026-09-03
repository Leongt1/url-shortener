package routes

import (
	"github.com/Leongt1/url-shortener/internal/handler"
	"github.com/gin-gonic/gin"
)

func Register(r *gin.Engine, h *handler.Handler) {
	r.POST("/shorten", h.Shorten)
	r.GET("/:code", h.Redirect)
}
