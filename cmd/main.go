package main

import (
	"context"
	"fmt"
	"go_worker_pool/internal/worker_pool"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	fmt.Println("Starting worker pool...")

	p := worker_pool.NewWorkerPool[int](ctx, 5, 10, 500)
	p.Start()

	go func() {
		for result := range p.GetResultChan() {
			fmt.Printf("Executed task: %d\n", result)
		}
	}()

	go func() {
		for i := 0; i < 1000; i++ {
			err := p.AddTask(func(ctx context.Context) int {
				return i
			})

			if err != nil {
				panic(err)
			}
		}
	}()

	<-ctx.Done()
	fmt.Println("Shutting down...")

	p.Wait()
}
