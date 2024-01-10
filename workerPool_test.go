package workerpool

import (
	"fmt"
	"os"
	"runtime"
	"runtime/trace"
	"testing"
	"time"

	. "example.com/graphes"
)

func TestWorkerPoolSubmit(t *testing.T) {

	f, err := os.Create("trace.out")
	if err != nil {
		t.Fatalf("failed to create trace output file: %v", err)
	}
	defer f.Close()

	err = trace.Start(f)
	if err != nil {
		t.Fatalf("failed to start trace: %v", err)
	}
	defer trace.Stop()

	maxWorkers := 4
	taskQueueSize := 100
	timeout := 10 * time.Second
	taskNums := 100

	workerPool := NewWorkerPool(maxWorkers, taskQueueSize, timeout)
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
		taskResult := <-workerPool.resultChan
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

// https://leetcode.com/problems/clone-graph/description/
func TestWorkerPoolCloneGraph(t *testing.T) {
	numNode := 200
	sampleGraph := InitGraph(GenerateGraph(numNode, true)) // Observation: is the graphes is barely connected without cyclyes, then it is really fast.

	clone := cloneGraph(sampleGraph, numNode, runtime.GOMAXPROCS(0))

	if !IsSameGraph(sampleGraph, clone, map[int]bool{}) {
		t.Error(
			"sampleGraph and cloneGraph are not the same",
		)
	}
}
