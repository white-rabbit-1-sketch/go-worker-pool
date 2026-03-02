package worker_pool

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
)

type WorkerPool[T any] struct {
	ctx                 context.Context
	wg                  sync.WaitGroup
	once                sync.Once
	taskChan            chan func(ctx context.Context) T
	resultChan          chan T
	minWorkersCount     int
	maxWorkersCount     int
	currentWorkersCount int32
	bufferSize          int
}

func NewWorkerPool[T any](ctx context.Context, minWorkersCount int, maxWorkersCount int, bufferSize int) *WorkerPool[T] {
	return &WorkerPool[T]{
		ctx:             ctx,
		taskChan:        make(chan func(ctx context.Context) T, bufferSize),
		resultChan:      make(chan T, bufferSize),
		minWorkersCount: minWorkersCount,
		maxWorkersCount: maxWorkersCount,
		bufferSize:      bufferSize,
	}
}

func (p *WorkerPool[T]) Start() {
	p.once.Do(func() {
		for i := 0; i < p.minWorkersCount; i++ {
			atomic.AddInt32(&p.currentWorkersCount, 1)
			p.wg.Go(func() {
				defer atomic.AddInt32(&p.currentWorkersCount, -1)
				w := NewWorker(p.ctx, p.taskChan, p.resultChan, false)
				w.Handle()
			})
		}

		go func() {
			p.wg.Wait()

			close(p.taskChan)
			close(p.resultChan)
		}()

		go func() {
			ticker := time.NewTicker(time.Millisecond * 500)
			for {
				select {
				case <-p.ctx.Done():
					return
				case <-ticker.C:
					if len(p.taskChan) > (p.bufferSize/2) && int(atomic.LoadInt32(&p.currentWorkersCount)) < p.maxWorkersCount {
						atomic.AddInt32(&p.currentWorkersCount, 1)
						p.wg.Go(func() {
							defer atomic.AddInt32(&p.currentWorkersCount, -1)
							w := NewWorker(p.ctx, p.taskChan, p.resultChan, true)
							w.Handle()
						})
					}
				}
			}
		}()
	})
}

func (p *WorkerPool[T]) Wait() {
	p.wg.Wait()
}

func (p *WorkerPool[T]) AddTask(task func(ctx context.Context) T) error {
	select {
	case p.taskChan <- task:
		return nil
	case <-p.ctx.Done():
		return p.ctx.Err()
	}
}

func (p *WorkerPool[T]) GetResultChan() <-chan T {
	return p.resultChan
}
