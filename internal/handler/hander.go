package handler

import (
	"errors"
	"net/http"

	"github.com/Leongt1/url-shortener/internal/models"
	"github.com/Leongt1/url-shortener/internal/repository"
	"github.com/Leongt1/url-shortener/internal/service"
	"github.com/gin-gonic/gin"
)

type Handler struct {
	svc *service.Service
}

func NewHandler(svc *service.Service) *Handler {
	return &Handler{
		svc: svc,
	}
}

func (h *Handler) Shorten(c *gin.Context) {
	var req models.UrlRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	shortenedStr, err := h.svc.Shorten(c.Request.Context(), req.Url)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"code":      shortenedStr,
		"short_url": "http://localhost:8080/" + shortenedStr,
	})
}

func (h *Handler) Redirect(c *gin.Context) {
	code := c.Param("code")

	longUrl, err := h.svc.Resolve(c.Request.Context(), code)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Redirect(http.StatusFound, longUrl)
}
