package workerpool

import (
	"context"
	"sync"
	"time"
)

// TaskFunc is the type of function that can be submitted to the worker pool.
// It returns a result and an error. You would replace interface{} with whatever
// result type your tasks are supposed to return.
type TaskFunc func() (interface{}, error)

type SubmitResult struct {
	Result interface{}
	Error  error
}

type WorkerPoolConfig struct {
	MaxWorkers int           // Maximum number of worker goroutines
	Timeout    time.Duration // Maximum time to wait for task completion
}

type WorkerPool struct {
	config          WorkerPoolConfig
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	publishers      chan TaskFunc
	workerStopChans []chan bool
}

func NewWorkerPool(maxWorkers int, taskQueueSize int, timeout time.Duration) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		config: WorkerPoolConfig{
			MaxWorkers: maxWorkers,
			Timeout:    timeout,
		},
		ctx:             ctx,
		cancel:          cancel,
		publishers:      make(chan TaskFunc, taskQueueSize),
		workerStopChans: make([]chan bool, maxWorkers),
	}

	for i := 0; i < maxWorkers; i++ {
		stopChan := make(chan bool)
		pool.workerStopChans[i] = stopChan
		go pool.worker(stopChan)
	}
	return pool
}

// worker is a method on the WorkerPool that processes tasks from the taskQueue.
func (wp *WorkerPool) worker(stopChan chan bool) {
	for {
		select {
		case <-wp.ctx.Done(): // Check if context was cancelled (pool is stopping)
			return
		case <-stopChan: // Check if this specific worker was told to stop
			return
		case task := <-wp.publishers: // Wait for a task
			// Execute the task
			if task != nil { // Check if the task is nil (should not happen) TODO: remove this check?
				task() // The task is executed here
				// Handle the result and error as needed
			}
		}
	}
}

func (wp *WorkerPool) Submit(task TaskFunc) <-chan SubmitResult {
	// Create a buffered channel for the result.
	resultChan := make(chan SubmitResult)
	wp.publishers <- task
	// work on the reuslt chan

	// Notes: queue the task and let the worker fetch it from the queue. the question is how to connectthe submitResult with the worker

	return resultChan

}

func main() {

}
