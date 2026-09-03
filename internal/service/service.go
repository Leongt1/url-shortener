package service

import (
	"context"
	"crypto/rand"
	"errors"
	"math/big"

	"github.com/Leongt1/url-shortener/internal/repository"
)

type Service struct {
	repo *repository.Repository
}

func NewService(repo *repository.Repository) *Service {
	return &Service{
		repo: repo,
	}
}

// shortens the given url into a Base62 shortened url
func (s *Service) Shorten(ctx context.Context, url string) (string, error) {
	retryCount := 5

	for range retryCount {
		rndStr, err := generateRandomStr()
		if err != nil {
			return "", err
		}

		ok, err := s.repo.SetUrl(ctx, rndStr, url, 0)
		if err != nil {
			return "", err
		}

		if ok {
			return rndStr, nil
		}
	}

	return "", errors.New("could not generate a unique code")
}

// maps the short url to the long url and returns it
func (s *Service) Resolve(ctx context.Context, code string) (string, error) {
	return s.repo.GetUrl(ctx, code)
}

// picks 6 random characters from 62 characters
func generateRandomStr() (string, error) {
	const letters = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz"
	bytes := make([]byte, 6)

	for i := 0; i < len(bytes); i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(letters))))
		if err != nil {
			return "", err
		}
		bytes[i] = letters[num.Int64()]
	}

	return string(bytes), nil
}
