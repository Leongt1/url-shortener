package models

type UrlRequest struct {
	Url string `json:"url" binding:"required,url"`
}
