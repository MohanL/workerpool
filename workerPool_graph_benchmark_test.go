package workerpool

import (
	"fmt"
	"runtime"
	"testing"

	. "example.com/graphes"
)

// Benchmark for single-threaded graph clone
func BenchmarkCloneGraphDenseSingleThreaded(b *testing.B) {
	sizes := []int{50, 100, 200, 400, 800, 1600}

	for _, size := range sizes {
		// Run the benchmark for each size.
		b.Run(
			fmt.Sprintf("Size%d", size),
			func(b *testing.B) {
				sampleGraph := InitGraph(GenerateGraph(size, true))
				b.ResetTimer() // Reset the timer to exclude graph generation time

				for i := 0; i < b.N; i++ {
					// Call the single-threaded clone function.
					_ = sampleGraph.CloneGraph()
				}
			},
		)
	}
}

// TODO: add profile explanation on what is blocking
// Benchmark for multi-threaded graph clone using WorkerPool.
func BenchmarkCloneGraphDenseMultiThreaded(b *testing.B) {
	sizes := []int{50, 100, 200, 400, 800, 1600}

	for _, size := range sizes {
		// Run the benchmark for each size.
		b.Run(
			fmt.Sprintf("Size%d", size),
			func(b *testing.B) {
				sampleGraph := InitGraph(GenerateGraph(size, true))
				b.ResetTimer() // Reset the timer to exclude graph generation time

				for i := 0; i < b.N; i++ {
					// Call the single-threaded clone function.
					_ = cloneGraph(sampleGraph, size, runtime.GOMAXPROCS(0))
				}
			},
		)
	}
}

func BenchmarkCloneGraphSparseSingleThreaded(b *testing.B) {
	sizes := []int{50, 100, 200, 400, 800, 1600}

	for _, size := range sizes {
		// Run the benchmark for each size.
		b.Run(
			fmt.Sprintf("Size%d", size),
			func(b *testing.B) {
				sampleGraph := InitGraph(GenerateGraph(size, false))
				b.ResetTimer() // Reset the timer to exclude graph generation time

				for i := 0; i < b.N; i++ {
					// Call the single-threaded clone function.
					_ = sampleGraph.CloneGraph()
				}
			},
		)
	}
}

// TODO: add profile explanation on what is blocking
// Benchmark for multi-threaded graph clone using WorkerPool.
func BenchmarkCloneGraphSparseMultiThreaded(b *testing.B) {
	sizes := []int{50, 100, 200, 400, 800, 1600}

	for _, size := range sizes {
		// Run the benchmark for each size.
		b.Run(
			fmt.Sprintf("Size%d", size),
			func(b *testing.B) {
				sampleGraph := InitGraph(GenerateGraph(size, true))
				b.ResetTimer() // Reset the timer to exclude graph generation time

				for i := 0; i < b.N; i++ {
					// Call the single-threaded clone function.
					_ = cloneGraph(sampleGraph, size, runtime.GOMAXPROCS(0))
				}
			},
		)
	}
}

func BenchmarkCloneGraphDenseMultiThreadedVariableThreadSize(b *testing.B) {
	maxThreads := runtime.GOMAXPROCS(0) * 2
	size := 50
	for threadSize := 1; threadSize < maxThreads; threadSize++ {
		// Run the benchmark for each size.
		b.Run(
			fmt.Sprintf("thread size %d", threadSize),
			func(b *testing.B) {
				sampleGraph := InitGraph(GenerateGraph(size, true))
				b.ResetTimer() // Reset the timer to exclude graph generation time

				for i := 0; i < b.N; i++ {
					// Call the single-threaded clone function.
					_ = cloneGraph(sampleGraph, size, threadSize)
				}
			},
		)
	}
}

func BenchmarkWorkerPoolMinEdgeReversalsSingleThreaded(b *testing.B) {

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
	for i := 0; i < b.N; i++ {
		MinEdgeReversals(n, edges)
	}
}

func BenchmarkWorkerPoolMinEdgeReversalsWorkerPool(b *testing.B) {

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
	for i := 0; i < b.N; i++ {
		minEdgeReversalsWithWorkerPool(n, edges)
	}
}

// how densely connected the graph is—whether it's sparse or dense—as this can also affect performance.
// Create a file to store trace output
// f, err := os.Create("trace.out")
// if err != nil {
// 	b.Fatalf("failed to create trace output file: %v", err)
// }
// defer f.Close()

// // Start the trace
// err = trace.Start(f)
// if err != nil {
// 	b.Fatalf("failed to start trace: %v", err)
// }
// defer trace.Stop()
