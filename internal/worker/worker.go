package worker

import (
	"context"
	"log"
	"time"

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
		select {
		case <-ctx.Done():
			log.Println("worker: shutting down")
			return
		default:
		}

		code, popped, err := w.repo.PopClick(ctx, 1*time.Second)
		if err != nil {
			if ctx.Err() != nil {
				log.Println("worker: shutting down")
				return
			}
			log.Println(err.Error())
			continue
		}

		if !popped {
			continue
		}

		if err := w.repo.IncrClickCount(context.Background(), code); err != nil {
			log.Println(err.Error())
			continue
		}
	}
}
