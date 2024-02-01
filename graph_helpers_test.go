package workerpool

import (
	"fmt"
	"sync"
	"time"

	. "example.com/graphes"
)

func cloneGraph(node *Node, nodeNums int, goRoutines int) *Node {

	if node == nil {
		return nil
	}

	// fixed by queue size

	config := WorkerPoolConfig{
		MaxWorkers:    goRoutines,
		Timeout:       10 * time.Second,
		TaskQueueSize: nodeNums * nodeNums,
	}
	wp := NewWorkerPool(config)
	defer wp.Stop()

	head_clone := node.Clone()
	var visited sync.Map

	visited.Store(node.Val, head_clone)
	var wg sync.WaitGroup //use to synchronize all tasks

	submitTasks(node, head_clone, &visited, &wg, wp)
	wg.Wait()

	processResults(wp, &wg)
	wg.Wait()

	return head_clone
}

func submitTasks(node *Node, clone *Node, visited *sync.Map, wg *sync.WaitGroup, wp *WorkerPool) {
	wg.Add(1)
	wp.Submit(func() (interface{}, error) {
		defer wg.Done()
		// // fmt.Printf("send neightbor clone %d to neighbour %d\n", neighborClone.(*Node).Val, node.Val)

		neighborCloneNeightborsChan, err := dfsCloneWithWorkerPool(node, clone, visited, wg, wp)
		return func() (interface{}, error) {
			defer wg.Done()
			for i := 0; i < len(clone.Neighbors); i++ {
				clone.Neighbors[i] = <-neighborCloneNeightborsChan
			}
			close(neighborCloneNeightborsChan)
			return nil, nil
		}, err
	})
}

func dfsCloneWithWorkerPool(node *Node, cloneNode *Node, visited *sync.Map, wg *sync.WaitGroup, wp *WorkerPool) (chan *Node, error) {
	neighborsChan := make(chan *Node, len(node.Neighbors))
	for _, neighbor := range node.Neighbors {
		localNeighbor := neighbor
		neighborClone, loaded := visited.LoadOrStore(localNeighbor.Val, localNeighbor.Clone()) //cas
		neighborsChan <- neighborClone.(*Node)
		if !loaded {
			submitTasks(localNeighbor, neighborClone.(*Node), visited, wg, wp)
		}
	}
	return neighborsChan, nil
}

// processResults processes the results from the worker pool.
func processResults(wp *WorkerPool, wg *sync.WaitGroup) {
	for {
		select {
		case i := <-wp.resultChan:
			if i.Error != nil {
				panic(i.Error) // Consider a better error handling strategy here.
			}
			if i.Result != nil {
				wg.Add(1)
				wp.Submit(i.Result.(func() (interface{}, error)))
			}
		default:
			return
		}
	}
}

type minEdgeReversalsResult struct {
	NodeId int
	min    int
}

// --------
// Functional helpers
// TODO: export to fp package
func reduce(slice []*TreeNode, initial int, fn func(acc int, value *TreeNode) int) int {
	accumulator := initial
	for _, value := range slice {
		accumulator = fn(accumulator, value)
	}
	return accumulator
}

func mapFunc(slice []*Node, fn func(value *Node) interface{}) []interface{} {
	result := make([]interface{}, len(slice))
	for i, value := range slice {
		result[i] = fn(value)
	}
	return result
}

// Helper function to check if a slice contains a specific value
func contains(slice []int, val int) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}

func minEdgeReversalsWithWorkerPool(n int, edges [][]int) []int {
	nodesMap := InitGraphWithNodesAndEdges(n, edges, false)
	count := make([]int, n)

	var dfs func(node *Node, visited map[int]bool) int

	dfs = func(node *Node, visited map[int]bool) int {
		count := 0
		visited[node.Val] = true
		for i := range node.Neighbors {
			if !visited[node.Neighbors[i].Val] {

				neighborVals := mapFunc(nodesMap[node.Val].Neighbors, func(value *Node) interface{} { return value.Val })
				intNeighborVals := make([]int, len(neighborVals))
				for i, val := range neighborVals {
					intNeighborVals[i] = val.(int)
				}

				if !contains(intNeighborVals, node.Neighbors[i].Val) {
					count += 1
				}
				count += dfs(node.Neighbors[i], visited)
			}
		}

		return count
	}
	config := WorkerPoolConfig{
		MaxWorkers:    1,
		Timeout:       10 * time.Second,
		TaskQueueSize: n * n * n,
		Logger:        &defaultLogger{},
	}
	workerPool := NewWorkerPool(config)

	for i := 0; i < n; i++ {
		nodeId := i
		task := func() (interface{}, error) {
			rootNodeI := GenTreeFromEdges(n, edges, nodeId)
			visited := make(map[int]bool)
			min := dfs(rootNodeI, visited)
			return minEdgeReversalsResult{NodeId: nodeId, min: min}, nil
		}
		workerPool.Submit(task)
	}

	for i := 0; i < n; i++ {
		taskResult := <-workerPool.resultChan
		if taskResult.Error != nil {
			fmt.Errorf("task error")
		}
		count[taskResult.Result.(minEdgeReversalsResult).NodeId] = taskResult.Result.(minEdgeReversalsResult).min
		// fmt.Printf("task id: %d, result: %d\n", taskResult.Result.(minEdgeReversalsResult).NodeId, taskResult.Result.(minEdgeReversalsResult).min)
	}
	return count
}

// V2
// ----------------------
func minEdgeReversalsWithWorkerPoolV2(n int, edges [][]int) []int {

	treeMap := InitGraphWithEdgesInfoV2(n, edges)
	count := make([]int, n)

	var dfs func(node int, visited map[int]bool, treeMap map[int]map[int]int) int

	dfs = func(node int, visited map[int]bool, treeMap map[int]map[int]int) int {
		count := 0
		visited[node] = true
		for k, v := range treeMap[node] {
			if !visited[k] {
				if v != 1 {
					count += 1
				}
				count += dfs(k, visited, treeMap)
			}
		}

		return count
	}

	config := WorkerPoolConfig{
		MaxWorkers:    10,
		Timeout:       10 * time.Second,
		TaskQueueSize: n * n * n,
		Logger:        &defaultLogger{},
	}
	workerPool := NewWorkerPool(config)

	for i := 0; i < n; i++ {
		nodeId := i
		task := func() (interface{}, error) {
			visited := make(map[int]bool)
			min := dfs(nodeId, visited, treeMap)
			return minEdgeReversalsResult{NodeId: nodeId, min: min}, nil
		}
		workerPool.Submit(task)
	}

	for i := 0; i < n; i++ {
		taskResult := <-workerPool.resultChan
		if taskResult.Error != nil {
			fmt.Errorf("task error")
		}
		count[taskResult.Result.(minEdgeReversalsResult).NodeId] = taskResult.Result.(minEdgeReversalsResult).min
	}
	return count
}

func InitGraphWithEdgesInfoV2(nodesNum int, edges [][]int) map[int]map[int]int {
	nodes_map := make(map[int]map[int]int)
	for i := 0; i < nodesNum; i++ {
		nodes_map[i] = make(map[int]int)
	}

	for i := range edges {
		nodes_map[edges[i][0]][edges[i][1]] = 1
		nodes_map[edges[i][1]][edges[i][0]] = 0
	}
	return nodes_map
}
