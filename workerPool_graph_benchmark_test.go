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
