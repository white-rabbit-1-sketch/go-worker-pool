# Strongly Typed Dynamic Worker Pool (Educational Example)

[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Go Version](https://img.shields.io/badge/Go-1.18%2B-blue.svg)](https://golang.org/doc/devel/release#go1.18)

> **Disclaimer**: This is a **demonstration project** created for educational purposes and as a showcase of Go concurrency patterns. It is not intended for production use without further auditing, logging, and metrics integration.

## Concept

A high-performance, elastic worker pool implementation in Go that adapts to the workload. This project demonstrates how to build robust concurrent systems using standard library primitives.

### Key Engineering Patterns:
* **Generics (`T any`)**: Full type safety for task results, eliminating `interface{}` and type assertions.
* **Dynamic Scaling**: The pool automatically spawns "seasonal" workers when the task queue is more than 50% full.
* **Self-Healing (Scale-Down)**: Extra workers terminate themselves after 10 seconds of inactivity to free up system resources.
* **Panic Isolation**: Each task is wrapped in a `recover()` block, preventing a single faulty task from crashing the entire worker or the host process.
* **Context Propagation**: Tasks are "context-aware", allowing for granular cancellation and timeout management.
* **Graceful Shutdown**: Properly handles OS signals, ensuring all in-flight tasks are processed before exit.

---

## Architecture

The pool maintains a core set of `minWorkers` and can scale up to `maxWorkers` during high-pressure periods.

* **Dispatcher**: A ticker-based monitor that evaluates queue pressure every 500ms.
* **Atomic Counters**: Uses `sync/atomic` for thread-safe worker tracking without mutex contention.
* **Flow Control**: Implements backpressure via buffered channels.

---

## 💻 Usage

```go
// Initialize a pool for 'bool' results
// ctx, minWorkers: 5, maxWorkers: 20, bufferSize: 100
pool := worker_pool.NewWorkerPool[bool](ctx, 5, 20, 100)
pool.Start()

// Start result consumer
go func() {
    for res := range pool.GetResultChan() {
        _ = res // Handle results
    }
}()

// Add tasks
err := pool.AddTask(func(taskCtx context.Context) bool {
    // Context-aware business logic
    return true
})

// Wait for graceful shutdown
pool.Wait()