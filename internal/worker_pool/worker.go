package worker_pool

import (
	"context"
	"fmt"
	"time"
)

type Worker[T any] struct {
	ctx           context.Context
	taskChan      chan func(ctx context.Context) T
	resultChan    chan T
	stopOnTimeout bool
}

func NewWorker[T any](
	ctx context.Context,
	taskChan chan func(ctx context.Context) T,
	resultChan chan T,
	stopOnTimeout bool,
) *Worker[T] {
	return &Worker[T]{
		ctx:           ctx,
		taskChan:      taskChan,
		resultChan:    resultChan,
		stopOnTimeout: stopOnTimeout,
	}
}

func (w *Worker[T]) Handle() {
	for {
		select {
		case task, ok := <-w.taskChan:
			if !ok {
				return
			}

			func() {
				defer func() {
					if r := recover(); r != nil {
						fmt.Println("Recovered in worker", r)
					}
				}()

				w.resultChan <- task(w.ctx)
			}()
		case <-w.ctx.Done():
			return
		case <-time.After(time.Second * 10):
			if w.stopOnTimeout {
				return
			}
		}
	}
}
