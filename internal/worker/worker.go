package worker

import (
	"context"
	"log"

	"github.com/Leongt1/url-shortener/internal/repository"
)

type Worker struct {
	repo *repository.Repository
}

func NewWorker(repo *repository.Repository) *Worker {
	return &Worker{repo: repo}
}

func (w *Worker) Run(ctx context.Context) {
	for {
		code, err := w.repo.PopClick(ctx, 0)
		if err != nil {
			log.Println(err.Error())
			continue
		}

		if err := w.repo.IncrClickCount(ctx, code); err != nil {
			log.Println(err.Error())
			continue
		}

		log.Printf("counted click: %s", code)
	}
}
