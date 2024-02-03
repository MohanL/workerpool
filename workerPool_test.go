package workerpool

import (
	"fmt"
	"testing"
	"time"
)

func TestWorkerPoolSubmit(t *testing.T) {

	// f, err := os.Create("trace.out")
	// if err != nil {
	// 	t.Fatalf("failed to create trace output file: %v", err)
	// }
	// defer f.Close()

	// err = trace.Start(f)
	// if err != nil {
	// 	t.Fatalf("failed to start trace: %v", err)
	// }
	// defer trace.Stop()

	maxWorkers := 4
	taskQueueSize := 100
	timeout := 10 * time.Second
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

	time.Sleep(1 * time.Second)
	for i := 0; i < taskNums; i++ {
		workerPool.Submit(task)
	}

	// this is blocking
	for i := 0; i < taskNums; i++ {
		taskResult := <-workerPool.ResultChan
		if taskResult.Error != nil {
			t.Error(taskResult.Error)
		}
		fmt.Printf(
			"task id: %d, result: %s\n", taskResult.TaskId, taskResult.Result)
	}

	fmt.Println("main thread sleep for 1 second")
	time.Sleep(1 * time.Second)
	workerPool.Stop()
	fmt.Println("worker pool stopped")
}
