package workerpool

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
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

type Logger interface {
	Log(message string)
}

type defaultLogger struct{}

func (l *defaultLogger) Log(message string) {
	fmt.Println(message)
}

type WorkerPoolConfig struct {
	MaxWorkers    int           // Maximum number of worker goroutines
	Timeout       time.Duration // Maximum time to wait for task completion
	TaskQueueSize int           // Use with Caution
	Logger        Logger
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
	logger          Logger
	stats           map[string]prometheus.Metric
}

func NewWorkerPool(workerPoolConfig WorkerPoolConfig) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	pool := &WorkerPool{
		config:          workerPoolConfig,
		ctx:             ctx,
		cancel:          cancel,
		publishers:      make(chan TaskFuncWithId, workerPoolConfig.TaskQueueSize),
		workerStopChans: make([]chan bool, workerPoolConfig.MaxWorkers),
		taskId:          atomic.Int64{},
		resultChan:      make(chan SubmitResult, workerPoolConfig.TaskQueueSize),
	}

	if workerPoolConfig.Logger != nil {
		pool.logger = workerPoolConfig.Logger
	} else {
		pool.logger = &defaultLogger{}
	}

	psRunningWorkers := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_pool_running_workers",
		Help: "The total number of running worker",
	})

	psTotalSubmittedTasks := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_pool_total_submitted_tasks",
		Help: "The total number of tasks submitted to the worker pool",
	})

	psTotalExecutedTasks := prometheus.NewCounter(prometheus.CounterOpts{
		Name: "worker_pool_total_executed_tasks",
		Help: "The total number of tasks executed by the worker pool",
	})

	psTasksQueueSize := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "worker_pool_tasks_queue_size",
		Help: "The size of the tasks queue",
	})

	// Register it with the default registry
	prometheus.MustRegister(psRunningWorkers)
	prometheus.MustRegister(psTotalSubmittedTasks)
	prometheus.MustRegister(psTotalExecutedTasks)
	prometheus.MustRegister(psTasksQueueSize)

	pool.stats = make(map[string]prometheus.Metric)
	pool.stats["psTotalSubmittedTasks"] = psTotalSubmittedTasks
	pool.stats["psTotalExecutedTasks"] = psTotalExecutedTasks
	pool.stats["psRunningWorkers"] = psRunningWorkers
	pool.stats["psTasksQueueSize"] = psTasksQueueSize

	for i := 0; i < workerPoolConfig.MaxWorkers; i++ {
		stopChan := make(chan bool)
		pool.workerStopChans[i] = stopChan
		pool.wg.Add(1)
		psRunningWorkers.Inc()
		go pool.worker(i+1, stopChan)
	}

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(":8080", nil)
	return pool
}

// worker is a method on the WorkerPool that processes tasks from the taskQueue.
func (wp *WorkerPool) worker(id int, stopChan chan bool) {
	defer wp.wg.Done()
	// TODO: do I need this??
	// defer wp.stats["psRunningWorkers"].Dec()
	defer wp.logger.Log(fmt.Sprintf("worker %d stopped\n", id))

	for {
		select {
		case <-wp.ctx.Done(): // Check if context was cancelled (pool is stopping)
			return
		case <-time.After(wp.config.Timeout):
			wp.logger.Log(fmt.Sprintf("worker %d timed out\n", id))
			return
		case <-stopChan: // Check if this specific worker was told to stop
			return
		case task, ok := <-wp.publishers: // Wait for a task
			if !ok {
				// The publishers channel was closed, no more tasks will come
				return
			}

			gauge, ok := wp.stats["psTasksQueueSize"].(prometheus.Gauge)
			if ok {
				gauge.Inc()
			}

			if task.Task != nil {
				wp.logger.Log(fmt.Sprintf("worker %d is working on task %d\n", id, task.TaskId))
				result, err := task.Task()

				if err != nil {
					wp.logger.Log(fmt.Sprintf("worker %d error on task %d: %v\n", id, task.TaskId, err))
				}
				counter, ok := wp.stats["psTotalExecutedTasks"].(prometheus.Counter)
				if ok {
					counter.Inc()
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
						wp.logger.Log(fmt.Sprintf("worker %d stuck on sending task %d result, resultChan is full, cannot send result\n", id, task.TaskId))
						// TODO: instead of panic what to do??
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
	counter, ok := wp.stats["psTotalSubmittedTasks"].(prometheus.Counter)
	if ok {
		counter.Inc()
	}
	TaskFuncWithId := TaskFuncWithId{
		Task:   task,
		TaskId: taskId,
	}

loop:
	for {
		select {
		case wp.publishers <- TaskFuncWithId:
			// Task sent successfully
			gauge, ok := wp.stats["psTasksQueueSize"].(prometheus.Gauge)
			if ok {
				gauge.Inc()
			}
			break loop
		default:
			// Channel is full, handle the case when the channel is full
			wp.logger.Log("publishers Channel is full, cannot send task")
		}
	}
	wp.logger.Log(fmt.Sprintf("worker pool submitted task %d\n", taskId))
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
