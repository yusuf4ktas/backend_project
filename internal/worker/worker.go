package worker

import (
	"context"
	"fmt"
	"log"

	"github.com/yusuf4ktas/backend-project/internal/service"
)

type Job struct {
	FromUserID      int64
	ToUserID        int64
	Amount          float64
	TransactionType string
}

type Worker struct {
	id         int
	jobQueue   chan Job
	workerPool chan chan Job // A channel to register this worker's jobQueue to the pool.
}

type Dispatcher struct {
	workerPool chan chan Job
	maxWorkers int
	jobQueue   chan Job
	service    service.TransactionService
}

func NewDispatcher(maxWorkers int, service service.TransactionService) *Dispatcher {
	return &Dispatcher{
		workerPool: make(chan chan Job, maxWorkers),
		maxWorkers: maxWorkers,
		jobQueue:   make(chan Job, 100),
		service:    service,
	}
}

func NewWorker(id int, workerPool chan chan Job) Worker {
	return Worker{
		id:         id,
		jobQueue:   make(chan Job),
		workerPool: workerPool,
	}
}

// Worker start listening for jobs.
func (w Worker) Start(ctx context.Context, service service.TransactionService) {
	go func() {
		for {
			// Worker is ready for a new job
			w.workerPool <- w.jobQueue

			select {
			case job := <-w.jobQueue:
				switch job.TransactionType {
				case "credit":
					_, err := service.Credit(context.Background(), job.ToUserID, job.Amount)
					if err != nil {
						log.Printf("ERROR: worker %d failed to process credit job for user %d: %v", w.id, job.ToUserID, err)
					} else {
						fmt.Printf("Worker %d: successfully processed credit for user %d of amount %.2f\n", w.id, job.ToUserID, job.Amount)
					}
				case "debit":
					_, err := service.Debit(context.Background(), job.FromUserID, job.Amount)
					if err != nil {
						log.Printf("ERROR: worker %d failed to process debit job for user %d: %v", w.id, job.FromUserID, err)
					} else {
						fmt.Printf("Worker %d: successfully processed debit for user %d of amount %.2f\n", w.id, job.FromUserID, job.Amount)
					}
				case "transfer", "": // type is "transfer" or empty
					_, err := service.Transfer(context.Background(), job.FromUserID, job.ToUserID, job.Amount)
					if err != nil {
						log.Printf("ERROR: worker %d failed to process transfer job from user %d to user %d: %v", w.id, job.FromUserID, job.ToUserID, err)
					} else {
						fmt.Printf("Worker %d: successfully processed transfer from user %d to %d\n", w.id, job.FromUserID, job.ToUserID)
					}
				default:
					log.Printf("ERROR: worker %d received job with unknown type: '%s'", w.id, job.TransactionType)
				}

			case <-ctx.Done():
				// The context was cancelled, so the worker should stop.
				fmt.Printf("Worker %d: stopping.\n", w.id)
				return
			}
		}
	}()
}

// Sarts all the workers and begins listening for jobs.
func (d *Dispatcher) Run(ctx context.Context) {
	for i := 0; i < d.maxWorkers; i++ {
		worker := NewWorker(i+1, d.workerPool)
		worker.Start(ctx, d.service) //runs its own goroutine
	}

	//main dispatch loop in a separate goroutine
	go d.dispatch(ctx)
}

// AddJob is a public method to add a new job to the queue.
func (d *Dispatcher) AddJob(job Job) {
	d.jobQueue <- job
}

func (d *Dispatcher) dispatch(ctx context.Context) {
	for {
		select {
		case job := <-d.jobQueue:
			jobChannel := <-d.workerPool

			jobChannel <- job

		case <-ctx.Done():
			return
		}
	}
}
