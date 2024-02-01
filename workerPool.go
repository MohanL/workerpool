package workerpool

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"
)

// TaskFunc is the type of function that can be submitted to the worker pool.
// It returns a result and an error. You would replace interface{} with whatever
// result type your tasks are supposed to return.
type TaskFunc func() (interface{}, error)

// type TaskFuncWithId func(int64 id) (interface{}, error)
type TaskFuncWithId struct {
	Task   TaskFunc
	TaskId int64
}

type SubmitResult struct {
	TaskId int64
	Result interface{}
	Error  error
}

type WorkerPoolConfig struct {
	MaxWorkers int           // Maximum number of worker goroutines
	Timeout    time.Duration // Maximum time to wait for task completion
}

type WorkerPool struct {
	isStopped       int32 // atomic flag
	config          WorkerPoolConfig
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	publishers      chan TaskFuncWithId
	workerStopChans []chan bool
	taskId          atomic.Int64
	resultChan      chan SubmitResult
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
		publishers:      make(chan TaskFuncWithId, taskQueueSize),
		workerStopChans: make([]chan bool, maxWorkers),
		taskId:          atomic.Int64{},
		resultChan:      make(chan SubmitResult, taskQueueSize),
	}

	for i := 0; i < maxWorkers; i++ {
		stopChan := make(chan bool)
		pool.workerStopChans[i] = stopChan
		pool.wg.Add(1)
		go pool.worker(i+1, stopChan)
	}
	return pool
}

// worker is a method on the WorkerPool that processes tasks from the taskQueue.
func (wp *WorkerPool) worker(id int, stopChan chan bool) {
	defer wp.wg.Done()
	// defer fmt.Printf("worker %d stopped\n", id)

	for {
		select {
		case <-wp.ctx.Done(): // Check if context was cancelled (pool is stopping)
			return
		case <-time.After(wp.config.Timeout):
			fmt.Printf("worker %d timed out\n", id)
			return
		case <-stopChan: // Check if this specific worker was told to stop
			return
		case task, ok := <-wp.publishers: // Wait for a task
			if !ok {
				// The publishers channel was closed, no more tasks will come
				return
			}
			if task.Task != nil {
				// fmt.Printf("worker %d is working on task %d\n", id, task.TaskId)
				result, err := task.Task()

				if err != nil {
					fmt.Printf("worker %d error on task %d: %v\n", id, task.TaskId, err)
				}

				if result == nil {
					continue
				}

				taskResult := SubmitResult{
					TaskId: task.TaskId,
					Result: result,
					Error:  err,
				}

			loop:
				for {
					select {
					case wp.resultChan <- taskResult:
						// Task sent successfully
						break loop
					default:
						// Channel is full, handle the case when the channel is full
						fmt.Printf("worker %d stuck on sending task %d result, resultChan is full, cannot send result\n", id, task.TaskId)
						panic("resultChan is full, cannot send result")
					}
				}
			}
		}
	}
}

func (wp *WorkerPool) Submit(task TaskFunc) (int64, <-chan SubmitResult, error) {
	// Create a buffered channel for the result.
	if atomic.LoadInt32(&wp.isStopped) == 1 {
		return 0, nil, errors.New("worker pool is not accepting new tasks")
	}
	taskId := wp.taskId.Add(1)

	TaskFuncWithId := TaskFuncWithId{
		Task:   task,
		TaskId: taskId,
	}

loop:
	for {
		select {
		case wp.publishers <- TaskFuncWithId:
			// Task sent successfully
			break loop
		default:
			// Channel is full, handle the case when the channel is full
			fmt.Println("publishers Channel is full, cannot send task")
		}
	}
	// fmt.Printf("worker pool submitted task %d\n", taskId)
	return taskId, wp.resultChan, nil
}

func (wp *WorkerPool) WaitAll() {
	wp.wg.Wait()
}

func (wp *WorkerPool) Stop() {
	atomic.StoreInt32(&wp.isStopped, 1)
	// First, stop all workers by cancelling the context.
	wp.cancel()

	// Wait for all workers to finish.
	wp.wg.Wait()

	for _, stopChan := range wp.workerStopChans {
		close(stopChan)
	}

	// Close the publishers channel to signal no more tasks will be sent.
	// This is safe only after we have ensured all workers have stopped.
	close(wp.publishers)
	close(wp.resultChan)

	//TODO: Drain the resultChan.
	// Optionally, you can also drain the resultChan here if needed,
	// and possibly close it if no more results will be processed.
	// Be aware that closing a channel while it is still being written to
	// by other goroutines will cause a panic.
}
