package main

import (
	"fmt"
	"time"

	. "github.com/MohanL/workerpool"
)

func main() {

	maxWorkers := 4
	taskQueueSize := 200
	timeout := 3 * time.Second
	taskNums := 100

	workerPoolConfig := WorkerPoolConfig{
		MaxWorkers:    maxWorkers,
		Timeout:       timeout,
		TaskQueueSize: taskQueueSize,
	}
	workerPool := NewWorkerPool(workerPoolConfig)
	// TODO: task arguments setup
	task := func() (interface{}, error) {
		return "hello", nil
	}

	for i := 0; i < taskNums; i++ {
		workerPool.Submit(task)
	}

	for i := 0; i < taskNums; i++ {
		taskResult := <-workerPool.ResultChan
		fmt.Printf(
			"task id: %d, result: %s\n", taskResult.TaskId, taskResult.Result)
	}
	workerPool.Stop()
	fmt.Println("worker pool stopped")
}
