package main

import (
	"fmt"
	"time"

	. "github.com/MohanL/workerpool"
)

func main() {

	maxWorkers := 4
	taskQueueSize := 200
	timeout := 5 * time.Second
	taskNums := 100

	workerPoolConfig := WorkerPoolConfig{
		MaxWorkers:    maxWorkers,
		Timeout:       timeout,
		TaskQueueSize: taskQueueSize,
	}
	workerPool := NewWorkerPool(workerPoolConfig)
	// TODO: task arguments setup
	task := func() (interface{}, error) {
		fmt.Println("hello")
		return "hello result", nil
	}

	for i := 0; i < taskNums; i++ {
		workerPool.Submit(task)
	}

	workerPool.WaitAll()
	workerPool.Stop()
	fmt.Println("worker pool stopped")
}
