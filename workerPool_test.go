package workerpool

import (
	"fmt"
	"os"
	"runtime"
	"runtime/trace"
	"slices"
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

// TODO: add test
func TestWorkerPoolMinEdgeReversals(t *testing.T) {

	n := 68
	edges := [][]int{
		{28, 0}, {39, 1}, {1, 28}, {64, 2}, {65, 2}, {48, 2}, {19, 3},
		{66, 4}, {4, 41}, {63, 5}, {6, 39}, {41, 7}, {7, 45}, {8, 46},
		{62, 8}, {9, 20}, {10, 67}, {11, 20}, {11, 49}, {58, 11}, {11, 25},
		{12, 33}, {12, 38}, {50, 13}, {14, 65}, {15, 38}, {15, 44},
		{42, 16}, {17, 36}, {45, 18}, {32, 18}, {19, 27}, {20, 47},
		{21, 57}, {22, 63}, {52, 22}, {23, 52}, {24, 52}, {43, 25},
		{54, 26}, {26, 34}, {27, 48}, {28, 50}, {28, 30}, {66, 29},
		{30, 56}, {44, 30}, {51, 30}, {60, 31}, {32, 35}, {32, 38},
		{34, 38}, {52, 36}, {37, 60}, {39, 62}, {40, 58}, {41, 59},
		{43, 42}, {60, 42}, {42, 44}, {48, 55}, {57, 50}, {51, 61},
		{51, 63}, {52, 55}, {53, 59}, {67, 55},
	}
	out := MinEdgeReversals(n, edges)

	concurrent_result := minEdgeReversalsWithWorkerPool(n, edges)

	if !slices.Equal(out, concurrent_result) {
		t.Errorf("Expected %v but got %v", out, concurrent_result)
	}

	concurrent_result2 := minEdgeReversalsWithWorkerPoolV2(n, edges)

	if !slices.Equal(out, concurrent_result2) {
		t.Errorf("Expected %v but got %v", out, concurrent_result2)
	}
}
